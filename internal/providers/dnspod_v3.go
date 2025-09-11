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
	"strconv"
	"strings"
	"time"
)

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
	if err := ValidateCreateRecordParams(domain, subdomain, recordType, value, ttl); err != nil {
		return "", fmt.Errorf("参数验证失败: %v", err)
	}

	// 构建类型安全的请求参数
	params := BuildCreateRecordParams(domain, subdomain, recordType, value, ttl)

	// 发送创建请求
	var resp CreateRecordResponse
	if err := p.makeRequest("CreateRecord", params, &resp); err != nil {
		return "", err
	}

	return SafeUint64ToString(resp.Response.RecordId), nil
}

// UpdateRecord 更新DNS记录
func (p *DNSPodV3Provider) UpdateRecord(domain, recordID, subdomain, recordType, value string, ttl int) error {
	// 参数验证 - 按照腾讯云API参数类型规范
	if err := ValidateModifyRecordParams(domain, recordID, subdomain, recordType, value, ttl); err != nil {
		return fmt.Errorf("参数验证失败: %v", err)
	}

	// 构建类型安全的请求参数
	params, err := BuildModifyRecordParams(domain, recordID, subdomain, recordType, value, ttl)
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
	if err := ValidateDomainName(domain); err != nil {
		return fmt.Errorf("域名验证失败: %v", err)
	}

	// 构建类型安全的请求参数
	params, err := BuildDeleteRecordParams(domain, recordID)
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
	if err := ValidateDomainName(domain); err != nil {
		return nil, fmt.Errorf("域名验证失败: %v", err)
	}

	// 构建类型安全的请求参数
	params := BuildDescribeRecordListParams(domain)

	// 发送查询请求
	var resp DescribeRecordListResponse
	if err := p.makeRequest("DescribeRecordList", params, &resp); err != nil {
		return nil, err
	}

	// 转换记录格式 - 使用类型安全的转换函数
	var records []DNSRecord
	for _, record := range resp.Response.RecordList {
		convertedRecord := ConvertRecordInfo(RecordInfo{
			RecordId: record.RecordId,
			Name:     record.Name,
			Type:     record.Type,
			Value:    record.Value,
			TTL:      record.TTL,
			Status:   record.Status,
			Line:     record.Line,
		}, domain)
		records = append(records, convertedRecord)
	}

	return records, nil
}

// GetDomains 获取账号下所有域名
func (p *DNSPodV3Provider) GetDomains() ([]Domain, error) {
	params := BuildDescribeDomainListParams("") // 传入空字符串以获取所有域名

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
	return p.makeRequestWithRetry(action, params, result, 3) // 默认重试3次
}

// makeRequestWithRetry 带重试机制的API请求
func (p *DNSPodV3Provider) makeRequestWithRetry(action string, params map[string]interface{}, result interface{}, maxRetries int) error {
	var lastErr error
	startTime := time.Now()

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// 记录重试信息
		if attempt > 0 {
			backoffDuration := time.Duration(attempt*attempt) * time.Second
			time.Sleep(backoffDuration) // 指数退避
		}

		// 执行单次请求
		requestId, err := p.doSingleRequest(action, params, result)

		// 记录API调用
		p.logAPICall(action, err == nil, requestId, time.Since(startTime))

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
	jsonParams, err := MarshalParams(params)
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

// 添加一些调试和错误处理的辅助函数

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

// logSignatureDebug 记录签名调试信息（仅在开发环境）
func (p *DNSPodV3Provider) logSignatureDebug(canonicalRequest, stringToSign, signature string) {
	// 在生产环境中，这些信息不应该被记录
	// 这里仅用于开发调试
	if false { // 可以通过环境变量控制
		fmt.Printf("=== DNSPod V3 签名调试信息 ===\n")
		fmt.Printf("规范请求串:\n%s\n", canonicalRequest)
		fmt.Printf("待签名字符串:\n%s\n", stringToSign)
		fmt.Printf("最终签名: %s\n", signature)
		fmt.Printf("===============================\n")
	}
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

// isRetryableError 判断错误是否可以重试
func (p *DNSPodV3Provider) isRetryableError(code string) bool {
	retryableCodes := []string{
		"InternalError",        // 内部错误
		"RequestLimitExceeded", // 请求频率限制
		"ResourceUnavailable",  // 资源不可用
	}

	for _, retryableCode := range retryableCodes {
		if code == retryableCode {
			return true
		}
	}

	return false
}

// logAPICall 记录API调用信息（用于审计和调试）
func (p *DNSPodV3Provider) logAPICall(action string, success bool, requestId string, duration time.Duration) {
	status := "SUCCESS"
	if !success {
		status = "FAILED"
	}

	// 在生产环境中，这些信息应该记录到日志系统
	// 这里仅作为示例
	if false { // 可以通过环境变量控制
		fmt.Printf("[DNSPod API] %s %s - RequestId: %s, Duration: %v\n",
			action, status, requestId, duration)
	}
}

// GetRecordByID 根据记录ID获取单个DNS记录
func (p *DNSPodV3Provider) GetRecordByID(domain, recordID string) (*DNSRecord, error) {
	// 参数验证
	if err := ValidateDomainName(domain); err != nil {
		return nil, fmt.Errorf("域名验证失败: %v", err)
	}

	// 将recordID从string转换为uint64
	recordId, err := SafeStringToUint64(recordID)
	if err != nil {
		return nil, fmt.Errorf("无效的记录ID: %v", err)
	}

	// 构建请求参数
	params := map[string]interface{}{
		"Domain":   domain,
		"RecordId": recordId,
	}

	// 发送查询请求
	var resp struct {
		Response struct {
			Error *struct {
				Code    string `json:"Code"`
				Message string `json:"Message"`
			} `json:"Error,omitempty"`
			RequestId string `json:"RequestId"`

			RecordInfo struct {
				RecordId uint64 `json:"RecordId"`
				Name     string `json:"Name"`
				Type     string `json:"Type"`
				Value    string `json:"Value"`
				TTL      uint64 `json:"TTL"`
				Status   string `json:"Status"`
				Line     string `json:"Line"`
			} `json:"RecordInfo,omitempty"`
		} `json:"Response"`
	}

	if err := p.makeRequest("DescribeRecord", params, &resp); err != nil {
		return nil, err
	}

	// 转换记录格式
	record := ConvertRecordInfo(RecordInfo{
		RecordId: resp.Response.RecordInfo.RecordId,
		Name:     resp.Response.RecordInfo.Name,
		Type:     resp.Response.RecordInfo.Type,
		Value:    resp.Response.RecordInfo.Value,
		TTL:      resp.Response.RecordInfo.TTL,
		Status:   resp.Response.RecordInfo.Status,
		Line:     resp.Response.RecordInfo.Line,
	}, domain)

	return &record, nil
}

// SetRecordStatus 设置记录状态（启用/禁用）
func (p *DNSPodV3Provider) SetRecordStatus(domain, recordID string, status string) error {
	// 参数验证
	if err := ValidateDomainName(domain); err != nil {
		return fmt.Errorf("域名验证失败: %v", err)
	}

	// 验证状态参数
	if status != "ENABLE" && status != "DISABLE" {
		return fmt.Errorf("无效的记录状态: %s，应为ENABLE或DISABLE", status)
	}

	// 将recordID从string转换为uint64
	recordId, err := SafeStringToUint64(recordID)
	if err != nil {
		return fmt.Errorf("无效的记录ID: %v", err)
	}

	// 构建请求参数
	params := map[string]interface{}{
		"Domain":   domain,
		"RecordId": recordId,
		"Status":   status,
	}

	// 发送设置状态请求
	var resp TencentCloudResponse
	return p.makeRequest("ModifyRecordStatus", params, &resp)
}

// GetDomainInfo 获取域名详细信息
func (p *DNSPodV3Provider) GetDomainInfo(domain string) (map[string]interface{}, error) {
	// 参数验证
	if err := ValidateDomainName(domain); err != nil {
		return nil, fmt.Errorf("域名验证失败: %v", err)
	}

	// 构建请求参数
	params := map[string]interface{}{
		"Domain": domain,
	}

	// 发送查询请求
	var resp struct {
		Response struct {
			Error *struct {
				Code    string `json:"Code"`
				Message string `json:"Message"`
			} `json:"Error,omitempty"`
			RequestId string `json:"RequestId"`

			DomainInfo DomainInfo `json:"DomainInfo,omitempty"`
		} `json:"Response"`
	}

	if err := p.makeRequest("DescribeDomain", params, &resp); err != nil {
		return nil, err
	}

	// 转换域名信息格式
	return ConvertDomainInfo(resp.Response.DomainInfo), nil
}

// BatchCreateRecords 批量创建DNS记录
func (p *DNSPodV3Provider) BatchCreateRecords(domain string, records []CreateRecordRequest) ([]string, error) {
	// 参数验证
	if err := ValidateDomainName(domain); err != nil {
		return nil, fmt.Errorf("域名验证失败: %v", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("记录列表不能为空")
	}

	if len(records) > 20 { // 限制批量操作的数量
		return nil, fmt.Errorf("批量创建记录数量不能超过20条")
	}

	var recordIds []string
	var errors []string

	// 逐个创建记录（DNSPod API v3暂不支持真正的批量操作）
	for i, record := range records {
		// 验证每个记录的参数
		if err := ValidateCreateRecordParams(domain, record.SubDomain, record.RecordType, record.Value,
			SafeUint64ToInt(*record.TTL)); err != nil {
			errors = append(errors, fmt.Sprintf("记录%d: %v", i+1, err))
			continue
		}

		// 创建记录
		recordId, err := p.CreateRecord(domain, record.SubDomain, record.RecordType, record.Value,
			SafeUint64ToInt(*record.TTL))
		if err != nil {
			errors = append(errors, fmt.Sprintf("记录%d创建失败: %v", i+1, err))
			continue
		}

		recordIds = append(recordIds, recordId)
	}

	// 如果有错误，返回部分成功的结果和错误信息
	if len(errors) > 0 {
		return recordIds, fmt.Errorf("部分记录创建失败: %s",
			fmt.Sprintf("[%s]", fmt.Sprintf("%v", errors)))
	}

	return recordIds, nil
}
