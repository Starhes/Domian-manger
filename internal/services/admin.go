package services

import (
	"domain-manager/internal/config"
	"domain-manager/internal/models"
	"domain-manager/internal/utils"
	"encoding/json"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminService struct {
	db     *gorm.DB
	cfg    *config.Config
	crypto *utils.CryptoService
}

func NewAdminService(db *gorm.DB, cfg *config.Config) *AdminService {
	// 初始化加密服务
	crypto, err := utils.NewCryptoService(cfg.EncryptionKey[:32]) // 使用前32字节作为密钥
	if err != nil {
		// 如果加密服务初始化失败，记录错误但不阻止服务启动
		// 在生产环境中，应该返回错误而不是继续
		crypto = nil
	}
	
	return &AdminService{
		db:     db,
		cfg:    cfg,
		crypto: crypto,
	}
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

	// 处理邮箱更新
	if email, ok := updates["email"].(string); ok && email != user.Email {
		// 使用统一的邮箱验证函数
		if err := models.ValidateEmail(email); err != nil {
			return nil, err
		}
		// 检查新邮箱是否已被占用
		var existingUser models.User
		if err := s.db.Where("email = ? AND id != ?", email, userID).First(&existingUser).Error; err == nil {
			return nil, errors.New("该邮箱已被其他用户注册")
		}
		user.Email = email
	}

	// 处理密码更新
	if password, ok := updates["password"].(string); ok && password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.New("密码加密失败")
		}
		user.Password = string(hashedPassword)
	}

	// 处理其他字段的更新
	if isActive, ok := updates["is_active"].(bool); ok {
		user.IsActive = isActive
	}
	if isAdmin, ok := updates["is_admin"].(bool); ok {
		user.IsAdmin = isAdmin
	}

	if err := s.db.Save(&user).Error; err != nil {
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

// SyncDomains 同步域名
func (s *AdminService) SyncDomains() error {
	dnsService := NewDNSService(s.db)
	return dnsService.SyncDomains()
}

// ========================= SMTP配置管理 =========================

// 获取所有SMTP配置
func (s *AdminService) GetSMTPConfigs() ([]models.SMTPConfigResponse, error) {
	var configs []models.SMTPConfig
	if err := s.db.Find(&configs).Error; err != nil {
		return nil, err
	}

	// 转换为响应格式（脱敏）
	var response []models.SMTPConfigResponse
	for _, config := range configs {
		response = append(response, models.SMTPConfigResponse{
			ID:          config.ID,
			Name:        config.Name,
			Host:        config.Host,
			Port:        config.Port,
			Username:    config.Username,
			FromEmail:   config.FromEmail,
			FromName:    config.FromName,
			IsActive:    config.IsActive,
			IsDefault:   config.IsDefault,
			UseTLS:      config.UseTLS,
			Description: config.Description,
			LastTestAt:  config.LastTestAt,
			TestResult:  config.TestResult,
			CreatedAt:   config.CreatedAt,
			UpdatedAt:   config.UpdatedAt,
		})
	}

	return response, nil
}

// 获取单个SMTP配置
func (s *AdminService) GetSMTPConfig(configID uint) (*models.SMTPConfigResponse, error) {
	var config models.SMTPConfig
	if err := s.db.First(&config, configID).Error; err != nil {
		return nil, errors.New("SMTP配置不存在")
	}

	return &models.SMTPConfigResponse{
		ID:          config.ID,
		Name:        config.Name,
		Host:        config.Host,
		Port:        config.Port,
		Username:    config.Username,
		FromEmail:   config.FromEmail,
		FromName:    config.FromName,
		IsActive:    config.IsActive,
		IsDefault:   config.IsDefault,
		UseTLS:      config.UseTLS,
		Description: config.Description,
		LastTestAt:  config.LastTestAt,
		TestResult:  config.TestResult,
		CreatedAt:   config.CreatedAt,
		UpdatedAt:   config.UpdatedAt,
	}, nil
}

// 创建SMTP配置
func (s *AdminService) CreateSMTPConfig(req models.CreateSMTPConfigRequest) (*models.SMTPConfigResponse, error) {
	// 检查配置名称是否已存在
	var existingConfig models.SMTPConfig
	if err := s.db.Where("name = ?", req.Name).First(&existingConfig).Error; err == nil {
		return nil, errors.New("配置名称已存在")
	}

	// 使用AES加密密码
	if s.crypto == nil {
		return nil, errors.New("加密服务未初始化")
	}
	
	encryptedPassword, err := s.crypto.Encrypt(req.Password)
	if err != nil {
		return nil, errors.New("密码加密失败: " + err.Error())
	}

	config := models.SMTPConfig{
		Name:        req.Name,
		Host:        req.Host,
		Port:        req.Port,
		Username:    req.Username,
		Password:    encryptedPassword,
		FromEmail:   req.FromEmail,
		FromName:    req.FromName,
		UseTLS:      req.UseTLS,
		Description: req.Description,
		IsActive:    false, // 默认不激活
		IsDefault:   false,
	}

	if err := s.db.Create(&config).Error; err != nil {
		return nil, errors.New("SMTP配置创建失败")
	}

	return s.GetSMTPConfig(config.ID)
}

// 更新SMTP配置
func (s *AdminService) UpdateSMTPConfig(configID uint, req models.UpdateSMTPConfigRequest) (*models.SMTPConfigResponse, error) {
	var config models.SMTPConfig
	if err := s.db.First(&config, configID).Error; err != nil {
		return nil, errors.New("SMTP配置不存在")
	}

	// 检查名称冲突
	if req.Name != "" && req.Name != config.Name {
		var existingConfig models.SMTPConfig
		if err := s.db.Where("name = ? AND id != ?", req.Name, configID).First(&existingConfig).Error; err == nil {
			return nil, errors.New("配置名称已存在")
		}
		config.Name = req.Name
	}

	// 更新其他字段
	if req.Host != "" {
		config.Host = req.Host
	}
	if req.Port > 0 {
		config.Port = req.Port
	}
	if req.Username != "" {
		config.Username = req.Username
	}
	if req.Password != "" {
		if s.crypto == nil {
			return nil, errors.New("加密服务未初始化")
		}
		
		encryptedPassword, err := s.crypto.Encrypt(req.Password)
		if err != nil {
			return nil, errors.New("密码加密失败: " + err.Error())
		}
		config.Password = encryptedPassword
	}
	if req.FromEmail != "" {
		config.FromEmail = req.FromEmail
	}
	if req.FromName != "" {
		config.FromName = req.FromName
	}
	if req.UseTLS != nil {
		config.UseTLS = *req.UseTLS
	}
	if req.Description != "" {
		config.Description = req.Description
	}

	if err := s.db.Save(&config).Error; err != nil {
		return nil, errors.New("SMTP配置更新失败")
	}

	return s.GetSMTPConfig(config.ID)
}

// 删除SMTP配置
func (s *AdminService) DeleteSMTPConfig(configID uint) error {
	var config models.SMTPConfig
	if err := s.db.First(&config, configID).Error; err != nil {
		return errors.New("SMTP配置不存在")
	}

	// 检查是否为默认配置
	if config.IsDefault {
		return errors.New("不能删除默认配置")
	}

	if err := s.db.Delete(&config).Error; err != nil {
		return errors.New("SMTP配置删除失败")
	}

	return nil
}

// 激活SMTP配置
func (s *AdminService) ActivateSMTPConfig(configID uint) error {
	var config models.SMTPConfig
	if err := s.db.First(&config, configID).Error; err != nil {
		return errors.New("SMTP配置不存在")
	}

	// 先取消其他配置的激活状态
	if err := s.db.Model(&models.SMTPConfig{}).Where("id != ?", configID).Update("is_active", false).Error; err != nil {
		return errors.New("更新其他配置状态失败")
	}

	// 激活当前配置
	config.IsActive = true
	if err := s.db.Save(&config).Error; err != nil {
		return errors.New("配置激活失败")
	}

	return nil
}

// 设置默认SMTP配置
func (s *AdminService) SetDefaultSMTPConfig(configID uint) error {
	var config models.SMTPConfig
	if err := s.db.First(&config, configID).Error; err != nil {
		return errors.New("SMTP配置不存在")
	}

	// 先取消其他配置的默认状态
	if err := s.db.Model(&models.SMTPConfig{}).Where("id != ?", configID).Update("is_default", false).Error; err != nil {
		return errors.New("更新其他配置状态失败")
	}

	// 设置当前配置为默认
	config.IsDefault = true
	if err := s.db.Save(&config).Error; err != nil {
		return errors.New("设置默认配置失败")
	}

	return nil
}

// 测试SMTP配置
func (s *AdminService) TestSMTPConfig(configID uint, toEmail string) error {
	var smtpConfig models.SMTPConfig
	if err := s.db.First(&smtpConfig, configID).Error; err != nil {
		return errors.New("SMTP配置不存在")
	}

	// 解密SMTP密码
	if s.crypto == nil {
		return errors.New("加密服务未初始化")
	}
	
	decryptedPassword, err := s.crypto.Decrypt(smtpConfig.Password)
	if err != nil {
		return errors.New("密码解密失败: " + err.Error())
	}

	// 创建临时的配置对象用于测试
	tempConfig := &config.Config{
		SMTPHost:     smtpConfig.Host,
		SMTPPort:     smtpConfig.Port,
		SMTPUser:     smtpConfig.Username,
		SMTPPassword: decryptedPassword,
		SMTPFrom:     smtpConfig.FromEmail,
	}

	// 创建邮件服务并发送测试邮件
	emailService := NewEmailService(tempConfig)
	subject := "SMTP配置测试邮件"
	body := "这是一封测试邮件，用于验证SMTP配置是否正常工作。"

	now := time.Now()
	if err := emailService.sendEmail(toEmail, subject, body); err != nil {
		// 记录测试失败
		smtpConfig.TestResult = "测试失败: " + err.Error()
		smtpConfig.LastTestAt = &now
		s.db.Save(&smtpConfig)
		return err
	}

	// 记录测试成功
	smtpConfig.TestResult = "测试成功"
	smtpConfig.LastTestAt = &now
	s.db.Save(&smtpConfig)

	return nil
}
