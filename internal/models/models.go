package models

import (
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

// 请求和响应结构体

// 用户注册请求
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
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
