package providers

import "fmt"

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