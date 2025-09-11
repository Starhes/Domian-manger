package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// 服务器配置
	Port        string
	Environment string
	BaseURL     string // 系统基础URL，用于生成邮件链接

	// 数据库配置
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBType     string // postgres 或 mysql

	// JWT配置
	JWTSecret string

	// 邮件配置
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string

	// DNSPod配置
	DNSPodToken string
}

func Load() *Config {
	// 尝试加载.env文件
	godotenv.Load()

	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		BaseURL:     getEnv("BASE_URL", ""), // 为空时将自动检测

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""), // 移除默认密码，强制用户设置
		DBName:     getEnv("DB_NAME", "domain_manager"),
		DBType:     getEnv("DB_TYPE", "postgres"),

		JWTSecret: getEnv("JWT_SECRET", ""), // 移除默认JWT密钥，强制用户设置

		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnvInt("SMTP_PORT", 587),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", "noreply@example.com"),

		DNSPodToken: getEnv("DNSPOD_TOKEN", ""),
	}

	// 如果没有设置BASE_URL，根据环境和端口自动生成
	if cfg.BaseURL == "" {
		if cfg.Environment == "development" {
			cfg.BaseURL = fmt.Sprintf("http://localhost:%s", cfg.Port)
		} else {
			// 生产环境默认使用HTTPS，域名需要用户配置
			cfg.BaseURL = fmt.Sprintf("https://localhost:%s", cfg.Port)
		}
	}

	// 验证必要的配置项
	if err := cfg.validate(); err != nil {
		panic("配置验证失败: " + err.Error())
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// validate 验证配置项的有效性
func (c *Config) validate() error {
	// 验证数据库密码
	if c.DBPassword == "" {
		return errors.New("数据库密码不能为空，请设置 DB_PASSWORD 环境变量")
	}

	// 验证JWT密钥
	if c.JWTSecret == "" {
		return errors.New("JWT密钥不能为空，请设置 JWT_SECRET 环境变量")
	}

	// JWT密钥长度检查
	if len(c.JWTSecret) < 32 {
		return errors.New("JWT密钥长度至少需要32个字符以确保安全性")
	}

	// 生产环境额外检查
	if c.Environment == "production" {
		// 检查是否使用了不安全的默认值
		unsafeValues := []string{
			"password", "admin123", "123456", "your_jwt_secret_key_change_this_in_production",
			"your_secure_password_here", "change-this-in-production",
		}

		for _, unsafe := range unsafeValues {
			if c.DBPassword == unsafe {
				return errors.New("生产环境不能使用默认或弱密码，请设置更强的 DB_PASSWORD")
			}
			if c.JWTSecret == unsafe {
				return errors.New("生产环境不能使用默认的JWT密钥，请设置更强的 JWT_SECRET")
			}
		}

		// 生产环境密码强度检查
		if len(c.DBPassword) < 8 {
			return errors.New("生产环境数据库密码长度至少需要8个字符")
		}
	}

	return nil
}
