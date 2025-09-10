package services

import (
	"domain-manager/internal/models"
	"domain-manager/internal/providers"
	"errors"
	"fmt"

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

	switch dnsProvider.Type {
	case "dnspod":
		return providers.NewDNSPodProvider(dnsProvider.Config)
	default:
		return nil, fmt.Errorf("不支持的DNS服务商类型: %s", dnsProvider.Type)
	}
}
