package services

import (
	"domain-manager/internal/config"
	"domain-manager/internal/models"
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

type AdminService struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewAdminService(db *gorm.DB, cfg *config.Config) *AdminService {
	return &AdminService{db: db, cfg: cfg}
}

// 获取所有用户
func (s *AdminService) GetUsers(page, pageSize int, search string) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := s.db.Model(&models.User{})

	if search != "" {
		query = query.Where("email LIKE ?", "%"+search+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// 获取单个用户
func (s *AdminService) GetUser(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("DNSRecords").First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	return &user, nil
}

// 更新用户
func (s *AdminService) UpdateUser(userID uint, updates map[string]interface{}) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	if err := s.db.Model(&user).Updates(updates).Error; err != nil {
		return nil, errors.New("用户更新失败")
	}

	return &user, nil
}

// 删除用户
func (s *AdminService) DeleteUser(userID uint) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	if user.IsAdmin {
		return errors.New("不能删除管理员账户")
	}

	// 软删除用户
	if err := s.db.Delete(&user).Error; err != nil {
		return errors.New("用户删除失败")
	}

	return nil
}

// 获取所有域名
func (s *AdminService) GetDomains() ([]models.Domain, error) {
	var domains []models.Domain
	err := s.db.Find(&domains).Error
	return domains, err
}

// 创建域名
func (s *AdminService) CreateDomain(name string) (*models.Domain, error) {
	// 检查域名是否已存在
	var existingDomain models.Domain
	if err := s.db.Where("name = ?", name).First(&existingDomain).Error; err == nil {
		return nil, errors.New("域名已存在")
	}

	domain := models.Domain{
		Name:     name,
		IsActive: true,
	}

	if err := s.db.Create(&domain).Error; err != nil {
		return nil, errors.New("域名创建失败")
	}

	return &domain, nil
}

// 更新域名
func (s *AdminService) UpdateDomain(domainID uint, updates map[string]interface{}) (*models.Domain, error) {
	var domain models.Domain
	if err := s.db.First(&domain, domainID).Error; err != nil {
		return nil, errors.New("域名不存在")
	}

	if err := s.db.Model(&domain).Updates(updates).Error; err != nil {
		return nil, errors.New("域名更新失败")
	}

	return &domain, nil
}

// 删除域名
func (s *AdminService) DeleteDomain(domainID uint) error {
	var domain models.Domain
	if err := s.db.First(&domain, domainID).Error; err != nil {
		return errors.New("域名不存在")
	}

	// 检查是否有关联的DNS记录
	var recordCount int64
	s.db.Model(&models.DNSRecord{}).Where("domain_id = ?", domainID).Count(&recordCount)
	if recordCount > 0 {
		return errors.New("该域名下还有DNS记录，请先删除相关记录")
	}

	if err := s.db.Delete(&domain).Error; err != nil {
		return errors.New("域名删除失败")
	}

	return nil
}

// 获取DNS服务商
func (s *AdminService) GetDNSProviders() ([]models.DNSProvider, error) {
	var providers []models.DNSProvider
	err := s.db.Find(&providers).Error
	return providers, err
}

// 创建DNS服务商
func (s *AdminService) CreateDNSProvider(name, providerType, config string) (*models.DNSProvider, error) {
	// 验证配置格式
	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(config), &configMap); err != nil {
		return nil, errors.New("配置格式无效")
	}

	provider := models.DNSProvider{
		Name:     name,
		Type:     providerType,
		Config:   config,
		IsActive: true,
	}

	if err := s.db.Create(&provider).Error; err != nil {
		return nil, errors.New("DNS服务商创建失败")
	}

	return &provider, nil
}

// 更新DNS服务商
func (s *AdminService) UpdateDNSProvider(providerID uint, updates map[string]interface{}) (*models.DNSProvider, error) {
	var provider models.DNSProvider
	if err := s.db.First(&provider, providerID).Error; err != nil {
		return nil, errors.New("DNS服务商不存在")
	}

	// 如果更新配置，验证格式
	if config, exists := updates["config"]; exists {
		var configMap map[string]interface{}
		if err := json.Unmarshal([]byte(config.(string)), &configMap); err != nil {
			return nil, errors.New("配置格式无效")
		}
	}

	if err := s.db.Model(&provider).Updates(updates).Error; err != nil {
		return nil, errors.New("DNS服务商更新失败")
	}

	return &provider, nil
}

// 删除DNS服务商
func (s *AdminService) DeleteDNSProvider(providerID uint) error {
	var provider models.DNSProvider
	if err := s.db.First(&provider, providerID).Error; err != nil {
		return errors.New("DNS服务商不存在")
	}

	if err := s.db.Delete(&provider).Error; err != nil {
		return errors.New("DNS服务商删除失败")
	}

	return nil
}

// 获取系统统计信息
func (s *AdminService) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 用户统计
	var userCount, activeUserCount int64
	s.db.Model(&models.User{}).Count(&userCount)
	s.db.Model(&models.User{}).Where("is_active = true").Count(&activeUserCount)

	// 域名统计
	var domainCount, activeDomainCount int64
	s.db.Model(&models.Domain{}).Count(&domainCount)
	s.db.Model(&models.Domain{}).Where("is_active = true").Count(&activeDomainCount)

	// DNS记录统计
	var recordCount int64
	s.db.Model(&models.DNSRecord{}).Count(&recordCount)

	// DNS服务商统计
	var providerCount int64
	s.db.Model(&models.DNSProvider{}).Count(&providerCount)

	stats["users"] = map[string]interface{}{
		"total":  userCount,
		"active": activeUserCount,
	}
	stats["domains"] = map[string]interface{}{
		"total":  domainCount,
		"active": activeDomainCount,
	}
	stats["records"] = recordCount
	stats["providers"] = providerCount

	return stats, nil
}
