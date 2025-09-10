package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
	resp, err := p.makeRequest("POST", "/Record.Create", data)
	if err != nil {
		return "", err
	}

	if resp.Status.Code != "1" {
		return "", fmt.Errorf("DNSPod API错误: %s", resp.Status.Message)
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

	return resp.Domain.ID, nil
}

// makeRequest 发送HTTP请求
func (p *DNSPodProvider) makeRequest(method, endpoint string, data map[string]string) (*DNSPodResponse, error) {
	// 构建请求URL
	url := p.baseURL + endpoint

	// 构建请求体
	var reqBody io.Reader
	if method == "POST" {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("请求数据编码失败: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	// 创建HTTP请求
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
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
