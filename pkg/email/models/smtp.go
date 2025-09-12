package models

import (
	"time"

	"gorm.io/gorm"
)

// SMTPConfig SMTP配置模型
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

// CreateSMTPConfigRequest SMTP配置创建请求
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

// UpdateSMTPConfigRequest SMTP配置更新请求
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

// TestSMTPConfigRequest SMTP配置测试请求
type TestSMTPConfigRequest struct {
	ToEmail string `json:"to_email" binding:"required,email"`
}

// SMTPConfigResponse SMTP配置响应（脱敏版本）
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