package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ValidatePort 验证端口号
func ValidatePort(port string) error {
	if port == "" {
		return errors.New("端口不能为空")
	}
	
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return errors.New("端口必须是数字")
	}
	
	if portNum < 1 || portNum > 65535 {
		return errors.New("端口必须在1-65535之间")
	}
	
	return nil
}

// ValidateConfigValue 验证配置值
func ValidateConfigValue(key, value string, isProduction bool) error {
	if value == "" {
		return fmt.Errorf("%s 不能为空", key)
	}
	
	switch key {
	case "JWT_SECRET":
		return validateJWTSecret(value, isProduction)
	case "ENCRYPTION_KEY":
		return validateEncryptionKey(value, isProduction)
	case "DB_PASSWORD":
		return validatePassword(value, isProduction)
	case "SMTP_PASSWORD":
		return validatePassword(value, isProduction)
	default:
		return nil
	}
}

// validateJWTSecret 验证JWT密钥
func validateJWTSecret(secret string, isProduction bool) error {
	if len(secret) < 32 {
		return errors.New("JWT密钥长度至少32位")
	}
	
	if isProduction && len(secret) < 64 {
		return errors.New("生产环境JWT密钥长度至少64位")
	}
	
	// 检查复杂度
	if isProduction {
		hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(secret)
		hasLower := regexp.MustCompile(`[a-z]`).MatchString(secret)
		hasNumber := regexp.MustCompile(`[0-9]`).MatchString(secret)
		hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~]`).MatchString(secret)
		
		complexity := 0
		if hasUpper { complexity++ }
		if hasLower { complexity++ }
		if hasNumber { complexity++ }
		if hasSpecial { complexity++ }
		
		if complexity < 3 {
			return errors.New("生产环境JWT密钥必须包含大写字母、小写字母、数字和特殊字符中的至少3种")
		}
	}
	
	return nil
}

// validateEncryptionKey 验证加密密钥
func validateEncryptionKey(key string, isProduction bool) error {
	// AES密钥必须是32字节的十六进制字符串
	if len(key) != 64 {
		return errors.New("加密密钥必须是64位十六进制字符串（32字节）")
	}
	
	// 验证是否为有效的十六进制
	matched, _ := regexp.MatchString(`^[0-9a-fA-F]{64}$`, key)
	if !matched {
		return errors.New("加密密钥必须是有效的十六进制字符串")
	}
	
	return nil
}

// validatePassword 验证密码强度
func validatePassword(password string, isProduction bool) error {
	if len(password) < 8 {
		return errors.New("密码长度至少8位")
	}
	
	if isProduction && len(password) < 12 {
		return errors.New("生产环境密码长度至少12位")
	}
	
	// 检查字符类型
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~]`).MatchString(password)
	
	if !hasLetter {
		return errors.New("密码必须包含字母")
	}
	
	if !hasNumber {
		return errors.New("密码必须包含数字")
	}
	
	if isProduction && !hasSpecial {
		return errors.New("生产环境密码必须包含特殊字符")
	}
	
	// 检查常见弱密码
	weakPasswords := []string{
		"password", "123456", "admin", "root", "user",
		"test", "demo", "guest", "default", "changeme",
	}
	
	lowerPassword := strings.ToLower(password)
	for _, weak := range weakPasswords {
		if strings.Contains(lowerPassword, weak) {
			return fmt.Errorf("密码不能包含常见弱密码词汇: %s", weak)
		}
	}
	
	return nil
}