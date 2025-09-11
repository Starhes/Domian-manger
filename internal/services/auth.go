package services

import (
	"crypto/rand"
	"domain-manager/internal/config"
	"domain-manager/internal/models"
	"encoding/hex"
	"errors"
	"time"

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
		emailService: NewEmailService(cfg),
	}
}

// 用户注册
func (s *AuthService) Register(req models.RegisterRequest) error {
	// 检查邮箱是否已存在
	var existingUser models.User
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return errors.New("邮箱已被注册")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}

	// 创建用户
	user := models.User{
		Email:          req.Email,
		Password:       string(hashedPassword),
		IsActive:       false, // 需要邮箱验证
		Status:         "normal",
		DNSRecordQuota: 10, // 默认配额
	}

	// 验证用户数据
	if err := user.ValidateUser(); err != nil {
		return err
	}

	if err := s.db.Create(&user).Error; err != nil {
		return errors.New("用户创建失败")
	}

	// 生成邮箱验证令牌
	token, err := s.generateRandomToken()
	if err != nil {
		return errors.New("验证令牌生成失败")
	}

	verification := models.EmailVerification{
		Email:     req.Email,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24小时有效期
	}

	if err := s.db.Create(&verification).Error; err != nil {
		return errors.New("验证记录创建失败")
	}

	// 发送验证邮件
	if err := s.emailService.SendVerificationEmail(req.Email, token); err != nil {
		// 邮件发送失败不影响注册流程，只记录错误
		// 可以考虑使用日志系统记录
		// 这里暂时不返回错误，让用户可以继续使用系统
		_ = err // 忽略邮件发送错误
	}

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
