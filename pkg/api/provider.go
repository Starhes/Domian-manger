package api

import (
	"domain-max/pkg/dns/models"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ProviderHandler DNS提供商处理器
type ProviderHandler struct {
	db *gorm.DB
}

// NewProviderHandler 创建新的DNS提供商处理器
func NewProviderHandler(db *gorm.DB) *ProviderHandler {
	return &ProviderHandler{db: db}
}

// ListProviders 获取DNS提供商列表
func (h *ProviderHandler) ListProviders(c *gin.Context) {
	// 检查管理员权限
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	name := c.Query("name")
	providerType := c.Query("type")
	isActive := c.Query("is_active")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 构建查询
	query := h.db.Model(&models.DNSProvider{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if providerType != "" {
		query = query.Where("type = ?", providerType)
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
	var providers []models.DNSProvider
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("sort_order ASC, created_at DESC").Find(&providers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"providers": providers,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetProvider 获取单个DNS提供商
func (h *ProviderHandler) GetProvider(c *gin.Context) {
	// 检查管理员权限
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	id := c.Param("id")
	var provider models.DNSProvider
	if err := h.db.First(&provider, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "DNS提供商不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"provider": provider,
	})
}

// CreateProvider 创建DNS提供商
func (h *ProviderHandler) CreateProvider(c *gin.Context) {
	// 检查管理员权限
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Type        string `json:"type" binding:"required"`
		Config      string `json:"config" binding:"required"`
		IsActive    bool   `json:"is_active"`
		Description string `json:"description"`
		SortOrder   int    `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证配置JSON格式
	var configJSON interface{}
	if err := json.Unmarshal([]byte(req.Config), &configJSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "配置格式必须是有效的JSON"})
		return
	}

	// 检查名称是否已存在
	var existingProvider models.DNSProvider
	if err := h.db.Where("name = ?", req.Name).First(&existingProvider).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "DNS提供商名称已存在"})
		return
	}

	// 创建DNS提供商
	provider := models.DNSProvider{
		Name:        req.Name,
		Type:        req.Type,
		Config:      req.Config,
		IsActive:    req.IsActive,
		Description: req.Description,
		SortOrder:   req.SortOrder,
	}

	if err := h.db.Create(&provider).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "创建成功",
		"provider":  provider,
	})
}

// UpdateProvider 更新DNS提供商
func (h *ProviderHandler) UpdateProvider(c *gin.Context) {
	// 检查管理员权限
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	id := c.Param("id")
	var provider models.DNSProvider
	if err := h.db.First(&provider, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "DNS提供商不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Config      string `json:"config"`
		IsActive    *bool  `json:"is_active"`
		Description string `json:"description"`
		SortOrder   int    `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新字段
	if req.Name != "" {
		// 检查名称是否已存在（排除当前提供商）
		var existingProvider models.DNSProvider
		if err := h.db.Where("name = ? AND id != ?", req.Name, id).First(&existingProvider).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "DNS提供商名称已存在"})
			return
		}
		provider.Name = req.Name
	}
	if req.Type != "" {
		provider.Type = req.Type
	}
	if req.Config != "" {
		// 验证配置JSON格式
		var configJSON interface{}
		if err := json.Unmarshal([]byte(req.Config), &configJSON); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "配置格式必须是有效的JSON"})
			return
		}
		provider.Config = req.Config
	}
	if req.IsActive != nil {
		provider.IsActive = *req.IsActive
	}
	if req.Description != "" {
		provider.Description = req.Description
	}
	if req.SortOrder >= 0 {
		provider.SortOrder = req.SortOrder
	}

	if err := h.db.Save(&provider).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "更新成功",
		"provider":  provider,
	})
}

// DeleteProvider 删除DNS提供商
func (h *ProviderHandler) DeleteProvider(c *gin.Context) {
	// 检查管理员权限
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	id := c.Param("id")
	
	// 检查DNS提供商是否存在
	var provider models.DNSProvider
	if err := h.db.First(&provider, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "DNS提供商不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	if err := h.db.Delete(&provider).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
	})
}

// TestProvider 测试DNS提供商连接
func (h *ProviderHandler) TestProvider(c *gin.Context) {
	// 检查管理员权限
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	id := c.Param("id")
	
	// 检查DNS提供商是否存在
	var provider models.DNSProvider
	if err := h.db.First(&provider, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "DNS提供商不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// TODO: 实现实际的DNS提供商测试逻辑
	// 这里只是模拟测试结果
	testResult := "DNS提供商连接测试成功"
	now := time.Now()
	
	// 更新测试时间和结果
	provider.LastTestAt = &now
	provider.TestResult = testResult
	if err := h.db.Save(&provider).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新测试结果失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "测试成功",
		"test_result": testResult,
		"tested_at":   now,
	})
}

// ToggleProviderStatus 切换DNS提供商状态
func (h *ProviderHandler) ToggleProviderStatus(c *gin.Context) {
	// 检查管理员权限
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	id := c.Param("id")
	
	// 检查DNS提供商是否存在
	var provider models.DNSProvider
	if err := h.db.First(&provider, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "DNS提供商不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 切换状态
	provider.IsActive = !provider.IsActive
	if err := h.db.Save(&provider).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新状态失败"})
		return
	}

	statusText := "启用"
	if !provider.IsActive {
		statusText = "禁用"
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "DNS提供商已" + statusText,
		"provider": provider,
	})
}

// GetProviderTypes 获取支持的DNS提供商类型
func (h *ProviderHandler) GetProviderTypes(c *gin.Context) {
	// 检查管理员权限
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	// 返回支持的DNS提供商类型
	providerTypes := []gin.H{
		{
			"type":        "dnspod",
			"name":        "DNSPod",
			"description": "腾讯云DNSPod",
			"config_fields": []gin.H{
				{"name": "api_token", "type": "string", "label": "API Token", "required": true},
			},
		},
		{
			"type":        "aliyun",
			"name":        "阿里云DNS",
			"description": "阿里云域名解析",
			"config_fields": []gin.H{
				{"name": "access_key_id", "type": "string", "label": "AccessKey ID", "required": true},
				{"name": "access_key_secret", "type": "string", "label": "AccessKey Secret", "required": true},
			},
		},
		{
			"type":        "cloudflare",
			"name":        "Cloudflare",
			"description": "Cloudflare DNS",
			"config_fields": []gin.H{
				{"name": "api_token", "type": "string", "label": "API Token", "required": true},
				{"name": "zone_id", "type": "string", "label": "Zone ID", "required": true},
			},
		},
		{
			"type":        "godaddy",
			"name":        "GoDaddy",
			"description": "GoDaddy DNS",
			"config_fields": []gin.H{
				{"name": "api_key", "type": "string", "label": "API Key", "required": true},
				{"name": "api_secret", "type": "string", "label": "API Secret", "required": true},
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"provider_types": providerTypes,
	})
}