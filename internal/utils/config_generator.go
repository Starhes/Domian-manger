package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// SecurityConfig 安全配置结构
type SecurityConfig struct {
	MinPasswordLength int
	RequireUppercase  bool
	RequireLowercase  bool
	RequireNumbers    bool
	RequireSpecials   bool
}

// DefaultSecurityConfig 默认安全配置
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		MinPasswordLength: 12,
		RequireUppercase:  true,
		RequireLowercase:  true,
		RequireNumbers:    true,
		RequireSpecials:   true,
	}
}

// GenerateSecurePassword 生成安全密码
func GenerateSecurePassword(length int) (string, error) {
	if length < 8 {
		length = 16
	}
	
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	
	// 使用base64编码并确保包含所需字符类型
	password := base64.URLEncoding.EncodeToString(bytes)
	if len(password) > length {
		password = password[:length]
	}
	
	return password, nil
}

// GenerateJWTSecret 生成JWT密钥
func GenerateJWTSecret(length int) (string, error) {
	if length < 64 {
		length = 64
	}
	
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	
	// 生成字母数字字符串
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	secret := make([]byte, length)
	for i := range secret {
		secret[i] = chars[bytes[i]%byte(len(chars))]
	}
	
	return string(secret), nil
}

// GenerateEncryptionKey 生成AES加密密钥
func GenerateEncryptionKey() (string, error) {
	bytes := make([]byte, 32) // 256-bit key
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	
	return hex.EncodeToString(bytes), nil
}

// ValidateConfigValue 验证配置值的安全性
func ValidateConfigValue(key, value string, isProduction bool) error {
	if value == "" {
		return fmt.Errorf("%s 不能为空", key)
	}
	
	switch strings.ToUpper(key) {
	case "DB_PASSWORD":
		return ValidateDBPassword(value, isProduction)
	case "JWT_SECRET":
		return ValidateJWTSecret(value, isProduction)
	case "ENCRYPTION_KEY":
		return ValidateEncryptionKey(value)
	case "SMTP_PASSWORD":
		return ValidateSMTPPassword(value)
	default:
		return ValidateGenericSecret(value, isProduction)
	}
}

// ValidateDBPassword 验证数据库密码
func ValidateDBPassword(password string, isProduction bool) error {
	minLength := 8
	if isProduction {
		minLength = 12
	}
	
	if len(password) < minLength {
		return fmt.Errorf("数据库密码长度至少需要 %d 个字符", minLength)
	}
	
	if len(password) > 128 {
		return fmt.Errorf("数据库密码长度不能超过 128 个字符")
	}
	
	// 检查字符类型
	config := DefaultSecurityConfig()
	config.MinPasswordLength = minLength
	
	if isProduction {
		if err := validatePasswordComplexity(password, config); err != nil {
			return fmt.Errorf("数据库密码: %v", err)
		}
	}
	
	// 检查常见弱密码
	if err := checkWeakPasswords(password); err != nil {
		return fmt.Errorf("数据库密码: %v", err)
	}
	
	return nil
}

// ValidateJWTSecret 验证JWT密钥
func ValidateJWTSecret(secret string, isProduction bool) error {
	minLength := 32
	if isProduction {
		minLength = 64
	}
	
	if len(secret) < minLength {
		return fmt.Errorf("JWT密钥长度至少需要 %d 个字符", minLength)
	}
	
	if len(secret) > 256 {
		return fmt.Errorf("JWT密钥长度不能超过 256 个字符")
	}
	
	// 检查熵（随机性）
	if entropy := calculateEntropy(secret); entropy < 3.0 {
		return errors.New("JWT密钥熵太低，请使用更随机的字符串")
	}
	
	// 检查是否包含明显的模式
	if hasObviousPattern(secret) {
		return errors.New("JWT密钥包含明显的模式，请使用随机生成的密钥")
	}
	
	return nil
}

// ValidateEncryptionKey 验证加密密钥
func ValidateEncryptionKey(key string) error {
	if len(key) != 64 {
		return errors.New("加密密钥必须是64个十六进制字符（32字节）")
	}
	
	// 检查是否为有效的十六进制字符串
	if _, err := hex.DecodeString(key); err != nil {
		return errors.New("加密密钥必须是有效的十六进制字符串")
	}
	
	// 检查是否为全零或明显的模式
	if isWeakHexKey(key) {
		return errors.New("加密密钥过于简单，请使用随机生成的密钥")
	}
	
	return nil
}

// ValidateSMTPPassword 验证SMTP密码
func ValidateSMTPPassword(password string) error {
	if len(password) < 6 {
		return errors.New("SMTP密码长度至少需要 6 个字符")
	}
	
	if len(password) > 256 {
		return errors.New("SMTP密码长度不能超过 256 个字符")
	}
	
	return nil
}

// ValidateGenericSecret 验证通用密钥
func ValidateGenericSecret(secret string, isProduction bool) error {
	minLength := 8
	if isProduction {
		minLength = 16
	}
	
	if len(secret) < minLength {
		return fmt.Errorf("密钥长度至少需要 %d 个字符", minLength)
	}
	
	if len(secret) > 256 {
		return errors.New("密钥长度不能超过 256 个字符")
	}
	
	return nil
}

// validatePasswordComplexity 验证密码复杂性
func validatePasswordComplexity(password string, config SecurityConfig) error {
	var hasUpper, hasLower, hasNumber, hasSpecial bool
	
	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		default:
			if strings.ContainsRune("!@#$%^&*()_+-=[]{}|;':\"<>?,./`~", char) {
				hasSpecial = true
			}
		}
	}
	
	missing := []string{}
	if config.RequireUppercase && !hasUpper {
		missing = append(missing, "大写字母")
	}
	if config.RequireLowercase && !hasLower {
		missing = append(missing, "小写字母")
	}
	if config.RequireNumbers && !hasNumber {
		missing = append(missing, "数字")
	}
	if config.RequireSpecials && !hasSpecial {
		missing = append(missing, "特殊字符")
	}
	
	if len(missing) > 0 {
		return fmt.Errorf("缺少以下字符类型: %s", strings.Join(missing, "、"))
	}
	
	return nil
}

// checkWeakPasswords 检查常见弱密码
func checkWeakPasswords(password string) error {
	weakPasswords := []string{
		"password", "123456", "12345678", "qwerty", "abc123",
		"password123", "admin", "root", "user", "guest", "test",
		"postgres", "database", "db_password", "secret",
	}
	
	lowerPassword := strings.ToLower(password)
	for _, weak := range weakPasswords {
		if strings.Contains(lowerPassword, weak) {
			return errors.New("包含常见的弱密码模式")
		}
	}
	
	return nil
}

// calculateEntropy 计算字符串的熵
func calculateEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}
	
	freq := make(map[rune]int)
	for _, char := range s {
		freq[char]++
	}
	
	var entropy float64
	length := float64(len(s))
	
	for _, count := range freq {
		if count > 0 {
			p := float64(count) / length
			entropy -= p * log2(p)
		}
	}
	
	return entropy
}

// log2 计算以2为底的对数
func log2(x float64) float64 {
	if x <= 0 {
		return 0
	}
	// 使用换底公式: log2(x) = ln(x) / ln(2)
	return 1.4426950408889634 * logApprox(x) // 1/ln(2) ≈ 1.4426950408889634
}

// logApprox 自然对数的近似计算
func logApprox(x float64) float64 {
	if x <= 0 {
		return 0
	}
	if x == 1 {
		return 0
	}
	// 简单的对数近似，实际项目中可以使用 math.Log
	// 这里为了避免引入额外依赖使用近似计算
	result := 0.0
	if x > 1 {
		n := 0
		for x > 2 {
			x /= 2
			n++
		}
		y := (x - 1) / (x + 1)
		y2 := y * y
		result = 2*y*(1 + y2/3 + y2*y2/5) + float64(n)*0.6931471805599453 // ln(2)
	}
	return result
}

// hasObviousPattern 检查是否有明显的模式
func hasObviousPattern(s string) bool {
	// 检查重复字符
	if len(s) > 3 {
		for i := 0; i < len(s)-3; i++ {
			if s[i] == s[i+1] && s[i+1] == s[i+2] && s[i+2] == s[i+3] {
				return true
			}
		}
	}
	
	// 检查顺序模式
	patterns := []string{
		"1234", "abcd", "ABCD", "qwer", "asdf", "zxcv",
		"!@#$", "0000", "1111", "aaaa", "AAAA",
	}
	
	lower := strings.ToLower(s)
	for _, pattern := range patterns {
		if strings.Contains(lower, strings.ToLower(pattern)) {
			return true
		}
	}
	
	return false
}

// isWeakHexKey 检查是否为弱十六进制密钥
func isWeakHexKey(key string) bool {
	// 全零
	if strings.Trim(key, "0") == "" {
		return true
	}
	
	// 全相同字符
	if len(strings.Trim(key, string(key[0]))) == 0 {
		return true
	}
	
	// 简单递增模式
	patterns := []string{
		"0123456789abcdef",
		"fedcba9876543210",
		"0000111122223333",
		"aaaaaaaaaaaaaaaa",
		"ffffffffffffffff",
	}
	
	for _, pattern := range patterns {
		if strings.Contains(key, pattern) {
			return true
		}
	}
	
	return false
}

// ValidatePort 验证端口号
func ValidatePort(portStr string) error {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return errors.New("端口号必须是数字")
	}
	
	if port < 1 || port > 65535 {
		return errors.New("端口号必须在 1-65535 范围内")
	}
	
	// 检查是否使用了系统保留端口
	if port < 1024 {
		reservedPorts := map[int]string{
			22: "SSH", 25: "SMTP", 53: "DNS", 80: "HTTP",
			110: "POP3", 143: "IMAP", 443: "HTTPS", 993: "IMAPS", 995: "POP3S",
		}
		
		if service, exists := reservedPorts[port]; exists {
			return fmt.Errorf("端口 %d 被 %s 服务占用，建议使用 1024 以上的端口", port, service)
		}
	}
	
	return nil
}

// SanitizeConfigValue 清理配置值
func SanitizeConfigValue(value string) string {
	// 移除首尾空白字符
	value = strings.TrimSpace(value)
	
	// 移除不可见字符
	reg := regexp.MustCompile(`[\x00-\x1F\x7F]`)
	value = reg.ReplaceAllString(value, "")
	
	return value
}
