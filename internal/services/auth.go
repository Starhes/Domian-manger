package services

import (
	"crypto/rand"
	"domain-manager/internal/config"
	"domain-manager/internal/models"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db           *gorm.DB
	cfg          *config.Config
	emailService *EmailService
}

func NewAuthService(db *gorm.DB, cfg *config.Config) *AuthService {
	return &AuthService{
		db:           db,
		cfg:          cfg,
		emailService: NewEmailServiceWithDB(cfg, db),
	}
}

// 用户注册
func (s *AuthService) Register(req models.RegisterRequest) error {
	return s.RegisterWithContext(nil, req)
}

// RegisterWithContext 用户注册（支持HTTP上下文）
func (s *AuthService) RegisterWithContext(c *gin.Context, req models.RegisterRequest) error {
	// 1. 验证注册请求数据
	if err := req.Validate(); err != nil {
		return err
	}

	// 2. 检查邮箱是否已存在（使用事务确保原子性）
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existingUser models.User
	if err := tx.Where("email = ?", strings.ToLower(strings.TrimSpace(req.Email))).First(&existingUser).Error; err == nil {
		tx.Rollback()
		return errors.New("该邮箱已被注册")
	}

	// 3. 检查是否有重复的待验证记录（防止重复注册）
	var existingVerification models.EmailVerification
	if err := tx.Where("email = ? AND used = ? AND expires_at > ?", 
		strings.ToLower(strings.TrimSpace(req.Email)), false, time.Now()).First(&existingVerification).Error; err == nil {
		tx.Rollback()
		return errors.New("该邮箱已有待验证的注册记录，请检查邮箱或等待过期后重试")
	}

	// 4. 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		tx.Rollback()
		return errors.New("密码加密失败")
	}

	// 5. 创建用户
	user := models.User{
		Email:          strings.ToLower(strings.TrimSpace(req.Email)),
		Password:       string(hashedPassword),
		Nickname:       strings.TrimSpace(req.Nickname),
		IsActive:       false, // 需要邮箱验证
		Status:         "normal",
		DNSRecordQuota: 10, // 默认配额
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		// 检查是否是唯一约束错误（邮箱重复）
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return errors.New("该邮箱已被注册")
		}
		return errors.New("用户创建失败")
	}

	// 6. 生成邮箱验证令牌
	token, err := s.generateRandomToken()
	if err != nil {
		tx.Rollback()
		return errors.New("验证令牌生成失败")
	}

	// 7. 创建验证记录
	verification := models.EmailVerification{
		Email:     user.Email,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24小时有效期
		Used:      false,
	}

	if err := tx.Create(&verification).Error; err != nil {
		tx.Rollback()
		return errors.New("验证记录创建失败")
	}

	// 8. 提交事务
	if err := tx.Commit().Error; err != nil {
		return errors.New("注册事务提交失败")
	}

	// 9. 发送验证邮件（事务外执行，失败不影响注册）
	go func() {
		if err := s.emailService.SendVerificationEmailWithContext(c, user.Email, token); err != nil {
			// 邮件发送失败，可以记录日志或实现重试机制
			// 这里暂时只记录错误，不影响注册流程
			// TODO: 添加日志记录
			_ = err
		}
	}()

	return nil
}

// 用户登录
func (s *AuthService) Login(req models.LoginRequest) (*models.LoginResponse, error) {
	var user models.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return nil, errors.New("邮箱或密码错误")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("邮箱或密码错误")
	}

	// 检查账户状态
	if user.Status == "banned" {
		return nil, errors.New("账户已被封禁，请联系管理员")
	}

	if user.Status == "suspended" {
		return nil, errors.New("账户已被暂停，请联系管理员")
	}

	// 检查账户是否激活（管理员账号无需验证）
	if !user.IsActive && !user.IsAdmin {
		return nil, errors.New("账户未激活，请检查邮箱验证链接")
	}

	// 更新登录信息
	now := time.Now()
	user.LastLoginAt = &now
	user.LoginCount++
	s.db.Model(&user).Updates(map[string]interface{}{
		"last_login_at": now,
		"login_count":   user.LoginCount,
	})

	// 生成JWT令牌
	token, err := s.generateJWT(user.ID)
	if err != nil {
		return nil, errors.New("令牌生成失败")
	}

	return &models.LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

// 邮箱验证
func (s *AuthService) VerifyEmail(token string) error {
	var verification models.EmailVerification
	if err := s.db.Where("token = ? AND used = false AND expires_at > ?",
		token, time.Now()).First(&verification).Error; err != nil {
		return errors.New("验证令牌无效或已过期")
	}

	// 标记令牌为已使用
	verification.Used = true
	if err := s.db.Save(&verification).Error; err != nil {
		return errors.New("验证状态更新失败")
	}

	// 激活用户账户
	if err := s.db.Model(&models.User{}).Where("email = ?", verification.Email).
		Update("is_active", true).Error; err != nil {
		return errors.New("账户激活失败")
	}

	return nil
}

// 忘记密码
func (s *AuthService) ForgotPassword(req models.ForgotPasswordRequest) error {
	// 检查用户是否存在
	var user models.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		// 为了安全，即使用户不存在也返回成功
		return nil
	}

	// 生成重置令牌
	token, err := s.generateRandomToken()
	if err != nil {
		return errors.New("重置令牌生成失败")
	}

	reset := models.PasswordReset{
		Email:     req.Email,
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour), // 1小时有效期
	}

	if err := s.db.Create(&reset).Error; err != nil {
		return errors.New("重置记录创建失败")
	}

	// 发送重置邮件
	if err := s.emailService.SendPasswordResetEmail(req.Email, token); err != nil {
		// 邮件发送失败不影响重置流程，只记录错误
		_ = err // 忽略邮件发送错误
	}

	return nil
}

// 重置密码
func (s *AuthService) ResetPassword(req models.ResetPasswordRequest) error {
	var reset models.PasswordReset
	if err := s.db.Where("token = ? AND used = false AND expires_at > ?",
		req.Token, time.Now()).First(&reset).Error; err != nil {
		return errors.New("重置令牌无效或已过期")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}

	// 更新用户密码
	if err := s.db.Model(&models.User{}).Where("email = ?", reset.Email).
		Update("password", string(hashedPassword)).Error; err != nil {
		return errors.New("密码更新失败")
	}

	// 标记重置令牌为已使用
	reset.Used = true
	if err := s.db.Save(&reset).Error; err != nil {
		return errors.New("重置状态更新失败")
	}

	return nil
}

// 生成JWT令牌
func (s *AuthService) generateJWT(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7天有效期
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

// UpdateProfile 更新用户资料
func (s *AuthService) UpdateProfile(userID uint, email, password string) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	// 检查是否有需要更新的字段
	updates := make(map[string]interface{})

	// 更新邮箱
	if email != "" && email != user.Email {
		// 检查新邮箱是否已被使用
		var existingUser models.User
		if err := s.db.Where("email = ? AND id != ?", email, userID).First(&existingUser).Error; err == nil {
			return nil, errors.New("该邮箱已被其他用户使用")
		}

		updates["email"] = email
		// 如果更换邮箱，需要重新验证
		updates["is_active"] = false

		// TODO: 发送新邮箱验证邮件
		// 这里可以生成新的验证token并发送验证邮件
	}

	// 更新密码
	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.New("密码加密失败")
		}
		updates["password"] = string(hashedPassword)
	}

	// 如果没有需要更新的字段
	if len(updates) == 0 {
		return &user, nil
	}

	// 执行更新
	if err := s.db.Model(&user).Updates(updates).Error; err != nil {
		return nil, errors.New("用户资料更新失败")
	}

	// 重新查询用户信息并返回
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, errors.New("获取更新后的用户信息失败")
	}

	return &user, nil
}

// 生成随机令牌
func (s *AuthService) generateRandomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
