package providers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// 腾讯云API参数类型定义
// 严格按照腾讯云API 3.0参数类型规范实现

// DNSPod V3 API请求参数结构

// DescribeDomainListRequest 查询域名列表请求
type DescribeDomainListRequest struct {
	// String类型：关键字搜索
	Keyword *string `json:"Keyword,omitempty"`
	// Integer类型：返回数量限制，最大值100
	Limit *uint64 `json:"Limit,omitempty"`
	// Integer类型：偏移量，用于分页
	Offset *uint64 `json:"Offset,omitempty"`
}

// DescribeRecordListRequest 查询记录列表请求
type DescribeRecordListRequest struct {
	// String类型：域名
	Domain string `json:"Domain"`
	// Integer类型：返回数量限制，最大值3000
	Limit *uint64 `json:"Limit,omitempty"`
	// Integer类型：偏移量，用于分页
	Offset *uint64 `json:"Offset,omitempty"`
	// String类型：子域名，用于精确搜索
	Subdomain *string `json:"Subdomain,omitempty"`
	// String类型：记录类型过滤
	RecordType *string `json:"RecordType,omitempty"`
}

// CreateRecordRequest 创建记录请求
type CreateRecordRequest struct {
	// String类型：域名
	Domain string `json:"Domain"`
	// String类型：子域名
	SubDomain string `json:"SubDomain"`
	// String类型：记录类型 (A, CNAME, TXT, MX等)
	RecordType string `json:"RecordType"`
	// String类型：线路类型，默认为"默认"
	RecordLine *string `json:"RecordLine,omitempty"`
	// String类型：记录值
	Value string `json:"Value"`
	// Integer类型：TTL值，范围1-604800
	TTL *uint64 `json:"TTL,omitempty"`
	// Integer类型：MX优先级，仅MX记录有效
	MX *uint64 `json:"MX,omitempty"`
}

// ModifyRecordRequest 修改记录请求
type ModifyRecordRequest struct {
	// String类型：域名
	Domain string `json:"Domain"`
	// Integer类型：记录ID
	RecordId uint64 `json:"RecordId"`
	// String类型：子域名
	SubDomain string `json:"SubDomain"`
	// String类型：记录类型
	RecordType string `json:"RecordType"`
	// String类型：线路类型
	RecordLine *string `json:"RecordLine,omitempty"`
	// String类型：记录值
	Value string `json:"Value"`
	// Integer类型：TTL值
	TTL *uint64 `json:"TTL,omitempty"`
	// Integer类型：MX优先级
	MX *uint64 `json:"MX,omitempty"`
}

// DeleteRecordRequest 删除记录请求
type DeleteRecordRequest struct {
	// String类型：域名
	Domain string `json:"Domain"`
	// Integer类型：记录ID
	RecordId uint64 `json:"RecordId"`
}

// 腾讯云API响应数据结构

// DomainInfo 域名信息
type DomainInfo struct {
	DomainId     uint64    `json:"DomainId"`     // Integer：域名ID
	Name         string    `json:"Name"`         // String：域名名称
	Status       string    `json:"Status"`       // String：域名状态
	TTL          uint64    `json:"TTL"`          // Integer：默认TTL
	CNAMESpeedup string    `json:"CNAMESpeedup"` // String：CNAME加速状态
	DNSStatus    string    `json:"DNSStatus"`    // String：DNS状态
	Grade        string    `json:"Grade"`        // String：域名等级
	GroupId      uint64    `json:"GroupId"`      // Integer：分组ID
	SearchEnginePush string `json:"SearchEnginePush"` // String：搜索引擎推送状态
	Remark       string    `json:"Remark"`       // String：备注
	CreatedOn    string    `json:"CreatedOn"`    // Timestamp：创建时间
	UpdatedOn    string    `json:"UpdatedOn"`    // Timestamp：更新时间
}

// RecordInfo 记录信息
type RecordInfo struct {
	RecordId    uint64 `json:"RecordId"`    // Integer：记录ID
	Name        string `json:"Name"`        // String：子域名
	Type        string `json:"Type"`        // String：记录类型
	Value       string `json:"Value"`       // String：记录值
	TTL         uint64 `json:"TTL"`         // Integer：TTL值
	Status      string `json:"Status"`      // String：记录状态
	Line        string `json:"Line"`        // String：线路类型
	LineId      string `json:"LineId"`      // String：线路ID
	Weight      uint64 `json:"Weight"`      // Integer：权重
	MonitorStatus string `json:"MonitorStatus"` // String：监控状态
	Remark      string `json:"Remark"`      // String：备注
	UpdatedOn   string `json:"UpdatedOn"`   // Timestamp：更新时间
	DomainId    uint64 `json:"DomainId"`    // Integer：域名ID
}

// 参数类型转换辅助函数

// StringPtr 创建String指针，用于可选参数
func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Uint64Ptr 创建Uint64指针，用于可选参数
func Uint64Ptr(i uint64) *uint64 {
	if i == 0 {
		return nil
	}
	return &i
}

// IntToUint64Ptr 将int转换为uint64指针
func IntToUint64Ptr(i int) *uint64 {
	if i <= 0 {
		return nil
	}
	val := uint64(i)
	return &val
}

// TimestampToString 将time.Time转换为腾讯云API的Timestamp字符串格式
func TimestampToString(t time.Time) string {
	// 腾讯云API使用格式：2022-01-01 00:00:00
	return t.Format("2006-01-02 15:04:05")
}

// DateToString 将time.Time转换为腾讯云API的Date字符串格式
func DateToString(t time.Time) string {
	// 腾讯云API使用格式：2022-01-01
	return t.Format("2006-01-02")
}

// ISO8601ToString 将time.Time转换为ISO8601格式
func ISO8601ToString(t time.Time) string {
	// ISO8601格式：2022-01-01T00:00:00+08:00
	return t.Format(time.RFC3339)
}

// StringToTimestamp 将腾讯云API的Timestamp字符串转换为time.Time
func StringToTimestamp(s string) (time.Time, error) {
	// 解析格式：2022-01-01 00:00:00
	return time.Parse("2006-01-02 15:04:05", s)
}

// StringToDate 将腾讯云API的Date字符串转换为time.Time
func StringToDate(s string) (time.Time, error) {
	// 解析格式：2022-01-01
	return time.Parse("2006-01-02", s)
}

// ISO8601ToTime 将ISO8601字符串转换为time.Time
func ISO8601ToTime(s string) (time.Time, error) {
	// 解析ISO8601格式
	return time.Parse(time.RFC3339, s)
}

// 参数验证函数

// ValidateStringLength 验证字符串长度
func ValidateStringLength(s string, min, max int) error {
	length := len(s)
	if length < min {
		return fmt.Errorf("字符串长度不能少于%d位", min)
	}
	if max > 0 && length > max {
		return fmt.Errorf("字符串长度不能超过%d位", max)
	}
	return nil
}

// ValidateIntegerRange 验证整数范围
func ValidateIntegerRange(value uint64, min, max uint64) error {
	if value < min {
		return fmt.Errorf("数值不能小于%d", min)
	}
	if max > 0 && value > max {
		return fmt.Errorf("数值不能大于%d", max)
	}
	return nil
}

// ValidateRecordType 验证DNS记录类型
func ValidateRecordType(recordType string) error {
	validTypes := []string{"A", "AAAA", "CNAME", "TXT", "MX", "NS", "SRV", "CAA"}
	for _, validType := range validTypes {
		if recordType == validType {
			return nil
		}
	}
	return fmt.Errorf("不支持的记录类型: %s", recordType)
}

// ValidateTTL 验证TTL值
func ValidateTTL(ttl uint64) error {
	// TTL范围：1-604800秒（7天）
	return ValidateIntegerRange(ttl, 1, 604800)
}

// ValidateDomainName 验证域名格式
func ValidateDomainName(domain string) error {
	if err := ValidateStringLength(domain, 1, 253); err != nil {
		return fmt.Errorf("域名%v", err)
	}
	
	// 简单的域名格式验证
	if len(domain) == 0 {
		return fmt.Errorf("域名不能为空")
	}
	
	// 更详细的域名格式验证可以在这里添加
	return nil
}

// ValidateSubdomain 验证子域名格式
func ValidateSubdomain(subdomain string) error {
	if err := ValidateStringLength(subdomain, 1, 63); err != nil {
		return fmt.Errorf("子域名%v", err)
	}
	
	// 子域名不能包含某些特殊字符
	// 更详细的验证逻辑可以在这里添加
	return nil
}

// 类型转换辅助函数

// SafeStringToUint64 安全地将字符串转换为uint64
func SafeStringToUint64(s string) (uint64, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.ParseUint(s, 10, 64)
}

// SafeUint64ToString 安全地将uint64转换为字符串
func SafeUint64ToString(i uint64) string {
	return strconv.FormatUint(i, 10)
}

// SafeIntToUint64 安全地将int转换为uint64
func SafeIntToUint64(i int) uint64 {
	if i < 0 {
		return 0
	}
	return uint64(i)
}

// SafeUint64ToInt 安全地将uint64转换为int
func SafeUint64ToInt(i uint64) int {
	// 检查溢出
	if i > uint64(^int(0)>>1) {
		return int(^int(0)>>1) // 返回int最大值
	}
	return int(i)
}

// JSON参数构建辅助函数

// BuildCreateRecordParams 构建创建记录的参数
func BuildCreateRecordParams(domain, subdomain, recordType, value string, ttl int) map[string]interface{} {
	params := map[string]interface{}{
		"Domain":     domain,     // String
		"SubDomain":  subdomain,  // String
		"RecordType": recordType, // String
		"Value":      value,      // String
	}
	
	// 可选参数
	if ttl > 0 {
		params["TTL"] = SafeIntToUint64(ttl) // Integer (uint64)
	}
	
	// 设置默认线路
	params["RecordLine"] = "默认" // String
	
	return params
}

// BuildModifyRecordParams 构建修改记录的参数
func BuildModifyRecordParams(domain, recordID, subdomain, recordType, value string, ttl int) (map[string]interface{}, error) {
	// 将recordID从string转换为uint64
	recordId, err := SafeStringToUint64(recordID)
	if err != nil {
		return nil, fmt.Errorf("无效的记录ID: %v", err)
	}
	
	params := map[string]interface{}{
		"Domain":     domain,     // String
		"RecordId":   recordId,   // Integer (uint64)
		"SubDomain":  subdomain,  // String
		"RecordType": recordType, // String
		"Value":      value,      // String
	}
	
	// 可选参数
	if ttl > 0 {
		params["TTL"] = SafeIntToUint64(ttl) // Integer (uint64)
	}
	
	// 设置默认线路
	params["RecordLine"] = "默认" // String
	
	return params, nil
}

// BuildDeleteRecordParams 构建删除记录的参数
func BuildDeleteRecordParams(domain, recordID string) (map[string]interface{}, error) {
	// 将recordID从string转换为uint64
	recordId, err := SafeStringToUint64(recordID)
	if err != nil {
		return nil, fmt.Errorf("无效的记录ID: %v", err)
	}
	
	params := map[string]interface{}{
		"Domain":   domain,   // String
		"RecordId": recordId, // Integer (uint64)
	}
	
	return params, nil
}

// BuildDescribeRecordListParams 构建查询记录列表的参数
func BuildDescribeRecordListParams(domain string) map[string]interface{} {
	params := map[string]interface{}{
		"Domain": domain,      // String
		"Limit":  uint64(100), // Integer (uint64) - 每页最多100条
		"Offset": uint64(0),   // Integer (uint64) - 从第一条开始
	}
	
	return params
}

// BuildDescribeDomainListParams 构建查询域名列表的参数
func BuildDescribeDomainListParams(keyword string) map[string]interface{} {
	params := map[string]interface{}{
		"Limit":  uint64(20), // Integer (uint64)
		"Offset": uint64(0),  // Integer (uint64)
	}
	
	// 可选参数
	if keyword != "" {
		params["Keyword"] = keyword // String
	}
	
	return params
}

// 参数验证函数

// ValidateCreateRecordParams 验证创建记录参数
func ValidateCreateRecordParams(domain, subdomain, recordType, value string, ttl int) error {
	// 验证域名
	if err := ValidateDomainName(domain); err != nil {
		return err
	}
	
	// 验证子域名
	if err := ValidateSubdomain(subdomain); err != nil {
		return err
	}
	
	// 验证记录类型和值的组合
	if err := ValidateRecordTypeAndValue(recordType, value); err != nil {
		return err
	}
	
	// 验证TTL
	if ttl > 0 {
		if err := ValidateTTL(uint64(ttl)); err != nil {
			return err
		}
	} else {
		// 如果TTL为0或负数，使用默认值
		ttl = int(GetDefaultTTL(recordType))
	}
	
	return nil
}

// ValidateModifyRecordParams 验证修改记录参数
func ValidateModifyRecordParams(domain, recordID, subdomain, recordType, value string, ttl int) error {
	// 验证记录ID
	if _, err := SafeStringToUint64(recordID); err != nil {
		return fmt.Errorf("无效的记录ID: %v", err)
	}
	
	// 验证其他参数
	return ValidateCreateRecordParams(domain, subdomain, recordType, value, ttl)
}

// 类型安全的JSON序列化

// MarshalParams 安全地序列化参数为JSON
func MarshalParams(params map[string]interface{}) ([]byte, error) {
	// 确保所有参数类型都符合腾讯云API规范
	for key, value := range params {
		switch v := value.(type) {
		case string:
			// String类型：直接使用
		case uint64:
			// Integer类型：确保使用uint64
		case int:
			// 转换int为uint64
			if v < 0 {
				return nil, fmt.Errorf("参数%s不能为负数", key)
			}
			params[key] = uint64(v)
		case bool:
			// Boolean类型：直接使用
		case float64:
			// Double类型：直接使用
		case float32:
			// Float类型：转换为float64
			params[key] = float64(v)
		case time.Time:
			// 时间类型：转换为字符串
			params[key] = TimestampToString(v)
		default:
			return nil, fmt.Errorf("不支持的参数类型: %T for key %s", v, key)
		}
	}
	
	return json.Marshal(params)
}

// 响应类型转换

// ParseTimestamp 解析腾讯云API返回的时间戳
func ParseTimestamp(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	
	// 尝试不同的时间格式
	formats := []string{
		"2006-01-02 15:04:05",  // Timestamp格式
		"2006-01-02",           // Date格式
		time.RFC3339,           // ISO8601格式
		time.RFC3339Nano,       // ISO8601带纳秒
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("无法解析时间戳: %s", s)
}

// ConvertRecordInfo 将腾讯云API的RecordInfo转换为内部DNSRecord格式
func ConvertRecordInfo(info RecordInfo, domainName string) DNSRecord {
	return DNSRecord{
		ID:        SafeUint64ToString(info.RecordId),
		Name:      info.Name + "." + domainName,
		Subdomain: info.Name,
		Type:      info.Type,
		Value:     info.Value,
		TTL:       SafeUint64ToInt(info.TTL),
		Status:    info.Status,
	}
}

// ConvertDomainInfo 将腾讯云API的DomainInfo转换为内部格式
func ConvertDomainInfo(info DomainInfo) map[string]interface{} {
	result := map[string]interface{}{
		"domain_id": info.DomainId,  // Integer → uint64
		"name":      info.Name,      // String
		"status":    info.Status,    // String
		"ttl":       info.TTL,       // Integer → uint64
		"grade":     info.Grade,     // String
		"group_id":  info.GroupId,   // Integer → uint64
		"remark":    info.Remark,    // String
	}
	
	// 时间字段转换
	if createdOn, err := ParseTimestamp(info.CreatedOn); err == nil {
		result["created_on"] = createdOn // Timestamp → time.Time
	}
	
	if updatedOn, err := ParseTimestamp(info.UpdatedOn); err == nil {
		result["updated_on"] = updatedOn // Timestamp → time.Time
	}
	
	return result
}

// 增强的记录类型验证
func ValidateRecordTypeAndValue(recordType, value string) error {
	if err := ValidateRecordType(recordType); err != nil {
		return err
	}
	
	// 根据记录类型验证记录值格式
	switch recordType {
	case "A":
		// IPv4地址验证
		if err := ValidateIPv4(value); err != nil {
			return fmt.Errorf("A记录值格式错误: %v", err)
		}
	case "AAAA":
		// IPv6地址验证
		if err := ValidateIPv6(value); err != nil {
			return fmt.Errorf("AAAA记录值格式错误: %v", err)
		}
	case "CNAME":
		// 域名格式验证
		if err := ValidateDomainName(value); err != nil {
			return fmt.Errorf("CNAME记录值格式错误: %v", err)
		}
	case "MX":
		// MX记录格式验证 (优先级 域名)
		if err := ValidateMXRecord(value); err != nil {
			return fmt.Errorf("MX记录值格式错误: %v", err)
		}
	case "TXT":
		// TXT记录长度验证
		if len(value) > 255 {
			return fmt.Errorf("TXT记录值长度不能超过255字符")
		}
	case "NS":
		// NS记录域名验证
		if err := ValidateDomainName(value); err != nil {
			return fmt.Errorf("NS记录值格式错误: %v", err)
		}
	case "SRV":
		// SRV记录格式验证
		if err := ValidateSRVRecord(value); err != nil {
			return fmt.Errorf("SRV记录值格式错误: %v", err)
		}
	case "CAA":
		// CAA记录格式验证
		if err := ValidateCAARecord(value); err != nil {
			return fmt.Errorf("CAA记录值格式错误: %v", err)
		}
	}
	
	return nil
}

// ValidateIPv4 验证IPv4地址格式
func ValidateIPv4(ip string) error {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return fmt.Errorf("IPv4地址格式错误")
	}
	
	for _, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil || num < 0 || num > 255 {
			return fmt.Errorf("IPv4地址格式错误")
		}
	}
	
	return nil
}

// ValidateIPv6 验证IPv6地址格式
func ValidateIPv6(ip string) error {
	// 简单的IPv6格式验证
	if len(ip) < 2 || len(ip) > 39 {
		return fmt.Errorf("IPv6地址格式错误")
	}
	
	// 检查是否包含有效的IPv6字符
	validChars := "0123456789abcdefABCDEF:"
	for _, char := range ip {
		if !strings.ContainsRune(validChars, char) {
			return fmt.Errorf("IPv6地址包含无效字符")
		}
	}
	
	return nil
}

// ValidateMXRecord 验证MX记录格式
func ValidateMXRecord(value string) error {
	parts := strings.Fields(value)
	if len(parts) != 2 {
		return fmt.Errorf("MX记录格式应为: 优先级 邮件服务器域名")
	}
	
	// 验证优先级
	priority, err := strconv.Atoi(parts[0])
	if err != nil || priority < 0 || priority > 65535 {
		return fmt.Errorf("MX记录优先级必须是0-65535之间的数字")
	}
	
	// 验证邮件服务器域名
	return ValidateDomainName(parts[1])
}

// ValidateSRVRecord 验证SRV记录格式
func ValidateSRVRecord(value string) error {
	parts := strings.Fields(value)
	if len(parts) != 4 {
		return fmt.Errorf("SRV记录格式应为: 优先级 权重 端口 目标")
	}
	
	// 验证优先级
	priority, err := strconv.Atoi(parts[0])
	if err != nil || priority < 0 || priority > 65535 {
		return fmt.Errorf("SRV记录优先级必须是0-65535之间的数字")
	}
	
	// 验证权重
	weight, err := strconv.Atoi(parts[1])
	if err != nil || weight < 0 || weight > 65535 {
		return fmt.Errorf("SRV记录权重必须是0-65535之间的数字")
	}
	
	// 验证端口
	port, err := strconv.Atoi(parts[2])
	if err != nil || port < 0 || port > 65535 {
		return fmt.Errorf("SRV记录端口必须是0-65535之间的数字")
	}
	
	// 验证目标域名
	return ValidateDomainName(parts[3])
}

// ValidateCAARecord 验证CAA记录格式
func ValidateCAARecord(value string) error {
	parts := strings.Fields(value)
	if len(parts) < 3 {
		return fmt.Errorf("CAA记录格式应为: 标志 标签 值")
	}
	
	// 验证标志
	flag, err := strconv.Atoi(parts[0])
	if err != nil || flag < 0 || flag > 255 {
		return fmt.Errorf("CAA记录标志必须是0-255之间的数字")
	}
	
	// 验证标签
	tag := parts[1]
	validTags := []string{"issue", "issuewild", "iodef"}
	isValidTag := false
	for _, validTag := range validTags {
		if tag == validTag {
			isValidTag = true
			break
		}
	}
	if !isValidTag {
		return fmt.Errorf("CAA记录标签必须是issue、issuewild或iodef之一")
	}
	
	return nil
}

// 扩展的线路类型定义
var ValidRecordLines = map[string]string{
	"默认":     "默认",
	"国内":     "国内",
	"国外":     "国外",
	"电信":     "电信",
	"联通":     "联通",
	"移动":     "移动",
	"铁通":     "铁通",
	"教育网":    "教育网",
	"搜索引擎":   "搜索引擎",
	"百度":     "百度",
	"谷歌":     "谷歌",
	"必应":     "必应",
	"搜狗":     "搜狗",
	"奇虎":     "奇虎",
	"有道":     "有道",
	"搜搜":     "搜搜",
}

// ValidateRecordLine 验证解析线路
func ValidateRecordLine(line string) error {
	if line == "" {
		return nil // 空值使用默认线路
	}
	
	if _, exists := ValidRecordLines[line]; !exists {
		return fmt.Errorf("不支持的解析线路: %s", line)
	}
	
	return nil
}

// GetDefaultTTL 根据记录类型获取默认TTL值
func GetDefaultTTL(recordType string) uint64 {
	switch recordType {
	case "A", "AAAA":
		return 600  // 10分钟
	case "CNAME":
		return 600  // 10分钟
	case "MX":
		return 3600 // 1小时
	case "TXT":
		return 600  // 10分钟
	case "NS":
		return 3600 // 1小时
	case "SRV":
		return 600  // 10分钟
	case "CAA":
		return 3600 // 1小时
	default:
		return 600  // 默认10分钟
	}
}