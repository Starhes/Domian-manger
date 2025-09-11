package providers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ========================= DNS提供商接口定义 =========================

// DNSProvider DNS服务商接口
type DNSProvider interface {
	// CreateRecord 创建DNS记录
	// domain: 主域名 (如 example.com)
	// subdomain: 子域名 (如 www)
	// recordType: 记录类型 (如 A, CNAME, TXT)
	// value: 记录值
	// ttl: 生存时间
	// 返回: 外部记录ID, 错误
	CreateRecord(domain, subdomain, recordType, value string, ttl int) (string, error)

	// UpdateRecord 更新DNS记录
	// domain: 主域名
	// recordID: 外部记录ID
	// subdomain: 子域名
	// recordType: 记录类型
	// value: 记录值
	// ttl: 生存时间
	UpdateRecord(domain, recordID, subdomain, recordType, value string, ttl int) error

	// DeleteRecord 删除DNS记录
	// domain: 主域名
	// recordID: 外部记录ID
	DeleteRecord(domain, recordID string) error

	// GetRecords 获取域名的所有记录
	GetRecords(domain string) ([]DNSRecord, error)

	// GetDomains 获取账号下的所有域名
	GetDomains() ([]Domain, error)
}

// Domain 域名信息结构
type Domain struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// DNSRecord DNS记录结构
type DNSRecord struct {
	ID        string `json:"id"`
	Name      string `json:"name"`      // 完整域名
	Subdomain string `json:"subdomain"` // 子域名部分
	Type      string `json:"type"`
	Value     string `json:"value"`
	TTL       int    `json:"ttl"`
	Status    string `json:"status"`
}

// NewDNSProvider 创建DNS服务商实例
func NewDNSProvider(providerType, configJSON string) (DNSProvider, error) {
	switch providerType {
	case "dnspod":
		// 旧版DNSPod API (dnsapi.cn)
		return NewDNSPodProvider(configJSON)
	case "dnspod_v3":
		// 腾讯云DNSPod API v3 (tencentcloudapi.com)
		return NewDNSPodV3Provider(configJSON)
	default:
		return nil, fmt.Errorf("不支持的DNS服务商类型: %s", providerType)
	}
}

// ========================= DNSPod 传统API实现 =========================

// DNSPodProvider DNSPod服务商实现
type DNSPodProvider struct {
	token   string
	baseURL string
}

// DNSPod配置结构
type DNSPodConfig struct {
	Token string `json:"token"`
}

// DNSPod API响应结构
type DNSPodResponse struct {
	Status struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"status"`
	Domain struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"domain"`
	Domains []DNSPodDomainInfo `json:"domains"`
	Record  struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Type   string `json:"type"`
		Value  string `json:"value"`
		TTL    string `json:"ttl"`
		Status string `json:"status"`
	} `json:"record"`
	Records []struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Type   string `json:"type"`
		Value  string `json:"value"`
		TTL    string `json:"ttl"`
		Status string `json:"status"`
	} `json:"records"`
}

// DNSPodDomainInfo 传统API的域名信息结构
type DNSPodDomainInfo struct {
	ID        json.Number `json:"id"`
	Name      string      `json:"name"`
	Status    string      `json:"status"`
	GroupID   string      `json:"group_id"`
	CreatedOn string      `json:"created_on"`
	UpdatedOn string      `json:"updated_on"`
}

// NewDNSPodProvider 创建DNSPod服务商实例
func NewDNSPodProvider(configJSON string) (DNSProvider, error) {
	var config DNSPodConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("DNSPod配置解析失败: %v", err)
	}

	provider := &DNSPodProvider{
		token:   config.Token,
		baseURL: "https://dnsapi.cn",
	}

	// 验证配置有效性
	if err := provider.validateConfig(); err != nil {
		return nil, fmt.Errorf("DNSPod配置验证失败: %v", err)
	}

	return provider, nil
}

// CreateRecord 创建DNS记录
func (p *DNSPodProvider) CreateRecord(domain, subdomain, recordType, value string, ttl int) (string, error) {
	// 参数验证
	if err := p.validateCreateRecordParams(domain, subdomain, recordType, value, ttl); err != nil {
		return "", fmt.Errorf("参数验证失败: %v", err)
	}

	// 获取域名ID
	domainID, err := p.getDomainID(domain)
	if err != nil {
		return "", fmt.Errorf("获取域名ID失败: %v", err)
	}

	// 准备请求数据
	data := map[string]string{
		"login_token": p.token,
		"format":      "json",
		"domain_id":   domainID,
		"sub_domain":  subdomain,
		"record_type": recordType,
		"record_line": "默认",
		"value":       value,
		"ttl":         strconv.Itoa(ttl),
	}

	// 发送创建请求
	resp, err := p.makeRequestWithRetry("POST", "/Record.Create", data, 3)
	if err != nil {
		return "", err
	}

	if resp.Status.Code != "1" {
		return "", p.createFriendlyError(resp.Status.Code, resp.Status.Message)
	}

	return resp.Record.ID, nil
}

// UpdateRecord 更新DNS记录
func (p *DNSPodProvider) UpdateRecord(domain, recordID, subdomain, recordType, value string, ttl int) error {
	// 获取域名ID
	domainID, err := p.getDomainID(domain)
	if err != nil {
		return fmt.Errorf("获取域名ID失败: %v", err)
	}

	// 准备请求数据
	data := map[string]string{
		"login_token": p.token,
		"format":      "json",
		"domain_id":   domainID,
		"record_id":   recordID,
		"sub_domain":  subdomain,
		"record_type": recordType,
		"record_line": "默认",
		"value":       value,
		"ttl":         strconv.Itoa(ttl),
	}

	// 发送更新请求
	resp, err := p.makeRequest("POST", "/Record.Modify", data)
	if err != nil {
		return err
	}

	if resp.Status.Code != "1" {
		return fmt.Errorf("DNSPod API错误: %s", resp.Status.Message)
	}

	return nil
}

// DeleteRecord 删除DNS记录
func (p *DNSPodProvider) DeleteRecord(domain, recordID string) error {
	// 获取域名ID
	domainID, err := p.getDomainID(domain)
	if err != nil {
		return fmt.Errorf("获取域名ID失败: %v", err)
	}

	// 准备请求数据
	data := map[string]string{
		"login_token": p.token,
		"format":      "json",
		"domain_id":   domainID,
		"record_id":   recordID,
	}

	// 发送删除请求
	resp, err := p.makeRequest("POST", "/Record.Remove", data)
	if err != nil {
		return err
	}

	if resp.Status.Code != "1" {
		return fmt.Errorf("DNSPod API错误: %s", resp.Status.Message)
	}

	return nil
}

// GetRecords 获取域名的所有记录
func (p *DNSPodProvider) GetRecords(domain string) ([]DNSRecord, error) {
	// 获取域名ID
	domainID, err := p.getDomainID(domain)
	if err != nil {
		return nil, fmt.Errorf("获取域名ID失败: %v", err)
	}

	// 准备请求数据
	data := map[string]string{
		"login_token": p.token,
		"format":      "json",
		"domain_id":   domainID,
	}

	// 发送查询请求
	resp, err := p.makeRequest("POST", "/Record.List", data)
	if err != nil {
		return nil, err
	}

	if resp.Status.Code != "1" {
		return nil, fmt.Errorf("DNSPod API错误: %s", resp.Status.Message)
	}

	// 转换记录格式
	var records []DNSRecord
	for _, record := range resp.Records {
		ttl, _ := strconv.Atoi(record.TTL)
		records = append(records, DNSRecord{
			ID:        record.ID,
			Name:      record.Name + "." + domain,
			Subdomain: record.Name,
			Type:      record.Type,
			Value:     record.Value,
			TTL:       ttl,
			Status:    record.Status,
		})
	}

	return records, nil
}

// GetDomains 获取账号下所有域名
func (p *DNSPodProvider) GetDomains() ([]Domain, error) {
	data := map[string]string{
		"login_token": p.token,
		"format":      "json",
	}

	resp, err := p.makeRequest("POST", "/Domain.List", data)
	if err != nil {
		return nil, err
	}

	if resp.Status.Code != "1" {
		return nil, fmt.Errorf("DNSPod API错误: %s", resp.Status.Message)
	}

	var domains []Domain
	for _, d := range resp.Domains {
		domains = append(domains, Domain{
			ID:     d.ID.String(),
			Name:   d.Name,
			Status: d.Status,
		})
	}

	return domains, nil
}

// getDomainID 获取域名ID
func (p *DNSPodProvider) getDomainID(domain string) (string, error) {
	data := map[string]string{
		"login_token": p.token,
		"format":      "json",
		"type":        "all",
		"keyword":     domain,
	}

	resp, err := p.makeRequest("POST", "/Domain.List", data)
	if err != nil {
		return "", err
	}

	if resp.Status.Code != "1" {
		return "", fmt.Errorf("DNSPod API错误: %s", resp.Status.Message)
	}

	for _, d := range resp.Domains {
		if d.Name == domain {
			return d.ID.String(), nil
		}
	}

	return "", fmt.Errorf("未找到域名 '%s'", domain)
}

// makeRequest 发送HTTP请求
func (p *DNSPodProvider) makeRequest(method, endpoint string, data map[string]string) (*DNSPodResponse, error) {
	// 构建请求URL
	urlStr := p.baseURL + endpoint

	// 构建请求体
	var reqBody io.Reader
	if method == "POST" {
		formData := url.Values{}
		for key, val := range data {
			formData.Set(key, val)
		}
		reqBody = bytes.NewBufferString(formData.Encode())
	}

	// 创建HTTP请求
	req, err := http.NewRequest(method, urlStr, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Domain-Manager/1.0")

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求发送失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("响应读取失败: %v", err)
	}

	// 解析响应
	var dnspodResp DNSPodResponse
	if err := json.Unmarshal(body, &dnspodResp); err != nil {
		return nil, fmt.Errorf("响应解析失败: %v", err)
	}

	return &dnspodResp, nil
}

// validateCreateRecordParams 验证创建记录参数（传统API版本）
func (p *DNSPodProvider) validateCreateRecordParams(domain, subdomain, recordType, value string, ttl int) error {
	// 基础参数验证
	if domain == "" {
		return fmt.Errorf("域名不能为空")
	}
	if subdomain == "" {
		return fmt.Errorf("子域名不能为空")
	}
	if recordType == "" {
		return fmt.Errorf("记录类型不能为空")
	}
	if value == "" {
		return fmt.Errorf("记录值不能为空")
	}

	// 验证记录类型
	validTypes := []string{"A", "AAAA", "CNAME", "TXT", "MX", "NS", "SRV", "CAA"}
	isValidType := false
	for _, validType := range validTypes {
		if recordType == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return fmt.Errorf("不支持的记录类型: %s", recordType)
	}

	// 验证TTL范围
	if ttl < 1 || ttl > 604800 {
		return fmt.Errorf("TTL值必须在1-604800秒之间")
	}

	return nil
}

// makeRequestWithRetry 带重试机制的请求方法
func (p *DNSPodProvider) makeRequestWithRetry(method, endpoint string, data map[string]string, maxRetries int) (*DNSPodResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// 指数退避
			time.Sleep(time.Duration(attempt*attempt) * time.Second)
		}

		resp, err := p.makeRequest(method, endpoint, data)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// 检查是否可以重试
		if !p.isRetryableError(err) {
			break
		}
	}

	return nil, fmt.Errorf("请求失败，已重试%d次: %v", maxRetries, lastErr)
}

// isRetryableError 判断错误是否可以重试
func (p *DNSPodProvider) isRetryableError(err error) bool {
	errStr := err.Error()
	// 网络错误或服务器错误可以重试
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "server error")
}

// validateConfig 验证DNSPod配置
func (p *DNSPodProvider) validateConfig() error {
	// 基础配置检查
	if p.token == "" {
		return fmt.Errorf("DNSPod Token不能为空")
	}

	// Token格式检查
	if !strings.Contains(p.token, ",") {
		return fmt.Errorf("DNSPod Token格式不正确，应为：ID,Token")
	}

	parts := strings.Split(p.token, ",")
	if len(parts) != 2 {
		return fmt.Errorf("DNSPod Token格式不正确，应为：ID,Token")
	}

	if parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("DNSPod Token的ID或Token部分不能为空")
	}

	// 尝试连接测试
	return p.testConnection()
}

// testConnection 测试DNSPod连接
func (p *DNSPodProvider) testConnection() error {
	// 尝试获取域名列表来验证token有效性
	data := map[string]string{
		"login_token": p.token,
		"format":      "json",
		"type":        "all",
		"length":      "1", // 只获取1个域名用于测试
	}

	resp, err := p.makeRequest("POST", "/Domain.List", data)
	if err != nil {
		return fmt.Errorf("连接DNSPod API失败: %v", err)
	}

	// 检查API响应
	switch resp.Status.Code {
	case "1":
		return nil // 连接成功
	case "-1":
		return fmt.Errorf("DNSPod Token无效或已过期")
	case "-7":
		return fmt.Errorf("DNSPod账户已被禁用")
	case "-8":
		return fmt.Errorf("DNSPod账户暂时被禁用（登录失败次数过多）")
	default:
		return fmt.Errorf("DNSPod连接测试失败: %s", resp.Status.Message)
	}
}

// createFriendlyError 创建友好的错误信息
func (p *DNSPodProvider) createFriendlyError(code, message string) error {
	var friendlyMessage string

	switch code {
	case "-1":
		friendlyMessage = "登录失败，请检查Token是否正确"
	case "-2":
		friendlyMessage = "API使用超出限制，请稍后重试"
	case "-3":
		friendlyMessage = "不是域名所有者或没有权限"
	case "-4":
		friendlyMessage = "记录不存在"
	case "-7":
		friendlyMessage = "您的账户已被禁用"
	case "-8":
		friendlyMessage = "登录失败次数过多，账户被暂时禁用"
	case "6":
		friendlyMessage = "域名ID错误"
	case "7":
		friendlyMessage = "域名被锁定"
	case "21":
		friendlyMessage = "域名不存在或不属于当前账户"
	case "22":
		friendlyMessage = "子域名不合法"
	case "23":
		friendlyMessage = "记录类型不正确"
	case "24":
		friendlyMessage = "记录线路不正确"
	case "25":
		friendlyMessage = "记录值不正确"
	case "26":
		friendlyMessage = "记录权重不正确"
	case "27":
		friendlyMessage = "记录的TTL值不正确"
	default:
		friendlyMessage = message
	}

	return fmt.Errorf("DNSPod API错误 [%s]: %s", code, friendlyMessage)
}

// ========================= DNSPod V3 API实现 =========================

// DNSPodV3Provider 腾讯云DNSPod API v3实现
type DNSPodV3Provider struct {
	secretId  string
	secretKey string
	region    string
	baseURL   string
}

// DNSPodV3配置结构
type DNSPodV3Config struct {
	SecretId  string `json:"secret_id"`
	SecretKey string `json:"secret_key"`
	Region    string `json:"region"` // 可选，默认为空（就近接入）
}

// 腾讯云API通用响应结构 - 严格按照官方规范
type TencentCloudResponse struct {
	Response struct {
		Error *struct {
			Code    string `json:"Code"`    // 错误码，用于标识具体错误类型
			Message string `json:"Message"` // 错误描述，可能会变更，不应依赖此值
		} `json:"Error,omitempty"` // Error字段存在表示请求失败
		RequestId string `json:"RequestId"` // 请求唯一标识，用于问题排查
	} `json:"Response"`
}

// 腾讯云API错误信息结构
type TencentCloudError struct {
	Code      string `json:"Code"`
	Message   string `json:"Message"`
	RequestId string `json:"RequestId"`
}

// 域名列表响应
type DescribeDomainsResponse struct {
	Response struct {
		Error *struct {
			Code    string `json:"Code"`
			Message string `json:"Message"`
		} `json:"Error,omitempty"`
		RequestId string `json:"RequestId"`

		// 业务数据字段
		DomainList []struct {
			DomainId uint64 `json:"DomainId"`
			Name     string `json:"Name"`
			Status   string `json:"Status"`
		} `json:"DomainList,omitempty"`
		DomainCountInfo struct {
			DomainTotal uint64 `json:"DomainTotal"`
		} `json:"DomainCountInfo,omitempty"`
	} `json:"Response"`
}

// 记录列表响应
type DescribeRecordListResponse struct {
	Response struct {
		Error *struct {
			Code    string `json:"Code"`
			Message string `json:"Message"`
		} `json:"Error,omitempty"`
		RequestId string `json:"RequestId"`

		// 业务数据字段
		RecordCountInfo struct {
			TotalCount uint64 `json:"TotalCount"`
		} `json:"RecordCountInfo,omitempty"`
		RecordList []struct {
			RecordId uint64 `json:"RecordId"`
			Name     string `json:"Name"`
			Type     string `json:"Type"`
			Value    string `json:"Value"`
			TTL      uint64 `json:"TTL"`
			Status   string `json:"Status"`
			Line     string `json:"Line"`
		} `json:"RecordList,omitempty"`
	} `json:"Response"`
}

// 创建记录响应
type CreateRecordResponse struct {
	Response struct {
		Error *struct {
			Code    string `json:"Code"`
			Message string `json:"Message"`
		} `json:"Error,omitempty"`
		RequestId string `json:"RequestId"`

		// 业务数据字段
		RecordId uint64 `json:"RecordId,omitempty"`
	} `json:"Response"`
}

// NewDNSPodV3Provider 创建腾讯云DNSPod服务商实例
func NewDNSPodV3Provider(configJSON string) (DNSProvider, error) {
	var config DNSPodV3Config
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("腾讯云DNSPod配置解析失败: %v", err)
	}

	// 验证配置参数
	if err := validateDNSPodV3Config(config); err != nil {
		return nil, err
	}

	// 默认使用就近接入
	baseURL := "https://dnspod.tencentcloudapi.com"
	if config.Region != "" {
		// 如果指定了地域，使用地域专用域名
		baseURL = fmt.Sprintf("https://dnspod.%s.tencentcloudapi.com", config.Region)
	}

	return &DNSPodV3Provider{
		secretId:  config.SecretId,
		secretKey: config.SecretKey,
		region:    config.Region,
		baseURL:   baseURL,
	}, nil
}

// validateDNSPodV3Config 验证腾讯云DNSPod配置
func validateDNSPodV3Config(config DNSPodV3Config) error {
	if config.SecretId == "" {
		return fmt.Errorf("SecretId不能为空")
	}

	if config.SecretKey == "" {
		return fmt.Errorf("SecretKey不能为空")
	}

	// 验证SecretId格式（通常以AKID开头）
	if !strings.HasPrefix(config.SecretId, "AKID") {
		return fmt.Errorf("SecretId格式不正确，应以AKID开头")
	}

	// 验证SecretId长度（通常为36位）
	if len(config.SecretId) != 36 {
		return fmt.Errorf("SecretId长度不正确，应为36位")
	}

	// 验证SecretKey长度（通常为32位）
	if len(config.SecretKey) != 32 {
		return fmt.Errorf("SecretKey长度不正确，应为32位")
	}

	// 验证地域格式（如果指定）
	if config.Region != "" {
		validRegions := []string{
			"ap-guangzhou", "ap-shanghai", "ap-nanjing", "ap-beijing",
			"ap-chengdu", "ap-chongqing", "ap-hongkong", "ap-singapore",
			"ap-jakarta", "ap-bangkok", "ap-seoul", "ap-tokyo",
			"na-ashburn", "na-siliconvalley", "sa-saopaulo", "eu-frankfurt",
		}

		isValidRegion := false
		for _, validRegion := range validRegions {
			if config.Region == validRegion {
				isValidRegion = true
				break
			}
		}

		if !isValidRegion {
			return fmt.Errorf("不支持的地域: %s", config.Region)
		}
	}

	return nil
}

// CreateRecord 创建DNS记录
func (p *DNSPodV3Provider) CreateRecord(domain, subdomain, recordType, value string, ttl int) (string, error) {
	// 参数验证 - 按照腾讯云API参数类型规范
	if err := validateCreateRecordParams(domain, subdomain, recordType, value, ttl); err != nil {
		return "", fmt.Errorf("参数验证失败: %v", err)
	}

	// 构建类型安全的请求参数
	params := buildCreateRecordParams(domain, subdomain, recordType, value, ttl)

	// 发送创建请求
	var resp CreateRecordResponse
	if err := p.makeRequest("CreateRecord", params, &resp); err != nil {
		return "", err
	}

	return strconv.FormatUint(resp.Response.RecordId, 10), nil
}

// UpdateRecord 更新DNS记录
func (p *DNSPodV3Provider) UpdateRecord(domain, recordID, subdomain, recordType, value string, ttl int) error {
	// 参数验证 - 按照腾讯云API参数类型规范
	if err := validateModifyRecordParams(domain, recordID, subdomain, recordType, value, ttl); err != nil {
		return fmt.Errorf("参数验证失败: %v", err)
	}

	// 构建类型安全的请求参数
	params, err := buildModifyRecordParams(domain, recordID, subdomain, recordType, value, ttl)
	if err != nil {
		return fmt.Errorf("构建请求参数失败: %v", err)
	}

	// 发送更新请求
	var resp TencentCloudResponse
	return p.makeRequest("ModifyRecord", params, &resp)
}

// DeleteRecord 删除DNS记录
func (p *DNSPodV3Provider) DeleteRecord(domain, recordID string) error {
	// 参数验证
	if err := validateDomainName(domain); err != nil {
		return fmt.Errorf("域名验证失败: %v", err)
	}

	// 构建类型安全的请求参数
	params, err := buildDeleteRecordParams(domain, recordID)
	if err != nil {
		return fmt.Errorf("构建请求参数失败: %v", err)
	}

	// 发送删除请求
	var resp TencentCloudResponse
	return p.makeRequest("DeleteRecord", params, &resp)
}

// GetRecords 获取域名的所有记录
func (p *DNSPodV3Provider) GetRecords(domain string) ([]DNSRecord, error) {
	// 参数验证
	if err := validateDomainName(domain); err != nil {
		return nil, fmt.Errorf("域名验证失败: %v", err)
	}

	// 构建类型安全的请求参数
	params := buildDescribeRecordListParams(domain)

	// 发送查询请求
	var resp DescribeRecordListResponse
	if err := p.makeRequest("DescribeRecordList", params, &resp); err != nil {
		return nil, err
	}

	// 转换记录格式
	var records []DNSRecord
	for _, record := range resp.Response.RecordList {
		records = append(records, DNSRecord{
			ID:        strconv.FormatUint(record.RecordId, 10),
			Name:      record.Name + "." + domain,
			Subdomain: record.Name,
			Type:      record.Type,
			Value:     record.Value,
			TTL:       int(record.TTL),
			Status:    record.Status,
		})
	}

	return records, nil
}

// GetDomains 获取账号下所有域名
func (p *DNSPodV3Provider) GetDomains() ([]Domain, error) {
	params := buildDescribeDomainListParams("") // 传入空字符串以获取所有域名

	var resp DescribeDomainsResponse
	if err := p.makeRequest("DescribeDomainList", params, &resp); err != nil {
		return nil, err
	}

	var domains []Domain
	for _, d := range resp.Response.DomainList {
		domains = append(domains, Domain{
			ID:     strconv.FormatUint(d.DomainId, 10),
			Name:   d.Name,
			Status: d.Status,
		})
	}

	return domains, nil
}

// makeRequest 发送腾讯云API请求
// 严格按照腾讯云API规范实现，包含重试机制
func (p *DNSPodV3Provider) makeRequest(action string, params map[string]interface{}, result interface{}) error {
	return p.makeRequestWithRetryV3(action, params, result, 3) // 默认重试3次
}

// makeRequestWithRetryV3 带重试机制的API请求
func (p *DNSPodV3Provider) makeRequestWithRetryV3(action string, params map[string]interface{}, result interface{}, maxRetries int) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// 记录重试信息
		if attempt > 0 {
			backoffDuration := time.Duration(attempt*attempt) * time.Second
			time.Sleep(backoffDuration) // 指数退避
		}

		// 执行单次请求
		_, err := p.doSingleRequest(action, params, result)

		if err == nil {
			return nil // 成功
		}

		lastErr = err

		// 检查是否可以重试（通过错误消息中的错误码判断）
		if strings.Contains(err.Error(), "腾讯云DNSPod API错误") {
			// 提取错误码进行重试判断
			isRetryable := false
			if strings.Contains(err.Error(), "[InternalError]") ||
				strings.Contains(err.Error(), "[RequestLimitExceeded]") ||
				strings.Contains(err.Error(), "[ResourceUnavailable]") {
				isRetryable = true
			}

			if !isRetryable {
				break // 不可重试的错误，直接返回
			}
		} else {
			break // 非腾讯云API错误，直接返回
		}
	}

	return fmt.Errorf("请求失败，已重试%d次: %v", maxRetries, lastErr)
}

// doSingleRequest 执行单次API请求
func (p *DNSPodV3Provider) doSingleRequest(action string, params map[string]interface{}, result interface{}) (string, error) {
	// 参数验证
	if action == "" {
		return "", fmt.Errorf("Action参数不能为空")
	}

	// 准备请求体 - 使用类型安全的序列化
	jsonParams, err := marshalParams(params)
	if err != nil {
		return "", fmt.Errorf("参数序列化失败: %v", err)
	}

	// 创建HTTP请求 - 使用POST方法和JSON格式
	req, err := http.NewRequest("POST", p.baseURL, bytes.NewBuffer(jsonParams))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头 - 按照公共参数规范
	timestamp := time.Now().Unix()

	// 验证时间戳
	if err := validateTimestamp(timestamp); err != nil {
		return "", fmt.Errorf("时间戳验证失败: %v", err)
	}

	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")
	host := strings.TrimPrefix(p.baseURL, "https://")

	// 必选公共参数（通过HTTP Header传递）
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Host", host)
	req.Header.Set("X-TC-Action", action)                              // Action - 操作的接口名称
	req.Header.Set("X-TC-Version", "2021-03-23")                       // Version - API版本
	req.Header.Set("X-TC-Timestamp", strconv.FormatInt(timestamp, 10)) // Timestamp - UNIX时间戳

	// 可选公共参数
	if p.region != "" {
		req.Header.Set("X-TC-Region", p.region) // Region - 地域参数
	}

	// 设置语言为中文（可选）
	req.Header.Set("X-TC-Language", "zh-CN")

	// 计算签名
	signature, err := p.calculateSignature(req, string(jsonParams), timestamp, date)
	if err != nil {
		return "", fmt.Errorf("签名计算失败: %v", err)
	}

	req.Header.Set("Authorization", signature)

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求发送失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("响应读取失败: %v", err)
	}

	// 解析响应
	if err := json.Unmarshal(body, result); err != nil {
		return "", fmt.Errorf("响应解析失败: %v", err)
	}

	// 获取RequestId用于日志记录
	var errorCheck TencentCloudResponse
	requestId := ""
	if json.Unmarshal(body, &errorCheck) == nil {
		requestId = errorCheck.Response.RequestId
	}

	// 检查是否有错误 - 根据腾讯云API返回结果规范处理
	if err := p.checkAPIError(body, result); err != nil {
		return requestId, err
	}

	return requestId, nil
}

// calculateSignature 计算TC3-HMAC-SHA256签名
// 严格按照腾讯云API公共参数规范实现
func (p *DNSPodV3Provider) calculateSignature(req *http.Request, payload string, timestamp int64, date string) (string, error) {
	// 步骤1：拼接规范请求串
	httpRequestMethod := req.Method
	canonicalURI := "/"
	canonicalQueryString := ""

	// 拼接规范头部 - 必须包含content-type和host，可选x-tc-action，按字母序排列
	contentType := req.Header.Get("Content-Type")
	host := req.Header.Get("Host")
	action := strings.ToLower(req.Header.Get("X-TC-Action"))
	canonicalHeaders := fmt.Sprintf("content-type:%s\nhost:%s\nx-tc-action:%s\n", contentType, host, action)
	signedHeaders := "content-type;host;x-tc-action"

	// 计算payload哈希
	hashedPayload := sha256Hash(payload)

	// 拼接规范请求串
	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		httpRequestMethod, canonicalURI, canonicalQueryString,
		canonicalHeaders, signedHeaders, hashedPayload)

	// 步骤2：拼接待签名字符串
	algorithm := "TC3-HMAC-SHA256"
	service := "dnspod" // 产品名，对应dnspod.tencentcloudapi.com
	credentialScope := fmt.Sprintf("%s/%s/tc3_request", date, service)
	hashedCanonicalRequest := sha256Hash(canonicalRequest)
	stringToSign := fmt.Sprintf("%s\n%d\n%s\n%s",
		algorithm, timestamp, credentialScope, hashedCanonicalRequest)

	// 步骤3：计算签名 - 使用派生签名密钥
	secretDate := hmacSha256([]byte("TC3"+p.secretKey), date)
	secretService := hmacSha256(secretDate, service)
	secretSigning := hmacSha256(secretService, "tc3_request")
	signature := hex.EncodeToString(hmacSha256(secretSigning, stringToSign))

	// 步骤4：拼接Authorization头 - 严格按照规范格式
	authorization := fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		algorithm, p.secretId, credentialScope, signedHeaders, signature)

	return authorization, nil
}

// sha256Hash 计算SHA256哈希
func sha256Hash(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// hmacSha256 计算HMAC-SHA256
func hmacSha256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

// validateTimestamp 验证时间戳是否在有效范围内（5分钟内）
func validateTimestamp(timestamp int64) error {
	now := time.Now().Unix()
	diff := now - timestamp
	if diff < 0 {
		diff = -diff
	}

	// 时间戳与服务器时间相差不能超过5分钟
	if diff > 300 {
		return fmt.Errorf("时间戳无效，与服务器时间相差超过5分钟")
	}

	return nil
}

// checkAPIError 检查腾讯云API返回的错误信息
// 严格按照腾讯云API返回结果规范处理错误
func (p *DNSPodV3Provider) checkAPIError(body []byte, result interface{}) error {
	// 解析通用错误结构
	var errorCheck TencentCloudResponse
	if err := json.Unmarshal(body, &errorCheck); err != nil {
		return fmt.Errorf("响应解析失败: %v", err)
	}

	// 检查是否有错误字段
	if errorCheck.Response.Error != nil {
		return p.createAPIError(errorCheck.Response.Error.Code,
			errorCheck.Response.Error.Message,
			errorCheck.Response.RequestId)
	}

	return nil
}

// createAPIError 创建标准化的API错误
func (p *DNSPodV3Provider) createAPIError(code, message, requestId string) error {
	// 根据错误码提供更友好的错误信息
	var friendlyMessage string

	switch code {
	case "AuthFailure.SignatureExpire":
		friendlyMessage = "签名已过期，请检查系统时间是否同步（误差不能超过5分钟）"
	case "AuthFailure.SignatureFailure":
		friendlyMessage = "签名验证失败，请检查SecretKey是否正确，或请求内容是否被篡改"
	case "AuthFailure.SecretIdNotFound":
		friendlyMessage = "SecretId不存在或已被禁用，请检查控制台中的密钥状态"
	case "AuthFailure.InvalidSecretId":
		friendlyMessage = "SecretId格式无效，请确认使用的是云API密钥"
	case "AuthFailure.TokenFailure":
		friendlyMessage = "临时凭证Token无效或已过期"
	case "InvalidParameter":
		friendlyMessage = "请求参数错误，请检查参数格式和取值范围"
	case "ResourceNotFound":
		friendlyMessage = "请求的资源不存在"
	case "ResourceUnavailable":
		friendlyMessage = "资源不可用"
	case "UnauthorizedOperation":
		friendlyMessage = "未授权的操作，请检查账户权限"
	case "RequestLimitExceeded":
		friendlyMessage = "请求频率超过限制，请稍后重试"
	case "InternalError":
		friendlyMessage = "内部错误，请稍后重试或联系技术支持"
	default:
		friendlyMessage = message
	}

	// 返回结构化错误信息
	return fmt.Errorf("腾讯云DNSPod API错误 [%s]: %s (RequestId: %s)",
		code, friendlyMessage, requestId)
}

// ========================= 辅助函数 =========================

// validateCreateRecordParams 验证创建记录参数
func validateCreateRecordParams(domain, subdomain, recordType, value string, ttl int) error {
	// 验证域名
	if err := validateDomainName(domain); err != nil {
		return err
	}

	// 验证子域名
	if err := validateSubdomain(subdomain); err != nil {
		return err
	}

	// 验证记录类型
	if err := validateRecordType(recordType); err != nil {
		return err
	}

	// 验证TTL
	if ttl > 0 {
		if err := validateTTL(uint64(ttl)); err != nil {
			return err
		}
	}

	return nil
}

// validateModifyRecordParams 验证修改记录参数
func validateModifyRecordParams(domain, recordID, subdomain, recordType, value string, ttl int) error {
	// 验证记录ID
	if _, err := strconv.ParseUint(recordID, 10, 64); err != nil {
		return fmt.Errorf("无效的记录ID: %v", err)
	}

	// 验证其他参数
	return validateCreateRecordParams(domain, subdomain, recordType, value, ttl)
}

// validateDomainName 验证域名格式
func validateDomainName(domain string) error {
	if len(domain) == 0 {
		return fmt.Errorf("域名不能为空")
	}
	if len(domain) > 253 {
		return fmt.Errorf("域名长度不能超过253个字符")
	}
	return nil
}

// validateSubdomain 验证子域名格式
func validateSubdomain(subdomain string) error {
	if len(subdomain) == 0 {
		return fmt.Errorf("子域名不能为空")
	}
	if len(subdomain) > 63 {
		return fmt.Errorf("子域名长度不能超过63个字符")
	}
	return nil
}

// validateRecordType 验证DNS记录类型
func validateRecordType(recordType string) error {
	validTypes := []string{"A", "AAAA", "CNAME", "TXT", "MX", "NS", "SRV", "CAA"}
	for _, validType := range validTypes {
		if recordType == validType {
			return nil
		}
	}
	return fmt.Errorf("不支持的记录类型: %s", recordType)
}

// validateTTL 验证TTL值
func validateTTL(ttl uint64) error {
	// TTL范围：1-604800秒（7天）
	if ttl < 1 || ttl > 604800 {
		return fmt.Errorf("TTL值必须在1-604800秒之间")
	}
	return nil
}

// buildCreateRecordParams 构建创建记录的参数
func buildCreateRecordParams(domain, subdomain, recordType, value string, ttl int) map[string]interface{} {
	params := map[string]interface{}{
		"Domain":     domain,
		"SubDomain":  subdomain,
		"RecordType": recordType,
		"Value":      value,
	}

	// 可选参数
	if ttl > 0 {
		params["TTL"] = uint64(ttl)
	}

	// 设置默认线路
	params["RecordLine"] = "默认"

	return params
}

// buildModifyRecordParams 构建修改记录的参数
func buildModifyRecordParams(domain, recordID, subdomain, recordType, value string, ttl int) (map[string]interface{}, error) {
	// 将recordID从string转换为uint64
	recordId, err := strconv.ParseUint(recordID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("无效的记录ID: %v", err)
	}

	params := map[string]interface{}{
		"Domain":     domain,
		"RecordId":   recordId,
		"SubDomain":  subdomain,
		"RecordType": recordType,
		"Value":      value,
	}

	// 可选参数
	if ttl > 0 {
		params["TTL"] = uint64(ttl)
	}

	// 设置默认线路
	params["RecordLine"] = "默认"

	return params, nil
}

// buildDeleteRecordParams 构建删除记录的参数
func buildDeleteRecordParams(domain, recordID string) (map[string]interface{}, error) {
	// 将recordID从string转换为uint64
	recordId, err := strconv.ParseUint(recordID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("无效的记录ID: %v", err)
	}

	params := map[string]interface{}{
		"Domain":   domain,
		"RecordId": recordId,
	}

	return params, nil
}

// buildDescribeRecordListParams 构建查询记录列表的参数
func buildDescribeRecordListParams(domain string) map[string]interface{} {
	params := map[string]interface{}{
		"Domain": domain,
		"Limit":  uint64(100),
		"Offset": uint64(0),
	}

	return params
}

// buildDescribeDomainListParams 构建查询域名列表的参数
func buildDescribeDomainListParams(keyword string) map[string]interface{} {
	params := map[string]interface{}{
		"Limit":  uint64(20),
		"Offset": uint64(0),
	}

	// 可选参数
	if keyword != "" {
		params["Keyword"] = keyword
	}

	return params
}

// marshalParams 安全地序列化参数为JSON
func marshalParams(params map[string]interface{}) ([]byte, error) {
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
			params[key] = v.Format("2006-01-02 15:04:05")
		default:
			return nil, fmt.Errorf("不支持的参数类型: %T for key %s", v, key)
		}
	}

	return json.Marshal(params)
}
