package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

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
	Domains []struct {
		ID   json.Number `json:"id"`
		Name string      `json:"name"`
	} `json:"domains"`
	Record struct {
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

	return &DNSPodProvider{
		token:   config.Token,
		baseURL: "https://dnsapi.cn",
	}, nil
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
		"login_token":   p.token,
		"format":        "json",
		"domain_id":     domainID,
		"sub_domain":    subdomain,
		"record_type":   recordType,
		"record_line":   "默认",
		"value":         value,
		"ttl":           strconv.Itoa(ttl),
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
		"login_token":   p.token,
		"format":        "json",
		"domain_id":     domainID,
		"record_id":     recordID,
		"sub_domain":    subdomain,
		"record_type":   recordType,
		"record_line":   "默认",
		"value":         value,
		"ttl":           strconv.Itoa(ttl),
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
