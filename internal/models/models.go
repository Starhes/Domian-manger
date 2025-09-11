package models

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
)

// 用户模型
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

	// 关联
	DNSRecords []DNSRecord `json:"dns_records,omitempty" gorm:"foreignKey:UserID"`
}

// 域名模型
type Domain struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"uniqueIndex;not null"` // 主域名，如 example.com
	DomainType  string         `json:"domain_type"`                      // 域名类型：二级域名、三级域名等
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	Description string         `json:"description"` // 域名描述
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	DNSRecords []DNSRecord `json:"dns_records,omitempty" gorm:"foreignKey:DomainID"`
}

// DNS记录模型
type DNSRecord struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	UserID     uint           `json:"user_id" gorm:"not null;index"`                           // 添加索引提升查询性能
	DomainID   uint           `json:"domain_id" gorm:"not null;index"`                         // 添加索引
	Subdomain  string         `json:"subdomain" gorm:"not null;size:63"`                       // 子域名，限制长度
	Type       string         `json:"type" gorm:"not null;size:10;index"`                      // 记录类型，添加索引
	Value      string         `json:"value" gorm:"not null;size:500"`                          // 记录值，限制长度
	TTL        int            `json:"ttl" gorm:"default:600;check:ttl >= 1 AND ttl <= 604800"` // TTL范围检查
	Priority   int            `json:"priority" gorm:"default:0"`                               // MX记录优先级
	ExternalID string         `json:"external_id" gorm:"size:100"`                             // DNS服务商记录ID
	Status     string         `json:"status" gorm:"default:active;size:20"`                    // 记录状态
	Comment    string         `json:"comment" gorm:"size:255"`                                 // 记录备注
	CreatedAt  time.Time      `json:"created_at" gorm:"index"`                                 // 添加时间索引
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	User   User   `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Domain Domain `json:"domain,omitempty" gorm:"foreignKey:DomainID;constraint:OnDelete:CASCADE"`
}

// DNS服务商模型
type DNSProvider struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null;size:100;uniqueIndex"` // 服务商名称唯一
	Type        string         `json:"type" gorm:"not null;size:50;index"`        // 服务商类型索引
	Config      string         `json:"config" gorm:"type:text"`                   // JSON格式配置
	IsActive    bool           `json:"is_active" gorm:"default:false;index"`      // 默认禁用，添加索引
	Description string         `json:"description" gorm:"size:500"`               // 服务商描述
	SortOrder   int            `json:"sort_order" gorm:"default:0"`               // 排序字段
	LastTestAt  *time.Time     `json:"last_test_at"`                              // 最后测试时间
	TestResult  string         `json:"test_result" gorm:"size:1000"`              // 测试结果
	CreatedAt   time.Time      `json:"created_at" gorm:"index"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// 邮箱验证模型
type EmailVerification struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"not null"`
	Token     string    `json:"token" gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	Used      bool      `json:"used" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
}

// 密码重置模型
type PasswordReset struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"not null"`
	Token     string    `json:"token" gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	Used      bool      `json:"used" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
}

// SMTP配置模型
type SMTPConfig struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null;size:100"`                    // 配置名称
	Host        string         `json:"host" gorm:"not null;size:255"`                    // SMTP服务器地址
	Port        int            `json:"port" gorm:"not null;default:587"`                 // SMTP端口
	Username    string         `json:"username" gorm:"not null;size:255"`                // 用户名
	Password    string         `json:"-" gorm:"not null;size:255"`                       // 密码，不返回给前端
	FromEmail   string         `json:"from_email" gorm:"not null;size:255"`              // 发件人邮箱
	FromName    string         `json:"from_name" gorm:"size:100"`                        // 发件人名称
	IsActive    bool           `json:"is_active" gorm:"default:false;index"`             // 是否启用
	IsDefault   bool           `json:"is_default" gorm:"default:false"`                  // 是否为默认配置
	UseTLS      bool           `json:"use_tls" gorm:"default:true"`                      // 是否使用TLS
	Description string         `json:"description" gorm:"size:500"`                      // 配置描述
	LastTestAt  *time.Time     `json:"last_test_at"`                                     // 最后测试时间
	TestResult  string         `json:"test_result" gorm:"size:1000"`                     // 测试结果
	CreatedAt   time.Time      `json:"created_at" gorm:"index"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// 请求和响应结构体

// 用户注册请求
type RegisterRequest struct {
	Email           string `json:"email" binding:"required,email,max=255"`
	Password        string `json:"password" binding:"required,min=8,max=100"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
	Nickname        string `json:"nickname" binding:"max=100"`
}

// 用户登录请求
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// 登录响应
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// DNS记录创建请求
type CreateDNSRecordRequest struct {
	DomainID  uint   `json:"domain_id" binding:"required"`
	Subdomain string `json:"subdomain" binding:"required"`
	Type      string `json:"type" binding:"required,oneof=A CNAME TXT MX"`
	Value     string `json:"value" binding:"required"`
	TTL       int    `json:"ttl"`
}

// DNS记录更新请求
type UpdateDNSRecordRequest struct {
	Subdomain string `json:"subdomain"`
	Type      string `json:"type" binding:"oneof=A CNAME TXT MX"`
	Value     string `json:"value"`
	TTL       int    `json:"ttl"`
}

// 忘记密码请求
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// 重置密码请求
type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// SMTP配置创建请求
type CreateSMTPConfigRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	Host        string `json:"host" binding:"required,max=255"`
	Port        int    `json:"port" binding:"required,min=1,max=65535"`
	Username    string `json:"username" binding:"required,max=255"`
	Password    string `json:"password" binding:"required,max=255"`
	FromEmail   string `json:"from_email" binding:"required,email,max=255"`
	FromName    string `json:"from_name" binding:"max=100"`
	UseTLS      bool   `json:"use_tls"`
	Description string `json:"description" binding:"max=500"`
}

// SMTP配置更新请求
type UpdateSMTPConfigRequest struct {
	Name        string `json:"name" binding:"max=100"`
	Host        string `json:"host" binding:"max=255"`
	Port        int    `json:"port" binding:"min=1,max=65535"`
	Username    string `json:"username" binding:"max=255"`
	Password    string `json:"password" binding:"max=255"` // 可选，为空则不更新密码
	FromEmail   string `json:"from_email" binding:"email,max=255"`
	FromName    string `json:"from_name" binding:"max=100"`
	UseTLS      *bool  `json:"use_tls"` // 使用指针以区分false和未设置
	Description string `json:"description" binding:"max=500"`
}

// SMTP配置测试请求
type TestSMTPConfigRequest struct {
	ToEmail string `json:"to_email" binding:"required,email"`
}

// SMTP配置响应（脱敏版本）
type SMTPConfigResponse struct {
	ID          uint       `json:"id"`
	Name        string     `json:"name"`
	Host        string     `json:"host"`
	Port        int        `json:"port"`
	Username    string     `json:"username"`
	FromEmail   string     `json:"from_email"`
	FromName    string     `json:"from_name"`
	IsActive    bool       `json:"is_active"`
	IsDefault   bool       `json:"is_default"`
	UseTLS      bool       `json:"use_tls"`
	Description string     `json:"description"`
	LastTestAt  *time.Time `json:"last_test_at"`
	TestResult  string     `json:"test_result"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
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
