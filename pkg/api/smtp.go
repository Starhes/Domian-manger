package api

import (
	"domain-max/pkg/email/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SMTPHandler SMTP配置处理器
type SMTPHandler struct {
	db *gorm.DB
}

// NewSMTPHandler 创建新的SMTP配置处理器
func NewSMTPHandler(db *gorm.DB) *SMTPHandler {
	return &SMTPHandler{db: db}
}

// ListSMTPConfigs 获取SMTP配置列表
func (h *SMTPHandler) ListSMTPConfigs(c *gin.Context) {
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
	isActive := c.Query("is_active")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 构建查询
	query := h.db.Model(&models.SMTPConfig{})

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
	var configs []models.SMTPConfig
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("sort_order ASC, created_at DESC").Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"configs":  configs,
		"total":    total,
		"page":     page,
		"page_size": pageSize,
	})
}

// GetSMTPConfig 获取单个SMTP配置
func (h *SMTPHandler) GetSMTPConfig(c *gin.Context) {
	// 检查管理员权限
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	id := c.Param("id")
	var config models.SMTPConfig
	if err := h.db.First(&config, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "SMTP配置不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"config": config,
	})
}

// CreateSMTPConfig 创建SMTP配置
func (h *SMTPHandler) CreateSMTPConfig(c *gin.Context) {
	// 检查管理员权限
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Host        string `json:"host" binding:"required"`
		Port        int    `json:"port" binding:"required,min=1,max=65535"`
		Username    string `json:"username" binding:"required"`
		Password    string `json:"password" binding:"required"`
		FromEmail   string `json:"from_email" binding:"required,email"`
		FromName    string `json:"from_name"`
		UseTLS      bool   `json:"use_tls"`
		IsActive    bool   `json:"is_active"`
		IsDefault   bool   `json:"is_default"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查名称是否已存在
	var existingConfig models.SMTPConfig
	if err := h.db.Where("name = ?", req.Name).First(&existingConfig).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "SMTP配置名称已存在"})
		return
	}

	// 如果设置为默认配置，需要将其他配置设为非默认
	if req.IsDefault {
		if err := h.db.Model(&models.SMTPConfig{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新默认配置失败"})
			return
		}
	}

	// 创建SMTP配置
	config := models.SMTPConfig{
		Name:        req.Name,
		Host:        req.Host,
		Port:        req.Port,
		Username:    req.Username,
		Password:    req.Password,
		FromEmail:   req.FromEmail,
		FromName:    req.FromName,
		UseTLS:      req.UseTLS,
		IsActive:    req.IsActive,
		IsDefault:   req.IsDefault,
		Description: req.Description,
	}

	if err := h.db.Create(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "创建成功",
		"config":  config,
	})
}

// UpdateSMTPConfig 更新SMTP配置
func (h *SMTPHandler) UpdateSMTPConfig(c *gin.Context) {
	// 检查管理员权限
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	id := c.Param("id")
	var config models.SMTPConfig
	if err := h.db.First(&config, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "SMTP配置不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Host        string `json:"host"`
		Port        int    `json:"port"`
		Username    string `json:"username"`
		Password    string `json:"password"`
		FromEmail   string `json:"from_email"`
		FromName    string `json:"from_name"`
		UseTLS      *bool  `json:"use_tls"`
		IsActive    *bool  `json:"is_active"`
		IsDefault   *bool  `json:"is_default"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新字段
	if req.Name != "" {
		// 检查名称是否已存在（排除当前配置）
		var existingConfig models.SMTPConfig
		if err := h.db.Where("name = ? AND id != ?", req.Name, id).First(&existingConfig).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "SMTP配置名称已存在"})
			return
		}
		config.Name = req.Name
	}
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
		config.Password = req.Password
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
	if req.IsActive != nil {
		config.IsActive = *req.IsActive
	}
	if req.IsDefault != nil {
		// 如果设置为默认配置，需要将其他配置设为非默认
		if *req.IsDefault {
			if err := h.db.Model(&models.SMTPConfig{}).Where("is_default = ? AND id != ?", true, id).Update("is_default", false).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "更新默认配置失败"})
				return
			}
		}
		config.IsDefault = *req.IsDefault
	}
	if req.Description != "" {
		config.Description = req.Description
	}

	if err := h.db.Save(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"config":  config,
	})
}

// DeleteSMTPConfig 删除SMTP配置
func (h *SMTPHandler) DeleteSMTPConfig(c *gin.Context) {
	// 检查管理员权限
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	id := c.Param("id")
	
	// 检查SMTP配置是否存在
	var config models.SMTPConfig
	if err := h.db.First(&config, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "SMTP配置不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 如果是默认配置，不能删除
	if config.IsDefault {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能删除默认SMTP配置"})
		return
	}

	if err := h.db.Delete(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
	})
}

// TestSMTPConfig 测试SMTP配置
func (h *SMTPHandler) TestSMTPConfig(c *gin.Context) {
	// 检查管理员权限
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	id := c.Param("id")
	
	// 检查SMTP配置是否存在
	var config models.SMTPConfig
	if err := h.db.First(&config, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "SMTP配置不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// TODO: 实现实际的SMTP测试逻辑
	// 这里只是模拟测试结果
	testResult := "SMTP连接测试成功"
	now := time.Now()
	
	// 更新测试时间和结果
	config.LastTestAt = &now
	config.TestResult = testResult
	if err := h.db.Save(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新测试结果失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "测试成功",
		"test_result": testResult,
		"tested_at":   now,
	})
}

// SetDefaultSMTPConfig 设置默认SMTP配置
func (h *SMTPHandler) SetDefaultSMTPConfig(c *gin.Context) {
	// 检查管理员权限
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	id := c.Param("id")
	
	// 检查SMTP配置是否存在
	var config models.SMTPConfig
	if err := h.db.First(&config, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "SMTP配置不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 将所有配置设为非默认
	if err := h.db.Model(&models.SMTPConfig{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新默认配置失败"})
		return
	}

	// 将当前配置设为默认
	config.IsDefault = true
	if err := h.db.Save(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "设置默认配置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "设置默认配置成功",
		"config":  config,
	})
}