package utils

import (
	"errors"
	"net"
	"regexp"
	"strings"
	"unicode"
)

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
