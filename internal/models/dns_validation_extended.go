package models

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// validateCAARecord 验证CAA记录值
func validateCAARecord(value string) error {
	// CAA记录格式：flags tag "value"
	// 例如：0 issue "letsencrypt.org"
	parts := strings.Fields(value)
	if len(parts) < 3 {
		return errors.New("CAA记录格式应为：flags tag \"value\"")
	}
	
	// 验证flags
	flags, err := strconv.Atoi(parts[0])
	if err != nil || flags < 0 || flags > 255 {
		return errors.New("CAA记录flags必须是0-255之间的数字")
	}
	
	// 验证tag
	validTags := map[string]bool{
		"issue":     true,
		"issuewild": true,
		"iodef":     true,
	}
	
	if !validTags[parts[1]] {
		return errors.New("CAA记录tag必须是issue、issuewild或iodef之一")
	}
	
	// 验证value部分（应该被引号包围）
	valueStart := strings.Index(value, "\"")
	valueEnd := strings.LastIndex(value, "\"")
	if valueStart == -1 || valueEnd == -1 || valueStart == valueEnd {
		return errors.New("CAA记录值必须用双引号包围")
	}
	
	return nil
}

// isReservedIP 检查是否是保留IP地址
func isReservedIP(ip string) bool {
	reservedRanges := []string{
		"0.0.0.0/8",        // "This" network
		"224.0.0.0/4",      // Multicast
		"240.0.0.0/4",      // Reserved for future use
		"255.255.255.255/32", // Broadcast
	}
	
	for _, cidr := range reservedRanges {
		if isIPInCIDR(ip, cidr) {
			return true
		}
	}
	
	// 检查单个保留地址
	reservedIPs := []string{
		"0.0.0.0",
		"255.255.255.255",
	}
	
	for _, reserved := range reservedIPs {
		if ip == reserved {
			return true
		}
	}
	
	return false
}

// ValidateCreateDNSRecordRequest 验证创建DNS记录请求
func (req *CreateDNSRecordRequest) Validate() error {
	// 基础字段验证
	if req.DomainID == 0 {
		return errors.New("域名ID不能为空")
	}
	
	if req.Subdomain == "" {
		return errors.New("子域名不能为空")
	}
	
	if req.Type == "" {
		return errors.New("记录类型不能为空")
	}
	
	if req.Value == "" {
		return errors.New("记录值不能为空")
	}
	
	// 验证子域名格式
	if err := validateSubdomain(req.Subdomain); err != nil {
		return fmt.Errorf("子域名格式错误: %v", err)
	}
	
	// 验证记录类型
	if err := validateDNSType(req.Type); err != nil {
		return fmt.Errorf("DNS记录类型错误: %v", err)
	}
	
	// 验证记录值（考虑私有IP设置）
	if err := req.validateValueWithPrivateIPCheck(); err != nil {
		return fmt.Errorf("DNS记录值错误: %v", err)
	}
	
	// 验证TTL
	if req.TTL != 0 {
		if err := validateTTL(req.TTL); err != nil {
			return fmt.Errorf("TTL值错误: %v", err)
		}
	}
	
	// 验证特定类型的字段
	if err := req.validateTypeSpecificFields(); err != nil {
		return err
	}
	
	return nil
}

// validateValueWithPrivateIPCheck 验证记录值（考虑私有IP设置）
func (req *CreateDNSRecordRequest) validateValueWithPrivateIPCheck() error {
	// 对于A记录，需要特殊处理私有IP检查
	if strings.ToUpper(req.Type) == "A" {
		parsedIP := net.ParseIP(req.Value)
		if parsedIP == nil {
			return errors.New("IPv4地址格式不正确")
		}
		
		if parsedIP.To4() == nil {
			return errors.New("不是有效的IPv4地址")
		}
		
		// 检查保留地址
		if isReservedIP(req.Value) {
			return errors.New("不允许使用保留IP地址")
		}
		
		// 检查私有IP（如果不允许）
		if !req.AllowPrivateIP && isPrivateIP(req.Value) {
			return errors.New("不允许使用私有网络地址，如需使用请勾选允许私有IP选项")
		}
		
		return nil
	}
	
	// 对于其他类型，使用标准验证
	return validateDNSValue(req.Type, req.Value)
}

// validateTypeSpecificFields 验证特定记录类型的字段
func (req *CreateDNSRecordRequest) validateTypeSpecificFields() error {
	switch strings.ToUpper(req.Type) {
	case "MX":
		if req.Priority < 0 || req.Priority > 65535 {
			return fmt.Errorf("MX记录优先级必须在0-65535之间")
		}
	case "SRV":
		if req.Priority < 0 || req.Priority > 65535 {
			return fmt.Errorf("SRV记录优先级必须在0-65535之间")
		}
		if req.Weight < 0 || req.Weight > 65535 {
			return fmt.Errorf("SRV记录权重必须在0-65535之间")
		}
		if req.Port < 1 || req.Port > 65535 {
			return fmt.Errorf("SRV记录端口必须在1-65535之间")
		}
	}
	return nil
}

// ValidateUpdateDNSRecordRequest 验证更新DNS记录请求
func (req *UpdateDNSRecordRequest) Validate() error {
	// 只验证非空字段
	if req.Subdomain != "" {
		if err := validateSubdomain(req.Subdomain); err != nil {
			return fmt.Errorf("子域名格式错误: %v", err)
		}
	}
	
	if req.Type != "" {
		if err := validateDNSType(req.Type); err != nil {
			return fmt.Errorf("DNS记录类型错误: %v", err)
		}
	}
	
	if req.Value != "" {
		if err := req.validateValueWithPrivateIPCheck(); err != nil {
			return fmt.Errorf("DNS记录值错误: %v", err)
		}
	}
	
	if req.TTL != 0 {
		if err := validateTTL(req.TTL); err != nil {
			return fmt.Errorf("TTL值错误: %v", err)
		}
	}
	
	// 验证特定类型的字段
	if err := req.validateTypeSpecificFieldsForUpdate(); err != nil {
		return err
	}
	
	return nil
}

// validateValueWithPrivateIPCheck 验证记录值（更新请求版本）
func (req *UpdateDNSRecordRequest) validateValueWithPrivateIPCheck() error {
	if req.Value == "" {
		return nil
	}
	
	// 对于A记录，需要特殊处理私有IP检查
	if strings.ToUpper(req.Type) == "A" {
		parsedIP := net.ParseIP(req.Value)
		if parsedIP == nil {
			return errors.New("IPv4地址格式不正确")
		}
		
		if parsedIP.To4() == nil {
			return errors.New("不是有效的IPv4地址")
		}
		
		// 检查保留地址
		if isReservedIP(req.Value) {
			return errors.New("不允许使用保留IP地址")
		}
		
		// 检查私有IP（如果不允许）
		if !req.AllowPrivateIP && isPrivateIP(req.Value) {
			return errors.New("不允许使用私有网络地址，如需使用请勾选允许私有IP选项")
		}
		
		return nil
	}
	
	// 对于其他类型，使用标准验证
	return validateDNSValue(req.Type, req.Value)
}

// validateTypeSpecificFieldsForUpdate 验证特定记录类型的字段（更新版本）
func (req *UpdateDNSRecordRequest) validateTypeSpecificFieldsForUpdate() error {
	if req.Type == "" {
		return nil // 如果类型未更新，跳过验证
	}
	
	switch strings.ToUpper(req.Type) {
	case "MX":
		if req.Priority < 0 || req.Priority > 65535 {
			return fmt.Errorf("MX记录优先级必须在0-65535之间")
		}
	case "SRV":
		if req.Priority < 0 || req.Priority > 65535 {
			return fmt.Errorf("SRV记录优先级必须在0-65535之间")
		}
		if req.Weight < 0 || req.Weight > 65535 {
			return fmt.Errorf("SRV记录权重必须在0-65535之间")
		}
		if req.Port < 1 || req.Port > 65535 {
			return fmt.Errorf("SRV记录端口必须在1-65535之间")
		}
	}
	return nil
}

// GetDNSRecordTypeDescription 获取DNS记录类型描述
func GetDNSRecordTypeDescription(recordType string) string {
	descriptions := map[string]string{
		"A":     "IPv4地址记录，将域名指向IPv4地址",
		"AAAA":  "IPv6地址记录，将域名指向IPv6地址",
		"CNAME": "别名记录，将域名指向另一个域名",
		"MX":    "邮件交换记录，指定邮件服务器",
		"TXT":   "文本记录，存储任意文本信息",
		"NS":    "域名服务器记录，指定域名的权威DNS服务器",
		"PTR":   "反向DNS记录，将IP地址指向域名",
		"SRV":   "服务记录，指定特定服务的位置",
		"CAA":   "证书颁发机构授权记录，控制SSL证书颁发",
	}
	
	if desc, exists := descriptions[strings.ToUpper(recordType)]; exists {
		return desc
	}
	return "未知记录类型"
}

// GetTTLOptions 获取TTL选项列表
func GetTTLOptions() []map[string]interface{} {
	return []map[string]interface{}{
		{"value": 60, "label": "1分钟 (60秒)", "description": "适用于频繁变更的记录"},
		{"value": 300, "label": "5分钟 (300秒)", "description": "适用于测试环境"},
		{"value": 600, "label": "10分钟 (600秒)", "description": "默认值，平衡性能和灵活性"},
		{"value": 1800, "label": "30分钟 (1800秒)", "description": "适用于相对稳定的记录"},
		{"value": 3600, "label": "1小时 (3600秒)", "description": "适用于稳定的记录"},
		{"value": 7200, "label": "2小时 (7200秒)", "description": "适用于很少变更的记录"},
		{"value": 14400, "label": "4小时 (14400秒)", "description": "适用于长期稳定的记录"},
		{"value": 43200, "label": "12小时 (43200秒)", "description": "适用于极少变更的记录"},
		{"value": 86400, "label": "1天 (86400秒)", "description": "适用于永久性记录"},
		{"value": 604800, "label": "7天 (604800秒)", "description": "最大值，适用于永不变更的记录"},
	}
}