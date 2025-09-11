package models

import (
	"errors"
	"regexp"
	"strings"
)

// ValidateDNSRecord 验证DNS记录的数据完整性
func (d *DNSRecord) ValidateDNSRecord() error {
	// 子域名验证
	if err := validateSubdomain(d.Subdomain); err != nil {
		return err
	}

	// 记录类型验证
	if err := validateRecordType(d.Type); err != nil {
		return err
	}

	// 记录值验证
	if err := validateRecordValue(d.Type, d.Value); err != nil {
		return err
	}

	// TTL验证
	if d.TTL < 1 || d.TTL > 604800 {
		return errors.New("TTL值必须在1-604800秒之间")
	}

	return nil
}

// ValidateUser 验证用户数据
func (u *User) ValidateUser() error {
	// 邮箱格式验证
	if !isValidEmail(u.Email) {
		return errors.New("邮箱格式不正确")
	}

	// 用户状态验证
	validStatuses := []string{"normal", "suspended", "banned"}
	if !contains(validStatuses, u.Status) {
		return errors.New("用户状态无效")
	}

	// DNS记录配额验证
	if u.DNSRecordQuota < 0 {
		return errors.New("DNS记录配额不能为负数")
	}

	return nil
}

// ValidateDomain 验证域名数据
func (d *Domain) ValidateDomain() error {
	// 域名格式验证
	if !isValidDomainName(d.Name) {
		return errors.New("域名格式不正确")
	}

	return nil
}

// ValidateDNSProvider 验证DNS服务商配置
func (d *DNSProvider) ValidateDNSProvider() error {
	// 服务商类型验证
	validTypes := []string{"dnspod", "dnspod_v3", "cloudflare", "aliyun"}
	if !contains(validTypes, d.Type) {
		return errors.New("不支持的DNS服务商类型")
	}

	// 名称不能为空
	if strings.TrimSpace(d.Name) == "" {
		return errors.New("服务商名称不能为空")
	}

	// 配置不能为空
	if strings.TrimSpace(d.Config) == "" {
		return errors.New("服务商配置不能为空")
	}

	return nil
}

// validateSubdomain 验证子域名格式
func validateSubdomain(subdomain string) error {
	if subdomain == "" {
		return errors.New("子域名不能为空")
	}

	// 子域名长度限制
	if len(subdomain) > 63 {
		return errors.New("子域名长度不能超过63个字符")
	}

	// 子域名格式验证
	subdomainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]*[a-zA-Z0-9])?$`)
	if !subdomainRegex.MatchString(subdomain) {
		return errors.New("子域名格式不正确，只能包含字母、数字和连字符，且不能以连字符开头或结尾")
	}

	return nil
}

// validateRecordType 验证DNS记录类型
func validateRecordType(recordType string) error {
	validTypes := []string{"A", "AAAA", "CNAME", "TXT", "MX", "NS", "SRV", "CAA"}
	if !contains(validTypes, recordType) {
		return errors.New("不支持的DNS记录类型")
	}
	return nil
}

// validateRecordValue 根据记录类型验证记录值
func validateRecordValue(recordType, value string) error {
	if value == "" {
		return errors.New("记录值不能为空")
	}

	switch recordType {
	case "A":
		return validateIPv4(value)
	case "AAAA":
		return validateIPv6(value)
	case "CNAME":
		return validateDomainName(value)
	case "MX":
		return validateMXRecord(value)
	case "TXT":
		return validateTXTRecord(value)
	default:
		// 对于其他类型，只做基础验证
		if len(value) > 500 {
			return errors.New("记录值长度不能超过500个字符")
		}
	}

	return nil
}

// validateIPv4 验证IPv4地址
func validateIPv4(ip string) error {
	ipRegex := regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	if !ipRegex.MatchString(ip) {
		return errors.New("无效的IPv4地址格式")
	}

	// 验证每个部分是否在0-255范围内
	parts := strings.Split(ip, ".")
	for _, part := range parts {
		if len(part) > 1 && part[0] == '0' {
			return errors.New("IPv4地址不能有前导零")
		}
		// 这里可以添加更详细的范围检查
	}

	return nil
}

// validateIPv6 验证IPv6地址（简化版）
func validateIPv6(ip string) error {
	ipv6Regex := regexp.MustCompile(`^([0-9a-fA-F]{0,4}:){1,7}[0-9a-fA-F]{0,4}$`)
	if !ipv6Regex.MatchString(ip) {
		return errors.New("无效的IPv6地址格式")
	}
	return nil
}

// validateDomainName 验证域名格式
func validateDomainName(domain string) error {
	if !isValidDomainName(domain) {
		return errors.New("无效的域名格式")
	}
	return nil
}

// validateMXRecord 验证MX记录格式
func validateMXRecord(value string) error {
	// MX记录格式通常是：优先级 主机名
	parts := strings.Fields(value)
	if len(parts) != 2 {
		return errors.New("MX记录格式错误，应为：优先级 主机名")
	}

	// 验证优先级是否为数字
	priorityRegex := regexp.MustCompile(`^\d+$`)
	if !priorityRegex.MatchString(parts[0]) {
		return errors.New("MX记录优先级必须是数字")
	}

	// 验证主机名格式
	return validateDomainName(parts[1])
}

// validateTXTRecord 验证TXT记录
func validateTXTRecord(value string) error {
	// TXT记录长度限制
	if len(value) > 255 {
		return errors.New("TXT记录长度不能超过255个字符")
	}
	return nil
}

// isValidEmail 验证邮箱格式
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// isValidDomainName 验证域名格式
func isValidDomainName(domain string) bool {
	if len(domain) == 0 || len(domain) > 253 {
		return false
	}

	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	return domainRegex.MatchString(domain)
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
