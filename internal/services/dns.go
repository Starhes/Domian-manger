package services

import (
	"domain-manager/internal/models"
	"domain-manager/internal/providers"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/net/publicsuffix"
	"gorm.io/gorm"
)

type DNSService struct {
	db *gorm.DB
}

func NewDNSService(db *gorm.DB) *DNSService {
	return &DNSService{db: db}
}

// 获取用户的DNS记录
func (s *DNSService) GetUserDNSRecords(userID uint) ([]models.DNSRecord, error) {
	var records []models.DNSRecord
	err := s.db.Preload("Domain").Where("user_id = ?", userID).Find(&records).Error
	return records, err
}

// 创建DNS记录
func (s *DNSService) CreateDNSRecord(userID uint, req models.CreateDNSRecordRequest) (*models.DNSRecord, error) {
	// 检查域名是否存在且可用
	var domain models.Domain
	if err := s.db.Where("id = ? AND is_active = true", req.DomainID).First(&domain).Error; err != nil {
		return nil, errors.New("域名不存在或不可用")
	}

	// 检查子域名是否已被占用
	var existingRecord models.DNSRecord
	if err := s.db.Where("domain_id = ? AND subdomain = ? AND type = ?", 
		req.DomainID, req.Subdomain, req.Type).First(&existingRecord).Error; err == nil {
		return nil, errors.New("该子域名记录已存在")
	}

	// 创建DNS记录
	record := models.DNSRecord{
		UserID:    userID,
		DomainID:  req.DomainID,
		Subdomain: req.Subdomain,
		Type:      req.Type,
		Value:     req.Value,
		TTL:       req.TTL,
	}

	if record.TTL == 0 {
		record.TTL = 600 // 默认TTL
	}

	// 调用DNS服务商API创建记录
	provider, err := s.getDNSProvider()
	if err != nil {
		return nil, fmt.Errorf("获取DNS服务商失败: %v", err)
	}

	externalID, err := provider.CreateRecord(domain.Name, record.Subdomain, record.Type, record.Value, record.TTL)
	if err != nil {
		return nil, fmt.Errorf("DNS记录创建失败: %v", err)
	}

	record.ExternalID = externalID

	if err := s.db.Create(&record).Error; err != nil {
		// 如果数据库保存失败，尝试删除已创建的DNS记录
		provider.DeleteRecord(domain.Name, externalID)
		return nil, errors.New("DNS记录保存失败")
	}

	// 预加载关联数据
	s.db.Preload("Domain").First(&record, record.ID)

	return &record, nil
}

// 更新DNS记录
func (s *DNSService) UpdateDNSRecord(userID, recordID uint, req models.UpdateDNSRecordRequest) (*models.DNSRecord, error) {
	var record models.DNSRecord
	if err := s.db.Preload("Domain").Where("id = ? AND user_id = ?", recordID, userID).First(&record).Error; err != nil {
		return nil, errors.New("DNS记录不存在")
	}

	// 获取DNS服务商
	provider, err := s.getDNSProvider()
	if err != nil {
		return nil, fmt.Errorf("获取DNS服务商失败: %v", err)
	}

	// 更新字段
	updated := false
	if req.Subdomain != "" && req.Subdomain != record.Subdomain {
		record.Subdomain = req.Subdomain
		updated = true
	}
	if req.Type != "" && req.Type != record.Type {
		record.Type = req.Type
		updated = true
	}
	if req.Value != "" && req.Value != record.Value {
		record.Value = req.Value
		updated = true
	}
	if req.TTL > 0 && req.TTL != record.TTL {
		record.TTL = req.TTL
		updated = true
	}

	if !updated {
		return &record, nil
	}

	// 调用DNS服务商API更新记录
	if err := provider.UpdateRecord(record.Domain.Name, record.ExternalID, record.Subdomain, record.Type, record.Value, record.TTL); err != nil {
		return nil, fmt.Errorf("DNS记录更新失败: %v", err)
	}

	// 保存到数据库
	if err := s.db.Save(&record).Error; err != nil {
		return nil, errors.New("DNS记录保存失败")
	}

	return &record, nil
}

// 删除DNS记录
func (s *DNSService) DeleteDNSRecord(userID, recordID uint) error {
	var record models.DNSRecord
	if err := s.db.Preload("Domain").Where("id = ? AND user_id = ?", recordID, userID).First(&record).Error; err != nil {
		return errors.New("DNS记录不存在")
	}

	// 获取DNS服务商
	provider, err := s.getDNSProvider()
	if err != nil {
		return fmt.Errorf("获取DNS服务商失败: %v", err)
	}

	// 调用DNS服务商API删除记录
	if err := provider.DeleteRecord(record.Domain.Name, record.ExternalID); err != nil {
		return fmt.Errorf("DNS记录删除失败: %v", err)
	}

	// 从数据库删除
	if err := s.db.Delete(&record).Error; err != nil {
		return errors.New("DNS记录删除失败")
	}

	return nil
}

// 获取可用域名列表
func (s *DNSService) GetAvailableDomains() ([]models.Domain, error) {
	var domains []models.Domain
	err := s.db.Where("is_active = true").Find(&domains).Error
	return domains, err
}

// 获取DNS服务商
func (s *DNSService) getDNSProvider() (providers.DNSProvider, error) {
	var dnsProvider models.DNSProvider
	if err := s.db.Where("is_active = true").First(&dnsProvider).Error; err != nil {
		return nil, errors.New("没有可用的DNS服务商")
	}

	// 使用新的provider工厂
	return providers.NewDNSProvider(dnsProvider.Type, dnsProvider.Config)
}

// SyncDomains 同步DNS服务商的域名到数据库
func (s *DNSService) SyncDomains() error {
	provider, err := s.getDNSProvider()
	if err != nil {
		return fmt.Errorf("获取DNS服务商失败: %v", err)
	}

	externalDomains, err := provider.GetDomains()
	if err != nil {
		return fmt.Errorf("从服务商获取域名列表失败: %v", err)
	}

	for _, extDomain := range externalDomains {
		var domain models.Domain
		err := s.db.Where("name = ?", extDomain.Name).First(&domain).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			// 查询出错
			continue
		}

		domainType := classifyDomain(extDomain.Name)

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 域名不存在，创建新记录
			newDomain := models.Domain{
				Name:       extDomain.Name,
				IsActive:   extDomain.Status == "enable", // 假设 "enable" 是活动状态
				DomainType: domainType,
				// 需要确定ProviderID和UserID的来源
			}
			s.db.Create(&newDomain)
		} else {
			// 域名已存在，更新信息
			s.db.Model(&domain).Updates(models.Domain{
				IsActive:   extDomain.Status == "enable",
				DomainType: domainType,
			})
		}
	}

	return nil
}

// classifyDomain 根据域名名称判断其类型
func classifyDomain(domainName string) string {
	// 使用公共后缀列表获取有效的顶级域名+1
	eTLDPlusOne, err := publicsuffix.EffectiveTLDPlusOne(domainName)
	if err != nil {
		// 对于无效的域名（例如没有点的裸域名），回退到简单分割
		if !strings.Contains(domainName, ".") {
			return "无法识别"
		}
		parts := strings.Split(domainName, ".")
		switch len(parts) {
		case 2:
			return "二级域名"
		case 3:
			return "三级域名"
		default:
			return "多级域名"
		}
	}

	// 如果域名本身就是eTLD+1，那么它就是二级域名
	if domainName == eTLDPlusOne {
		return "二级域名"
	}

	// 检查是否是eTLD+1的子域名
	if strings.HasSuffix(domainName, "."+eTLDPlusOne) {
		prefix := strings.TrimSuffix(domainName, "."+eTLDPlusOne)
		// 计算前缀中的点数来确定级别
		dots := strings.Count(prefix, ".")
		switch dots {
		case 0:
			return "三级域名" // a.example.com
		case 1:
			return "四级域名" // a.b.example.com
		default:
			return "多级域名"
		}
	}

	return "未知类型"
}
