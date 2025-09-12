package config

import (
	"domain-max/pkg/utils"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

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

	// 加密配置
	EncryptionKey string // 用于加密敏感数据如SMTP密码

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

		EncryptionKey: getEnv("ENCRYPTION_KEY", ""), // 用于加密敏感数据

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
	isProduction := c.Environment == "production"
	
	// 验证端口
	if err := utils.ValidatePort(c.Port); err != nil {
		return fmt.Errorf("端口配置错误: %v", err)
	}
	
	// 验证必要的配置项
	requiredConfigs := map[string]string{
		"DB_PASSWORD":    c.DBPassword,
		"JWT_SECRET":     c.JWTSecret,
		"ENCRYPTION_KEY": c.EncryptionKey,
	}
	
	for key, value := range requiredConfigs {
		if value == "" {
			return fmt.Errorf("%s 不能为空，请设置相应的环境变量", key)
		}
		
		// 使用统一的配置验证
		if err := utils.ValidateConfigValue(key, value, isProduction); err != nil {
			return fmt.Errorf("%s 配置错误: %v", key, err)
		}
	}
	
	// 验证可选的SMTP配置
	if c.SMTPPassword != "" {
		if err := utils.ValidateConfigValue("SMTP_PASSWORD", c.SMTPPassword, isProduction); err != nil {
			return fmt.Errorf("SMTP_PASSWORD 配置错误: %v", err)
		}
	}
	
	// 生产环境额外安全检查
	if isProduction {
		if err := c.validateProductionSecurity(); err != nil {
			return err
		}
	}
	
	// 验证数据库类型
	validDBTypes := []string{"postgres", "mysql"}
	if !contains(validDBTypes, c.DBType) {
		return fmt.Errorf("不支持的数据库类型: %s，支持的类型: %s", c.DBType, strings.Join(validDBTypes, ", "))
	}
	
	return nil
}

// validateProductionSecurity 生产环境安全验证
func (c *Config) validateProductionSecurity() error {
	// 检查是否使用了明显的测试/默认值
	dangerousValues := []string{
		"test", "demo", "example", "sample", "default",
		"localhost", "127.0.0.1", "your_", "change_this",
		"password", "secret", "key", "token",
	}
	
	securityConfigs := map[string]string{
		"数据库密码":   c.DBPassword,
		"JWT密钥":   c.JWTSecret,
		"加密密钥":    c.EncryptionKey,
	}
	
	for configName, configValue := range securityConfigs {
		lowerValue := strings.ToLower(configValue)
		for _, dangerous := range dangerousValues {
			if strings.Contains(lowerValue, dangerous) {
				return fmt.Errorf("生产环境的%s包含不安全的词汇 '%s'，请使用随机生成的密钥", configName, dangerous)
			}
		}
	}
	
	// 检查BaseURL是否配置为生产域名
	if c.BaseURL != "" && strings.Contains(c.BaseURL, "localhost") {
		return errors.New("生产环境不能使用localhost作为BaseURL，请配置正确的域名")
	}
	
	return nil
}

// contains 检查切片是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}