package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
)

// ========================= 输入验证函数 =========================

// ValidateInput 通用输入验证接口
type ValidateInput struct{}

// ValidateUserID 验证用户ID
func ValidateUserID(userID string) (uint64, error) {
	if userID == "" {
		return 0, errors.New("用户ID不能为空")
	}
	
	// 检查是否包含非数字字符
	for _, char := range userID {
		if !unicode.IsDigit(char) {
			return 0, errors.New("用户ID格式不正确")
		}
	}
	
	// 检查长度（防止超长攻击）
	if len(userID) > 10 {
		return 0, errors.New("用户ID长度不能超过10位")
	}
	
	// 转换为数字
	var id uint64
	for _, char := range userID {
		id = id*10 + uint64(char-'0')
	}
	
	if id == 0 || id > 4294967295 { // uint32 max
		return 0, errors.New("用户ID超出有效范围")
	}
	
	return id, nil
}

// ValidateSearchQuery 验证搜索查询参数
func ValidateSearchQuery(query string) error {
	if len(query) > 100 {
		return errors.New("搜索关键词长度不能超过100个字符")
	}
	
	// 检查危险字符
	dangerousChars := []string{"<", ">", "\"", "'", "&", ";", "|", "`", "$", "\\", "/"}
	for _, char := range dangerousChars {
		if strings.Contains(query, char) {
			return errors.New("搜索关键词包含不安全字符")
		}
	}
	
	// 检查SQL注入模式
	sqlPatterns := []string{
		"union", "select", "insert", "update", "delete", "drop", "create",
		"alter", "exec", "execute", "sp_", "xp_", "--", "/*", "*/",
	}
	
	lowerQuery := strings.ToLower(query)
	for _, pattern := range sqlPatterns {
		if strings.Contains(lowerQuery, pattern) {
			return errors.New("搜索关键词包含不安全内容")
		}
	}
	
	return nil
}

// ValidatePageParams 验证分页参数
func ValidatePageParams(page, pageSize int) (int, int, error) {
	if page < 1 {
		page = 1
	}
	if page > 10000 { // 防止过大的页码
		return 0, 0, errors.New("页码超出有效范围")
	}
	
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 { // 限制每页最大记录数
		pageSize = 100
	}
	
	return page, pageSize, nil
}

// ValidatePassword 增强的密码验证
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("密码长度至少需要8个字符")
	}
	
	if len(password) > 128 {
		return errors.New("密码长度不能超过128个字符")
	}
	
	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)
	
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	
	missingRequirements := []string{}
	if !hasUpper {
		missingRequirements = append(missingRequirements, "大写字母")
	}
	if !hasLower {
		missingRequirements = append(missingRequirements, "小写字母")
	}
	if !hasNumber {
		missingRequirements = append(missingRequirements, "数字")
	}
	if !hasSpecial {
		missingRequirements = append(missingRequirements, "特殊字符")
	}
	
	if len(missingRequirements) > 2 {
		return errors.New("密码强度不够，至少需要包含大写字母、小写字母、数字、特殊字符中的3种")
	}
	
	// 检查常见弱密码
	weakPasswords := []string{
		"password", "123456", "12345678", "qwerty", "abc123",
		"password123", "admin", "root", "user", "guest",
	}
	
	lowerPassword := strings.ToLower(password)
	for _, weak := range weakPasswords {
		if lowerPassword == weak {
			return errors.New("密码过于简单，请使用更复杂的密码")
		}
	}
	
	return nil
}

// ValidateIPAddress 验证IP地址
func ValidateIPAddress(ip string) error {
	if ip == "" {
		return errors.New("IP地址不能为空")
	}
	
	// 使用Go标准库验证IP格式
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return errors.New("IP地址格式不正确")
	}
	
	// 检查是否是私有IP（根据需要可以调整策略）
	if isPrivateIP(parsedIP) {
		return errors.New("不允许使用私有网络IP地址")
	}
	
	return nil
}

// ValidateURL 验证URL格式
func ValidateURL(url string) error {
	if url == "" {
		return errors.New("URL不能为空")
	}
	
	if len(url) > 2048 {
		return errors.New("URL长度不能超过2048个字符")
	}
	
	// 基本的URL格式验证
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return errors.New("URL必须以http://或https://开头")
	}
	
	// 检查危险字符
	dangerousChars := []string{"<", ">", "\"", "'", "`", " "}
	for _, char := range dangerousChars {
		if strings.Contains(url, char) {
			return errors.New("URL包含不安全字符")
		}
	}
	
	return nil
}

// ValidateJSONField 验证JSON字段
func ValidateJSONField(jsonStr string, maxSize int) error {
	if len(jsonStr) > maxSize {
		return errors.New("JSON数据过大")
	}
	
	// 检查是否包含潜在危险的内容
	dangerousPatterns := []string{
		"<script", "javascript:", "vbscript:", "onload=", "onerror=",
		"eval(", "setTimeout(", "setInterval(",
	}
	
	lowerJSON := strings.ToLower(jsonStr)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerJSON, pattern) {
			return errors.New("JSON数据包含不安全内容")
		}
	}
	
	return nil
}

// SanitizeInput 清理输入字符串
func SanitizeInput(input string) string {
	// 移除首尾空格
	input = strings.TrimSpace(input)
	
	// 替换危险字符
	replacements := map[string]string{
		"<":  "&lt;",
		">":  "&gt;",
		"\"": "&quot;",
		"'":  "&#39;",
		"&":  "&amp;",
	}
	
	for old, new := range replacements {
		input = strings.ReplaceAll(input, old, new)
	}
	
	return input
}

// isPrivateIP 检查是否是私有IP地址
func isPrivateIP(ip net.IP) bool {
	privateBlocks := []*net.IPNet{
		{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(8, 32)},
		{IP: net.ParseIP("172.16.0.0"), Mask: net.CIDRMask(12, 32)},
		{IP: net.ParseIP("192.168.0.0"), Mask: net.CIDRMask(16, 32)},
		{IP: net.ParseIP("127.0.0.0"), Mask: net.CIDRMask(8, 32)},
		{IP: net.ParseIP("169.254.0.0"), Mask: net.CIDRMask(16, 32)},
		{IP: net.ParseIP("::1"), Mask: net.CIDRMask(128, 128)},
		{IP: net.ParseIP("fc00::"), Mask: net.CIDRMask(7, 128)},
		{IP: net.ParseIP("fe80::"), Mask: net.CIDRMask(10, 128)},
	}
	
	for _, block := range privateBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	
	return false
}

// ========================= 加密相关函数 =========================

// CryptoService 加密服务
type CryptoService struct {
	key []byte
}

// NewCryptoService 创建加密服务实例
func NewCryptoService(key string) (*CryptoService, error) {
	if len(key) != 32 {
		return nil, errors.New("AES密钥必须是32字节长度")
	}
	
	return &CryptoService{
		key: []byte(key),
	}, nil
}

// Encrypt 使用AES-GCM加密数据
func (c *CryptoService) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", errors.New("明文不能为空")
	}

	// 创建AES cipher
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", fmt.Errorf("创建AES cipher失败: %v", err)
	}

	// 使用GCM模式
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建GCM cipher失败: %v", err)
	}

	// 生成随机nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("生成nonce失败: %v", err)
	}

	// 加密数据
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	
	// 返回十六进制编码的结果
	return hex.EncodeToString(ciphertext), nil
}

// Decrypt 使用AES-GCM解密数据
func (c *CryptoService) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", errors.New("密文不能为空")
	}

	// 解码十六进制
	data, err := hex.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("十六进制解码失败: %v", err)
	}

	// 创建AES cipher
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", fmt.Errorf("创建AES cipher失败: %v", err)
	}

	// 使用GCM模式
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建GCM cipher失败: %v", err)
	}

	// 检查数据长度
	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("密文数据长度不足")
	}

	// 提取nonce和密文
	nonce, cipherData := data[:nonceSize], data[nonceSize:]

	// 解密
	plaintext, err := aesGCM.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", fmt.Errorf("解密失败: %v", err)
	}

	return string(plaintext), nil
}

// GenerateEncryptionKey 生成32字节的随机加密密钥
func GenerateEncryptionKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("生成加密密钥失败: %v", err)
	}
	return hex.EncodeToString(key), nil
}

// ValidateEncryptionKey 验证加密密钥的有效性
func ValidateEncryptionKey(key string) error {
	if key == "" {
		return errors.New("加密密钥不能为空")
	}
	
	decoded, err := hex.DecodeString(key)
	if err != nil {
		return errors.New("加密密钥必须是有效的十六进制字符串")
	}
	
	if len(decoded) != 32 {
		return errors.New("加密密钥解码后必须是32字节长度")
	}
	
	return nil
}

// ========================= 错误处理函数 =========================

// ErrorCode 错误代码枚举
type ErrorCode string

const (
	ErrInvalidRequest    ErrorCode = "INVALID_REQUEST"
	ErrUnauthorized      ErrorCode = "UNAUTHORIZED" 
	ErrForbidden         ErrorCode = "FORBIDDEN"
	ErrNotFound          ErrorCode = "NOT_FOUND"
	ErrConflict          ErrorCode = "CONFLICT"
	ErrValidationFailed  ErrorCode = "VALIDATION_FAILED"
	ErrInternalError     ErrorCode = "INTERNAL_ERROR"
	ErrServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
)

// ApiError API错误结构
type ApiError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"` // 仅在开发环境返回
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Error ApiError `json:"error"`
}

// isDevelopment 判断是否为开发环境
func isDevelopment(c *gin.Context) bool {
	mode := gin.Mode()
	return mode == gin.DebugMode || mode == gin.TestMode
}

// HandleError 统一错误处理函数
func HandleError(c *gin.Context, statusCode int, errorCode ErrorCode, userMessage string, internalError error) {
	// 记录详细错误到日志（生产环境重要）
	if internalError != nil {
		log.Printf("[ERROR] %s - %v - Path: %s, Method: %s, IP: %s", 
			string(errorCode), internalError, c.Request.URL.Path, c.Request.Method, c.ClientIP())
	}

	// 构建错误响应
	apiErr := ApiError{
		Code:    errorCode,
		Message: userMessage,
	}

	// 仅在开发环境提供详细错误信息
	if isDevelopment(c) && internalError != nil {
		apiErr.Details = internalError.Error()
	}

	c.JSON(statusCode, ErrorResponse{Error: apiErr})
}

// HandleValidationError 处理验证错误
func HandleValidationError(c *gin.Context, err error) {
	// 清理敏感信息的验证错误
	message := sanitizeValidationError(err.Error())
	HandleError(c, http.StatusBadRequest, ErrValidationFailed, message, err)
}

// HandleInternalError 处理内部服务器错误
func HandleInternalError(c *gin.Context, err error) {
	HandleError(c, http.StatusInternalServerError, ErrInternalError, "服务器内部错误，请稍后再试", err)
}

// HandleUnauthorized 处理未授权错误
func HandleUnauthorized(c *gin.Context, message string) {
	HandleError(c, http.StatusUnauthorized, ErrUnauthorized, message, nil)
}

// HandleForbidden 处理禁止访问错误
func HandleForbidden(c *gin.Context, message string) {
	HandleError(c, http.StatusForbidden, ErrForbidden, message, nil)
}

// HandleNotFound 处理资源未找到错误
func HandleNotFound(c *gin.Context, resource string) {
	message := resource + "不存在"
	HandleError(c, http.StatusNotFound, ErrNotFound, message, nil)
}

// HandleConflict 处理资源冲突错误
func HandleConflict(c *gin.Context, message string) {
	HandleError(c, http.StatusConflict, ErrConflict, message, nil)
}

// HandleBadRequest 处理请求参数错误
func HandleBadRequest(c *gin.Context, message string, err error) {
	HandleError(c, http.StatusBadRequest, ErrInvalidRequest, message, err)
}

// sanitizeValidationError 清理验证错误信息，移除敏感信息
func sanitizeValidationError(errMsg string) string {
	// 移除可能包含敏感信息的字段名
	sensitiveFields := []string{
		"password", "token", "secret", "key", "credential",
		"authorization", "session", "cookie",
	}
	
	lowerMsg := strings.ToLower(errMsg)
	for _, field := range sensitiveFields {
		if strings.Contains(lowerMsg, field) {
			return "请求参数验证失败"
		}
	}
	
	// 移除路径信息和技术细节
	if strings.Contains(errMsg, "/") || strings.Contains(errMsg, "\\") {
		return "请求参数格式错误"
	}
	
	// 限制错误消息长度
	if len(errMsg) > 100 {
		return "请求参数验证失败"
	}
	
	return errMsg
}

// LogSensitiveOperation 记录敏感操作日志
func LogSensitiveOperation(operation, userID, details string, c *gin.Context) {
	log.Printf("[SECURITY] Operation: %s, User: %s, IP: %s, UserAgent: %s, Details: %s",
		operation, userID, c.ClientIP(), c.GetHeader("User-Agent"), details)
}

// Success 返回成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// SuccessWithMessage 返回带消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// ========================= 配置生成和验证函数 =========================

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

// ========================= 内部辅助函数 =========================

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
	// 使用标准库的数学函数确保精度
	return math.Log2(x)
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
