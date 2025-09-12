package models

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Domain 域名模型
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

// DNSRecord DNS记录模型
type DNSRecord struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	UserID     uint           `json:"user_id" gorm:"not null;index"`                           // 添加索引提升查询性能
	DomainID   uint           `json:"domain_id" gorm:"not null;index"`                         // 添加索引
	Subdomain  string         `json:"subdomain" gorm:"not null;size:63"`                       // 子域名，限制长度
	Type       string         `json:"type" gorm:"not null;size:10;index"`                      // 记录类型，添加索引
	Value      string         `json:"value" gorm:"not null;size:500"`                          // 记录值，限制长度
	TTL        int            `json:"ttl" gorm:"default:600;check:ttl >= 1 AND ttl <= 604800"` // TTL范围检查
	Priority   int            `json:"priority" gorm:"default:0"`                               // MX和SRV记录优先级
	Weight     int            `json:"weight" gorm:"default:0"`                                 // SRV记录权重
	Port       int            `json:"port" gorm:"default:0"`                                   // SRV记录端口
	ExternalID string         `json:"external_id" gorm:"size:100"`                             // DNS服务商记录ID
	Status     string         `json:"status" gorm:"default:active;size:20"`                    // 记录状态
	Comment    string         `json:"comment" gorm:"size:500"`                                 // 记录备注，增加长度
	CreatedAt  time.Time      `json:"created_at" gorm:"index"`                                 // 添加时间索引
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Domain Domain `json:"domain,omitempty" gorm:"foreignKey:DomainID;constraint:OnDelete:CASCADE"`
}

// DNSProvider DNS服务商模型
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

// 请求和响应结构体

// CreateDNSRecordRequest DNS记录创建请求
type CreateDNSRecordRequest struct {
	DomainID       uint   `json:"domain_id" binding:"required"`
	Subdomain      string `json:"subdomain" binding:"required"`
	Type           string `json:"type" binding:"required,oneof=A AAAA CNAME TXT MX NS PTR SRV CAA"`
	Value          string `json:"value" binding:"required"`
	TTL            int    `json:"ttl"`
	Priority       int    `json:"priority"`        // MX和SRV记录的优先级
	Weight         int    `json:"weight"`          // SRV记录的权重
	Port           int    `json:"port"`            // SRV记录的端口
	Comment        string `json:"comment"`         // 记录备注
	AllowPrivateIP bool   `json:"allow_private_ip"` // 是否允许私有IP
}

// UpdateDNSRecordRequest DNS记录更新请求
type UpdateDNSRecordRequest struct {
	Subdomain      string `json:"subdomain"`
	Type           string `json:"type" binding:"oneof=A AAAA CNAME TXT MX NS PTR SRV CAA"`
	Value          string `json:"value"`
	TTL            int    `json:"ttl"`
	Priority       int    `json:"priority"`
	Weight         int    `json:"weight"`
	Port           int    `json:"port"`
	Comment        string `json:"comment"`
	AllowPrivateIP bool   `json:"allow_private_ip"`
}

// BatchDNSRecordRequest DNS记录批量操作请求
type BatchDNSRecordRequest struct {
	Records []CreateDNSRecordRequest `json:"records" binding:"required,min=1,max=50"`
}

// DNSRecordExportResponse DNS记录导出响应
type DNSRecordExportResponse struct {
	Records []DNSRecordExport `json:"records"`
	Total   int               `json:"total"`
}

// DNSRecordExport DNS记录导出格式
type DNSRecordExport struct {
	Subdomain string `json:"subdomain"`
	Type      string `json:"type"`
	Value     string `json:"value"`
	TTL       int    `json:"ttl"`
	Priority  int    `json:"priority,omitempty"`
	Weight    int    `json:"weight,omitempty"`
	Port      int    `json:"port,omitempty"`
	Comment   string `json:"comment,omitempty"`
	Domain    string `json:"domain"`
}

// ========================= DNS记录验证方法 =========================

// ValidateDNSRecord 验证DNS记录的有效性
func (r *DNSRecord) ValidateDNSRecord() error {
	// 验证子域名格式
	if err := validateSubdomain(r.Subdomain); err != nil {
		return fmt.Errorf("子域名格式错误: %v", err)
	}
	
	// 验证记录类型
	if err := validateDNSType(r.Type); err != nil {
		return fmt.Errorf("DNS记录类型错误: %v", err)
	}
	
	// 根据记录类型验证记录值
	if err := validateDNSValue(r.Type, r.Value); err != nil {
		return fmt.Errorf("DNS记录值错误: %v", err)
	}
	
	// 验证TTL值
	if err := validateTTL(r.TTL); err != nil {
		return fmt.Errorf("TTL值错误: %v", err)
	}
	
	// 验证特定记录类型的额外字段
	if err := r.validateTypeSpecificFields(); err != nil {
		return err
	}
	
	return nil
}

// validateTypeSpecificFields 验证特定记录类型的字段
func (r *DNSRecord) validateTypeSpecificFields() error {
	switch strings.ToUpper(r.Type) {
	case "MX":
		if r.Priority < 0 || r.Priority > 65535 {
			return fmt.Errorf("MX记录优先级必须在0-65535之间")
		}
	case "SRV":
		if r.Priority < 0 || r.Priority > 65535 {
			return fmt.Errorf("SRV记录优先级必须在0-65535之间")
		}
		if r.Weight < 0 || r.Weight > 65535 {
			return fmt.Errorf("SRV记录权重必须在0-65535之间")
		}
		if r.Port < 1 || r.Port > 65535 {
			return fmt.Errorf("SRV记录端口必须在1-65535之间")
		}
	}
	return nil
}

// validateSubdomain 验证子域名格式
func validateSubdomain(subdomain string) error {
	if len(subdomain) == 0 {
		return errors.New("子域名不能为空")
	}
	
	if len(subdomain) > 63 {
		return errors.New("子域名长度不能超过63个字符")
	}
	
	// 支持通配符子域名
	if subdomain == "*" {
		return nil
	}
	
	// 支持下划线（用于某些特殊记录如_dmarc, _spf等）
	validSubdomainPattern := regexp.MustCompile(`^[a-zA-Z0-9_]([a-zA-Z0-9\-_]*[a-zA-Z0-9_])?$`)
	if !validSubdomainPattern.MatchString(subdomain) {
		return errors.New("子域名只能包含字母、数字、连字符和下划线，且不能以连字符开头或结尾")
	}
	
	// 放宽保留名称检查，只检查真正危险的名称
	dangerousNames := []string{"localhost", "broadcasthost"}
	for _, dangerous := range dangerousNames {
		if strings.EqualFold(subdomain, dangerous) {
			return fmt.Errorf("子域名 '%s' 是系统保留名称，不允许使用", subdomain)
		}
	}
	
	return nil
}

// validateDNSType 验证DNS记录类型
func validateDNSType(dnsType string) error {
	validTypes := map[string]bool{
		"A":     true,
		"AAAA":  true,
		"CNAME": true,
		"MX":    true,
		"TXT":   true,
		"NS":    true,
		"PTR":   true,
		"SRV":   true,
		"CAA":   true, // 证书颁发机构授权记录
	}
	
	if !validTypes[strings.ToUpper(dnsType)] {
		return fmt.Errorf("不支持的DNS记录类型: %s", dnsType)
	}
	
	return nil
}

// validateDNSValue 根据DNS记录类型验证记录值
func validateDNSValue(dnsType, value string) error {
	if len(value) == 0 {
		return errors.New("DNS记录值不能为空")
	}
	
	if len(value) > 1000 { // 增加长度限制以支持更长的TXT记录
		return errors.New("DNS记录值长度不能超过1000个字符")
	}
	
	switch strings.ToUpper(dnsType) {
	case "A":
		return validateIPv4Address(value)
	case "AAAA":
		return validateIPv6Address(value)
	case "CNAME":
		return validateDomainName(value)
	case "MX":
		return validateMXRecord(value)
	case "TXT":
		return validateTXTRecord(value)
	case "NS":
		return validateDomainName(value)
	case "PTR":
		return validateDomainName(value)
	case "SRV":
		return validateSRVRecord(value)
	case "CAA":
		return validateCAARecord(value)
	default:
		// 对于其他类型，只进行基本的长度检查
		return nil
	}
}

// validateIPv4Address 验证IPv4地址
func validateIPv4Address(ip string) error {
	// 使用标准库验证IP地址格式
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return errors.New("IPv4地址格式不正确")
	}
	
	// 确保是IPv4地址
	if parsedIP.To4() == nil {
		return errors.New("不是有效的IPv4地址")
	}
	
	// 检查是否是保留地址
	if isReservedIP(ip) {
		return errors.New("不允许使用保留IP地址（如0.0.0.0、255.255.255.255等）")
	}
	
	return nil
}

// validateIPv6Address 验证IPv6地址
func validateIPv6Address(ip string) error {
	// 简单的IPv6格式检查
	if strings.Count(ip, ":") < 2 || strings.Count(ip, ":") > 7 {
		return errors.New("IPv6地址格式不正确")
	}
	
	// 检查是否包含非法字符
	validIPv6Pattern := regexp.MustCompile(`^[0-9a-fA-F:]+$`)
	if !validIPv6Pattern.MatchString(ip) {
		return errors.New("IPv6地址包含非法字符")
	}
	
	return nil
}

// validateDomainName 验证域名格式
func validateDomainName(domain string) error {
	if len(domain) == 0 {
		return errors.New("域名不能为空")
	}
	
	if len(domain) > 253 {
		return errors.New("域名长度不能超过253个字符")
	}
	
	// 检查域名格式
	domainPattern := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-\.]*[a-zA-Z0-9])?$`)
	if !domainPattern.MatchString(domain) {
		return errors.New("域名格式不正确")
	}
	
	// 检查是否以点结尾（完全限定域名）
	if strings.HasSuffix(domain, ".") {
		domain = domain[:len(domain)-1] // 移除尾部的点进行验证
	}
	
	// 验证每个标签
	labels := strings.Split(domain, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return errors.New("域名标签长度必须在1-63个字符之间")
		}
		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return errors.New("域名标签不能以连字符开头或结尾")
		}
	}
	
	return nil
}

// validateMXRecord 验证MX记录值
func validateMXRecord(value string) error {
	// MX记录值格式：优先级 邮件服务器域名
	parts := strings.Fields(value)
	if len(parts) != 2 {
		return errors.New("MX记录格式应为：优先级 邮件服务器域名")
	}
	
	// 验证优先级
	priority, err := strconv.Atoi(parts[0])
	if err != nil || priority < 0 || priority > 65535 {
		return errors.New("MX记录优先级必须是0-65535之间的数字")
	}
	
	// 验证邮件服务器域名
	return validateDomainName(parts[1])
}

// validateTXTRecord 验证TXT记录值
func validateTXTRecord(value string) error {
	// TXT记录可以包含任意文本，但需要检查长度和一些安全问题
	if len(value) > 255 {
		return errors.New("单个TXT记录长度不能超过255个字符")
	}
	
	// 检查是否包含危险的脚本标签
	dangerousPatterns := []string{"<script", "javascript:", "vbscript:", "onload=", "onerror="}
	lowerValue := strings.ToLower(value)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerValue, pattern) {
			return errors.New("TXT记录不能包含脚本内容")
		}
	}
	
	return nil
}

// validateSRVRecord 验证SRV记录值
func validateSRVRecord(value string) error {
	// SRV记录格式：优先级 权重 端口 目标域名
	parts := strings.Fields(value)
	if len(parts) != 4 {
		return errors.New("SRV记录格式应为：优先级 权重 端口 目标域名")
	}
	
	// 验证优先级
	if priority, err := strconv.Atoi(parts[0]); err != nil || priority < 0 || priority > 65535 {
		return errors.New("SRV记录优先级必须是0-65535之间的数字")
	}
	
	// 验证权重
	if weight, err := strconv.Atoi(parts[1]); err != nil || weight < 0 || weight > 65535 {
		return errors.New("SRV记录权重必须是0-65535之间的数字")
	}
	
	// 验证端口
	if port, err := strconv.Atoi(parts[2]); err != nil || port < 1 || port > 65535 {
		return errors.New("SRV记录端口必须是1-65535之间的数字")
	}
	
	// 验证目标域名
	return validateDomainName(parts[3])
}

// validateCAARecord 验证CAA记录值
func validateCAARecord(value string) error {
	// CAA记录格式：flags tag value
	parts := strings.Fields(value)
	if len(parts) < 3 {
		return errors.New("CAA记录格式应为：flags tag value")
	}
	
	// 验证flags
	if flags, err := strconv.Atoi(parts[0]); err != nil || flags < 0 || flags > 255 {
		return errors.New("CAA记录flags必须是0-255之间的数字")
	}
	
	// 验证tag
	validTags := []string{"issue", "issuewild", "iodef"}
	tagValid := false
	for _, validTag := range validTags {
		if parts[1] == validTag {
			tagValid = true
			break
		}
	}
	if !tagValid {
		return errors.New("CAA记录tag必须是issue、issuewild或iodef之一")
	}
	
	return nil
}

// validateTTL 验证TTL值
func validateTTL(ttl int) error {
	if ttl < 1 {
		return errors.New("TTL值不能小于1秒")
	}
	
	if ttl > 604800 { // 7天
		return errors.New("TTL值不能超过604800秒（7天）")
	}
	
	return nil
}

// isReservedIP 检查是否是保留IP地址
func isReservedIP(ip string) bool {
	reservedIPs := []string{
		"0.0.0.0",
		"255.255.255.255",
		"127.0.0.1",
	}
	
	for _, reserved := range reservedIPs {
		if ip == reserved {
			return true
		}
	}
	
	return false
}