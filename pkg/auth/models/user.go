package models

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	Email          string         `json:"email" gorm:"uniqueIndex;not null;size:255"`
	Password       string         `json:"-" gorm:"not null;size:255"` // bcrypt哈希后的密码
	Nickname       string         `json:"nickname" gorm:"size:100"`   // 用户昵称
	Avatar         string         `json:"avatar" gorm:"size:500"`     // 头像URL
	IsActive       bool           `json:"is_active" gorm:"default:false;index"`
	IsAdmin        bool           `json:"is_admin" gorm:"default:false;index"`
	LastLoginAt    *time.Time     `json:"last_login_at"`                        // 最后登录时间
	LoginCount     int            `json:"login_count" gorm:"default:0"`         // 登录次数
	DNSRecordQuota int            `json:"dns_record_quota" gorm:"default:10"`   // DNS记录配额
	Status         string         `json:"status" gorm:"default:normal;size:20"` // 用户状态：normal, suspended, banned
	CreatedAt      time.Time      `json:"created_at" gorm:"index"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

// EmailVerification 邮箱验证模型
type EmailVerification struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"not null"`
	Token     string    `json:"token" gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	Used      bool      `json:"used" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
}

// PasswordReset 密码重置模型
type PasswordReset struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"not null"`
	Token     string    `json:"token" gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	Used      bool      `json:"used" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
}

// 请求和响应结构体

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Email           string `json:"email" binding:"required,email,max=255"`
	Password        string `json:"password" binding:"required,min=8,max=100"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
	Nickname        string `json:"nickname" binding:"max=100"`
}

// LoginRequest 用户登录请求
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	User         User   `json:"user"`
}

// ForgotPasswordRequest 忘记密码请求
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// ========================= 验证方法 =========================

// ValidatePassword 验证密码强度
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("密码长度至少8位")
	}
	
	if len(password) > 100 {
		return errors.New("密码长度不能超过100位")
	}
	
	// 检查是否包含字母
	hasLetter, _ := regexp.MatchString(`[a-zA-Z]`, password)
	if !hasLetter {
		return errors.New("密码必须包含字母")
	}
	
	// 检查是否包含数字
	hasNumber, _ := regexp.MatchString(`[0-9]`, password)
	if !hasNumber {
		return errors.New("密码必须包含数字")
	}
	
	// 检查是否包含特殊字符（可选，但推荐）
	hasSpecial, _ := regexp.MatchString(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~]`, password)
	
	// 如果密码长度小于12，则必须包含特殊字符
	if len(password) < 12 && !hasSpecial {
		return errors.New("密码长度小于12位时必须包含特殊字符")
	}
	
	return nil
}

// ValidateEmail 验证邮箱格式
func ValidateEmail(email string) error {
	if len(email) == 0 {
		return errors.New("邮箱不能为空")
	}
	
	if len(email) > 255 {
		return errors.New("邮箱长度不能超过255位")
	}
	
	// 基本的邮箱格式检查
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("邮箱格式不正确")
	}
	
	// 检查是否包含常见的不安全字符
	unsafeChars := []string{"<", ">", "\"", "'", "&", ";", "|", "`", "$"}
	for _, char := range unsafeChars {
		if strings.Contains(email, char) {
			return errors.New("邮箱包含不安全字符")
		}
	}
	
	return nil
}

// ValidateNickname 验证用户昵称
func ValidateNickname(nickname string) error {
	if len(nickname) > 100 {
		return errors.New("昵称长度不能超过100位")
	}
	
	if len(nickname) > 0 {
		// 检查是否包含不安全字符
		unsafeChars := []string{"<", ">", "\"", "'", "&", ";", "|", "`", "$", "\\"}
		for _, char := range unsafeChars {
			if strings.Contains(nickname, char) {
				return errors.New("昵称包含不安全字符")
			}
		}
		
		// 去除首尾空格
		nickname = strings.TrimSpace(nickname)
		if len(nickname) == 0 {
			return errors.New("昵称不能只包含空格")
		}
	}
	
	return nil
}

// Validate 验证注册请求
func (req *RegisterRequest) Validate() error {
	// 验证邮箱
	if err := ValidateEmail(req.Email); err != nil {
		return err
	}
	
	// 验证密码
	if err := ValidatePassword(req.Password); err != nil {
		return err
	}
	
	// 验证确认密码
	if req.Password != req.ConfirmPassword {
		return errors.New("两次输入的密码不一致")
	}
	
	// 验证昵称
	if err := ValidateNickname(req.Nickname); err != nil {
		return err
	}
	
	return nil
}