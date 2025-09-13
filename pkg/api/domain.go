package api

import (
	"domain-max/pkg/dns/models"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DomainHandler 域名处理器
type DomainHandler struct {
	db *gorm.DB
}

// NewDomainHandler 创建新的域名处理器
func NewDomainHandler(db *gorm.DB) *DomainHandler {
	return &DomainHandler{db: db}
}

// ListDomains 获取域名列表
func (h *DomainHandler) ListDomains(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	name := c.Query("name")
	isActive := c.Query("is_active")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 构建查询
	query := h.db.Model(&models.Domain{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if isActive != "" {
		query = query.Where("is_active = ?", isActive == "true")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 分页查询
	var domains []models.Domain
	offset := (page - 1) * pageSize
	if err := query.Preload("DNSRecords", func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ?", userID).Order("created_at DESC")
	}).Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&domains).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"domains": domains,
		"total":   total,
		"page":    page,
		"page_size": pageSize,
	})
}

// GetDomain 获取单个域名
func (h *DomainHandler) GetDomain(c *gin.Context) {
	id := c.Param("id")
	var domain models.Domain
	if err := h.db.Preload("DNSRecords").First(&domain, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "域名不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"domain": domain,
	})
}

// CreateDomain 创建域名
func (h *DomainHandler) CreateDomain(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		DomainType  string `json:"domain_type"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证域名格式
	if err := validateDomainName(req.Name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查域名是否已存在
	var existingDomain models.Domain
	if err := h.db.Where("name = ?", req.Name).First(&existingDomain).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "域名已存在"})
		return
	}

	// 创建域名
	domain := models.Domain{
		Name:        req.Name,
		DomainType:  req.DomainType,
		IsActive:    true,
		Description: req.Description,
	}

	if err := h.db.Create(&domain).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "创建成功",
		"domain":  domain,
	})
}

// UpdateDomain 更新域名
func (h *DomainHandler) UpdateDomain(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	id := c.Param("id")
	var domain models.Domain
	if err := h.db.First(&domain, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "域名不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		DomainType  string `json:"domain_type"`
		IsActive    *bool  `json:"is_active"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新字段
	if req.Name != "" {
		// 验证域名格式
		if err := validateDomainName(req.Name); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// 检查域名是否已存在（排除当前域名）
		var existingDomain models.Domain
		if err := h.db.Where("name = ? AND id != ?", req.Name, id).First(&existingDomain).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "域名已存在"})
			return
		}
		domain.Name = req.Name
	}
	if req.DomainType != "" {
		domain.DomainType = req.DomainType
	}
	if req.IsActive != nil {
		domain.IsActive = *req.IsActive
	}
	if req.Description != "" {
		domain.Description = req.Description
	}

	if err := h.db.Save(&domain).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"domain":  domain,
	})
}

// DeleteDomain 删除域名
func (h *DomainHandler) DeleteDomain(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	id := c.Param("id")
	
	// 检查域名是否存在
	var domain models.Domain
	if err := h.db.First(&domain, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "域名不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 检查是否有关联的DNS记录
	var count int64
	if err := h.db.Model(&models.DNSRecord{}).Where("domain_id = ?", id).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该域名下还有DNS记录，无法删除"})
		return
	}

	if err := h.db.Delete(&domain).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
	})
}

// GetDomainDNSRecords 获取域名的DNS记录
func (h *DomainHandler) GetDomainDNSRecords(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	domainID := c.Param("id")
	
	// 检查域名是否存在
	var domain models.Domain
	if err := h.db.First(&domain, domainID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "域名不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	recordType := c.Query("type")
	subdomain := c.Query("subdomain")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 构建查询
	query := h.db.Model(&models.DNSRecord{}).Where("domain_id = ? AND user_id = ?", domainID, userID)

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
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"records": records,
		"total":   total,
		"page":    page,
		"page_size": pageSize,
		"domain":  domain,
	})
}

// GetDomainStats 获取域名统计信息
func (h *DomainHandler) GetDomainStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	domainID := c.Param("id")
	
	// 检查域名是否存在
	var domain models.Domain
	if err := h.db.First(&domain, domainID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "域名不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 统计各种类型的DNS记录数量
	var stats []struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}

	if err := h.db.Model(&models.DNSRecord{}).
		Where("domain_id = ? AND user_id = ?", domainID, userID).
		Select("type, count(*) as count").
		Group("type").
		Find(&stats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 统计总记录数
	var totalRecords int64
	if err := h.db.Model(&models.DNSRecord{}).
		Where("domain_id = ? AND user_id = ?", domainID, userID).
		Count(&totalRecords).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"domain":        domain,
		"total_records": totalRecords,
		"type_stats":    stats,
	})
}

// validateDomainName 验证域名格式
func validateDomainName(domain string) error {
	if len(domain) == 0 {
		return errors.New("域名不能为空")
	}
	
	if len(domain) > 253 {
		return errors.New("域名长度不能超过253个字符")
	}
	
	// 检查域名格式
	domainPattern := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-\.]*[a-zA-Z0-9])?$`)
	if !domainPattern.MatchString(domain) {
		return errors.New("域名格式不正确")
	}
	
	// 检查是否以点结尾（完全限定域名）
	if strings.HasSuffix(domain, ".") {
		domain = domain[:len(domain)-1] // 移除尾部的点进行验证
	}
	
	// 验证每个标签
	labels := strings.Split(domain, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return errors.New("域名标签长度必须在1-63个字符之间")
		}
		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return errors.New("域名标签不能以连字符开头或结尾")
		}
	}
	
	return nil
}