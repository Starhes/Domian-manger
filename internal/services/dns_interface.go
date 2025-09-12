package services

import (
	"domain-manager/internal/models"
	"gorm.io/gorm"
)

// DNSServiceInterface 定义DNS服务接口
type DNSServiceInterface interface {
	GetUserDNSRecords(userID uint) ([]models.DNSRecord, error)
	CreateDNSRecord(userID uint, req models.CreateDNSRecordRequest) (*models.DNSRecord, error)
	UpdateDNSRecord(userID, recordID uint, req models.UpdateDNSRecordRequest) (*models.DNSRecord, error)
	DeleteDNSRecord(userID, recordID uint) error
	GetAvailableDomains() ([]models.Domain, error)
	SyncDomains() error
	BatchCreateDNSRecords(userID uint, req models.BatchDNSRecordRequest) ([]models.DNSRecord, []error)
	ExportDNSRecords(userID uint) (*models.DNSRecordExportResponse, error)
	ImportDNSRecords(userID uint, records []models.DNSRecordExport) ([]models.DNSRecord, []error)
	ValidateDNSRecordFile(content string) ([]models.DNSRecordExport, error)
	GetDB() *gorm.DB
}

// GetDB 获取数据库连接
func (s *DNSService) GetDB() *gorm.DB {
	return s.db
}