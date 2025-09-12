package services

import (
	"domain-manager/internal/models"
	"domain-manager/internal/providers"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// BatchCreateDNSRecords 批量创建DNS记录
func (s *DNSService) BatchCreateDNSRecords(userID uint, req models.BatchDNSRecordRequest) ([]models.DNSRecord, []error) {
	var successRecords []models.DNSRecord
	var errorList []error

	// 检查用户配额
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, []error{errors.New("用户不存在")}
	}

	var currentRecordCount int64
	s.db.Model(&models.DNSRecord{}).Where("user_id = ?", userID).Count(&currentRecordCount)

	if user.DNSRecordQuota > 0 && int(currentRecordCount)+len(req.Records) > user.DNSRecordQuota {
		return nil, []error{fmt.Errorf("批量创建将超出DNS记录配额限制，当前：%d，配额：%d，尝试创建：%d", 
			currentRecordCount, user.DNSRecordQuota, len(req.Records))}
	}

	// 获取DNS服务商
	provider, err := s.getDNSProvider()
	if err != nil {
		return nil, []error{fmt.Errorf("获取DNS服务商失败: %v", err)}
	}

	// 逐个创建记录
	for i, recordReq := range req.Records {
		record, err := s.createSingleRecord(userID, recordReq, provider)
		if err != nil {
			errorList = append(errorList, fmt.Errorf("第%d条记录创建失败: %v", i+1, err))
		} else {
			successRecords = append(successRecords, *record)
		}
	}

	return successRecords, errorList
}

// createSingleRecord 创建单个DNS记录（内部方法）
func (s *DNSService) createSingleRecord(userID uint, req models.CreateDNSRecordRequest, provider providers.DNSProvider) (*models.DNSRecord, error) {
	// 验证请求数据
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("请求参数验证失败: %v", err)
	}

	// 检查域名是否存在且可用
	var domain models.Domain
	if err := s.db.Where("id = ? AND is_active = true", req.DomainID).First(&domain).Error; err != nil {
		return nil, errors.New("域名不存在或不可用")
	}

	// 检查子域名是否已被占用
	var existingRecord models.DNSRecord
	if err := s.db.Where("domain_id = ? AND subdomain = ? AND type = ?",
		req.DomainID, req.Subdomain, req.Type).First(&existingRecord).Error; err == nil {
		return nil, fmt.Errorf("该子域名的%s记录已存在", req.Type)
	}

	// 创建DNS记录
	record := models.DNSRecord{
		UserID:    userID,
		DomainID:  req.DomainID,
		Subdomain: req.Subdomain,
		Type:      req.Type,
		Value:     req.Value,
		TTL:       req.TTL,
		Priority:  req.Priority,
		Weight:    req.Weight,
		Port:      req.Port,
		Comment:   req.Comment,
		Status:    "active",
	}

	if record.TTL == 0 {
		record.TTL = 600 // 默认TTL
	}

	// 验证DNS记录数据
	if err := record.ValidateDNSRecord(); err != nil {
		return nil, fmt.Errorf("DNS记录验证失败: %v", err)
	}

	// 调用DNS服务商API创建记录
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

// ExportDNSRecords 导出用户的DNS记录
func (s *DNSService) ExportDNSRecords(userID uint) (*models.DNSRecordExportResponse, error) {
	var records []models.DNSRecord
	err := s.db.Preload("Domain").Where("user_id = ?", userID).Find(&records).Error
	if err != nil {
		return nil, fmt.Errorf("获取DNS记录失败: %v", err)
	}

	var exportRecords []models.DNSRecordExport
	for _, record := range records {
		exportRecord := models.DNSRecordExport{
			Subdomain: record.Subdomain,
			Type:      record.Type,
			Value:     record.Value,
			TTL:       record.TTL,
			Comment:   record.Comment,
			Domain:    record.Domain.Name,
		}

		// 只有在有值时才包含优先级、权重和端口
		if record.Priority > 0 {
			exportRecord.Priority = record.Priority
		}
		if record.Weight > 0 {
			exportRecord.Weight = record.Weight
		}
		if record.Port > 0 {
			exportRecord.Port = record.Port
		}

		exportRecords = append(exportRecords, exportRecord)
	}

	return &models.DNSRecordExportResponse{
		Records: exportRecords,
		Total:   len(exportRecords),
	}, nil
}

// ImportDNSRecords 导入DNS记录
func (s *DNSService) ImportDNSRecords(userID uint, records []models.DNSRecordExport) ([]models.DNSRecord, []error) {
	var successRecords []models.DNSRecord
	var errorList []error

	// 检查用户配额
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, []error{errors.New("用户不存在")}
	}

	var currentRecordCount int64
	s.db.Model(&models.DNSRecord{}).Where("user_id = ?", userID).Count(&currentRecordCount)

	if user.DNSRecordQuota > 0 && int(currentRecordCount)+len(records) > user.DNSRecordQuota {
		return nil, []error{fmt.Errorf("导入将超出DNS记录配额限制，当前：%d，配额：%d，尝试导入：%d", 
			currentRecordCount, user.DNSRecordQuota, len(records))}
	}

	// 获取DNS服务商
	provider, err := s.getDNSProvider()
	if err != nil {
		return nil, []error{fmt.Errorf("获取DNS服务商失败: %v", err)}
	}

	// 逐个导入记录
	for i, exportRecord := range records {
		// 查找域名ID
		var domain models.Domain
		if err := s.db.Where("name = ? AND is_active = true", exportRecord.Domain).First(&domain).Error; err != nil {
			errorList = append(errorList, fmt.Errorf("第%d条记录导入失败: 域名 %s 不存在或不可用", i+1, exportRecord.Domain))
			continue
		}

		// 转换为创建请求
		createReq := models.CreateDNSRecordRequest{
			DomainID:  domain.ID,
			Subdomain: exportRecord.Subdomain,
			Type:      exportRecord.Type,
			Value:     exportRecord.Value,
			TTL:       exportRecord.TTL,
			Priority:  exportRecord.Priority,
			Weight:    exportRecord.Weight,
			Port:      exportRecord.Port,
			Comment:   exportRecord.Comment,
		}

		record, err := s.createSingleRecord(userID, createReq, provider)
		if err != nil {
			errorList = append(errorList, fmt.Errorf("第%d条记录导入失败: %v", i+1, err))
		} else {
			successRecords = append(successRecords, *record)
		}
	}

	return successRecords, errorList
}

// ValidateDNSRecordFile 验证DNS记录文件格式
func (s *DNSService) ValidateDNSRecordFile(content string) ([]models.DNSRecordExport, error) {
	lines := strings.Split(content, "\n")
	var records []models.DNSRecordExport
	var errors []string

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue // 跳过空行和注释行
		}

		// 解析记录格式：subdomain.domain TYPE value [TTL] [priority] [weight] [port] [comment]
		parts := strings.Fields(line)
		if len(parts) < 3 {
			errors = append(errors, fmt.Sprintf("第%d行格式错误：至少需要子域名、类型和值", i+1))
			continue
		}

		// 解析域名和子域名
		fullDomain := parts[0]
		domainParts := strings.Split(fullDomain, ".")
		if len(domainParts) < 2 {
			errors = append(errors, fmt.Sprintf("第%d行域名格式错误", i+1))
			continue
		}

		subdomain := domainParts[0]
		domain := strings.Join(domainParts[1:], ".")

		record := models.DNSRecordExport{
			Subdomain: subdomain,
			Domain:    domain,
			Type:      strings.ToUpper(parts[1]),
			Value:     parts[2],
			TTL:       600, // 默认TTL
		}

		// 解析可选字段
		if len(parts) > 3 {
			if ttl := parseInt(parts[3]); ttl > 0 {
				record.TTL = ttl
			}
		}
		if len(parts) > 4 {
			record.Priority = parseInt(parts[4])
		}
		if len(parts) > 5 {
			record.Weight = parseInt(parts[5])
		}
		if len(parts) > 6 {
			record.Port = parseInt(parts[6])
		}
		if len(parts) > 7 {
			record.Comment = strings.Join(parts[7:], " ")
		}

		records = append(records, record)
	}

	if len(errors) > 0 {
		return nil, fmt.Errorf("文件格式错误：\n%s", strings.Join(errors, "\n"))
	}

	return records, nil
}

// parseInt 安全地解析整数
func parseInt(s string) int {
	if i, err := fmt.Sscanf(s, "%d", new(int)); err == nil && i == 1 {
		var result int
		fmt.Sscanf(s, "%d", &result)
		return result
	}
	return 0
}