package providers

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
