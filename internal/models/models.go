package models

import (
	"time"

	"gorm.io/gorm"
)

// 用户模型
type User struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Email       string         `json:"email" gorm:"uniqueIndex;not null"`
	Password    string         `json:"-" gorm:"not null"` // 不返回密码字段
	IsActive    bool           `json:"is_active" gorm:"default:false"`
	IsAdmin     bool           `json:"is_admin" gorm:"default:false"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	
	// 关联
	DNSRecords []DNSRecord `json:"dns_records,omitempty" gorm:"foreignKey:UserID"`
}

// 域名模型
type Domain struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"uniqueIndex;not null"` // 主域名，如 example.com
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	// 关联
	DNSRecords []DNSRecord `json:"dns_records,omitempty" gorm:"foreignKey:DomainID"`
	DomainType string
	UserID     uint
	User       User
}

// DNS记录模型
type DNSRecord struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	UserID     uint           `json:"user_id" gorm:"not null"`
	DomainID   uint           `json:"domain_id" gorm:"not null"`
	Subdomain  string         `json:"subdomain" gorm:"not null"` // 子域名，如 www
	Type       string         `json:"type" gorm:"not null"`      // A, CNAME, TXT等
	Value      string         `json:"value" gorm:"not null"`     // 记录值
	TTL        int            `json:"ttl" gorm:"default:600"`    // 生存时间
	ExternalID string         `json:"external_id"`               // DNS服务商返回的记录ID
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
	
	// 关联
	User   User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Domain Domain `json:"domain,omitempty" gorm:"foreignKey:DomainID"`
}

// DNS服务商模型
type DNSProvider struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"not null"`        // 服务商名称，如 DNSPod
	Type      string         `json:"type" gorm:"not null"`        // 服务商类型，如 dnspod
	Config    string         `json:"config" gorm:"type:text"`     // JSON格式的配置信息
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
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
