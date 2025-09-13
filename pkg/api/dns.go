package api

import (
	authmodels "domain-max/pkg/auth/models"
	"domain-max/pkg/dns/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DNSHandler DNS记录处理器
type DNSHandler struct {
	db *gorm.DB
}

// NewDNSHandler 创建新的DNS处理器
func NewDNSHandler(db *gorm.DB) *DNSHandler {
	return &DNSHandler{db: db}
}

// ListDNSRecords 获取DNS记录列表
func (h *DNSHandler) ListDNSRecords(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	domainID := c.Query("domain_id")
	recordType := c.Query("type")
	subdomain := c.Query("subdomain")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 构建查询
	query := h.db.Model(&models.DNSRecord{}).Where("user_id = ?", userID)

	if domainID != "" {
		query = query.Where("domain_id = ?", domainID)
	}
	if recordType != "" {
		query = query.Where("type = ?", recordType)
	}
	if subdomain != "" {
		query = query.Where("subdomain LIKE ?", "%"+subdomain+"%")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 分页查询
	var records []models.DNSRecord
	offset := (page - 1) * pageSize
	if err := query.Preload("Domain").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"records": records,
		"total":   total,
		"page":   page,
		"page_size": pageSize,
	})
}

// GetDNSRecord 获取单个DNS记录
func (h *DNSHandler) GetDNSRecord(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	id := c.Param("id")
	var record models.DNSRecord
	if err := h.db.Preload("Domain").Where("id = ? AND user_id = ?", id, userID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "记录不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"record": record,
	})
}

// CreateDNSRecord 创建DNS记录
func (h *DNSHandler) CreateDNSRecord(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	var req models.CreateDNSRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查用户配额
	var user authmodels.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 统计当前DNS记录数量
	var count int64
	if err := h.db.Model(&models.DNSRecord{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	if count >= int64(user.DNSRecordQuota) {
		c.JSON(http.StatusForbidden, gin.H{"error": "已达到DNS记录配额上限"})
		return
	}

	// 验证域名是否存在且属于用户
	var domain models.Domain
	if err := h.db.First(&domain, req.DomainID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "域名不存在"})
		return
	}

	// 创建DNS记录
	record := models.DNSRecord{
		UserID:    userID.(uint),
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

	// 验证记录
	if err := record.ValidateDNSRecord(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Create(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	// 重新加载关联数据
	h.db.Preload("Domain").First(&record, record.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "创建成功",
		"record":  record,
	})
}

// UpdateDNSRecord 更新DNS记录
func (h *DNSHandler) UpdateDNSRecord(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	id := c.Param("id")
	var record models.DNSRecord
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "记录不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	var req models.UpdateDNSRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新字段
	if req.Subdomain != "" {
		record.Subdomain = req.Subdomain
	}
	if req.Type != "" {
		record.Type = req.Type
	}
	if req.Value != "" {
		record.Value = req.Value
	}
	if req.TTL != 0 {
		record.TTL = req.TTL
	}
	record.Priority = req.Priority
	record.Weight = req.Weight
	record.Port = req.Port
	if req.Comment != "" {
		record.Comment = req.Comment
	}

	// 验证记录
	if err := record.ValidateDNSRecord(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Save(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	// 重新加载关联数据
	h.db.Preload("Domain").First(&record, record.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"record":  record,
	})
}

// DeleteDNSRecord 删除DNS记录
func (h *DNSHandler) DeleteDNSRecord(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	id := c.Param("id")
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.DNSRecord{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
	})
}

// BatchCreateDNSRecords 批量创建DNS记录
func (h *DNSHandler) BatchCreateDNSRecords(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	var req models.BatchDNSRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查用户配额
	var user authmodels.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 统计当前DNS记录数量
	var count int64
	if err := h.db.Model(&models.DNSRecord{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	if count+int64(len(req.Records)) > int64(user.DNSRecordQuota) {
		c.JSON(http.StatusForbidden, gin.H{"error": "批量创建将超过DNS记录配额上限"})
		return
	}

	// 验证所有域名是否存在
	domainIDs := make([]uint, 0, len(req.Records))
	for _, recordReq := range req.Records {
		domainIDs = append(domainIDs, recordReq.DomainID)
	}

	var domains []models.Domain
	if err := h.db.Where("id IN ?", domainIDs).Find(&domains).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询域名失败"})
		return
	}

	if len(domains) != len(domainIDs) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "部分域名不存在"})
		return
	}

	// 创建DNS记录
	records := make([]models.DNSRecord, 0, len(req.Records))
	for _, recordReq := range req.Records {
		record := models.DNSRecord{
			UserID:    userID.(uint),
			DomainID:  recordReq.DomainID,
			Subdomain: recordReq.Subdomain,
			Type:      recordReq.Type,
			Value:     recordReq.Value,
			TTL:       recordReq.TTL,
			Priority:  recordReq.Priority,
			Weight:    recordReq.Weight,
			Port:      recordReq.Port,
			Comment:   recordReq.Comment,
			Status:    "active",
		}

		// 验证记录
		if err := record.ValidateDNSRecord(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		records = append(records, record)
	}

	if err := h.db.CreateInBatches(records, 100).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量创建失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "批量创建成功",
		"count":   len(records),
	})
}

// ExportDNSRecords 导出DNS记录
func (h *DNSHandler) ExportDNSRecords(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	domainID := c.Query("domain_id")

	// 构建查询
	query := h.db.Model(&models.DNSRecord{}).Where("user_id = ?", userID)
	if domainID != "" {
		query = query.Where("domain_id = ?", domainID)
	}

	var records []models.DNSRecord
	if err := query.Preload("Domain").Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 转换为导出格式
	exports := make([]models.DNSRecordExport, 0, len(records))
	for _, record := range records {
		exports = append(exports, models.DNSRecordExport{
			Subdomain: record.Subdomain,
			Type:      record.Type,
			Value:     record.Value,
			TTL:       record.TTL,
			Priority:  record.Priority,
			Weight:    record.Weight,
			Port:      record.Port,
			Comment:   record.Comment,
			Domain:    record.Domain.Name,
		})
	}

	c.JSON(http.StatusOK, models.DNSRecordExportResponse{
		Records: exports,
		Total:   len(exports),
	})
}