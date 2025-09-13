package api

import (
	authmodels "domain-max/pkg/auth/models"
	dnsmodels "domain-max/pkg/dns/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserHandler 用户处理器
type UserHandler struct {
	db *gorm.DB
}

// NewUserHandler 创建新的用户处理器
func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

// ListUsers 获取用户列表（管理员功能）
func (h *UserHandler) ListUsers(c *gin.Context) {
	// 检查管理员权限
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	email := c.Query("email")
	nickname := c.Query("nickname")
	isActive := c.Query("is_active")
	isAdminFilter := c.Query("is_admin")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 构建查询
	query := h.db.Model(&authmodels.User{})

	if email != "" {
		query = query.Where("email LIKE ?", "%"+email+"%")
	}
	if nickname != "" {
		query = query.Where("nickname LIKE ?", "%"+nickname+"%")
	}
	if isActive != "" {
		query = query.Where("is_active = ?", isActive == "true")
	}
	if isAdminFilter != "" {
		query = query.Where("is_admin = ?", isAdminFilter == "true")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 分页查询
	var users []authmodels.User
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users":    users,
		"total":    total,
		"page":     page,
		"page_size": pageSize,
	})
}

// GetUser 获取单个用户信息（管理员功能）
func (h *UserHandler) GetUser(c *gin.Context) {
	// 检查管理员权限
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	id := c.Param("id")
	var user authmodels.User
	if err := h.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// CreateUser 创建用户（管理员功能）
func (h *UserHandler) CreateUser(c *gin.Context) {
	// 检查管理员权限
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	var req struct {
		Email           string `json:"email" binding:"required,email"`
		Password        string `json:"password" binding:"required,min=8,max=100"`
		Nickname       string `json:"nickname"`
		IsActive       bool   `json:"is_active"`
		IsAdmin        bool   `json:"is_admin"`
		DNSRecordQuota int    `json:"dns_record_quota"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证邮箱
	if err := authmodels.ValidateEmail(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证密码
	if err := authmodels.ValidatePassword(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证昵称
	if err := authmodels.ValidateNickname(req.Nickname); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查邮箱是否已存在
	var existingUser authmodels.User
	if err := h.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "邮箱已被注册"})
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// 创建用户
	user := authmodels.User{
		Email:          req.Email,
		Password:       string(hashedPassword),
		Nickname:       req.Nickname,
		IsActive:       req.IsActive,
		IsAdmin:        req.IsAdmin,
		DNSRecordQuota: req.DNSRecordQuota,
		Status:         "normal",
	}

	if user.DNSRecordQuota == 0 {
		user.DNSRecordQuota = 10 // 默认配额
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "创建成功",
		"user":    user,
	})
}

// UpdateUser 更新用户信息（管理员功能）
func (h *UserHandler) UpdateUser(c *gin.Context) {
	// 检查管理员权限
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	id := c.Param("id")
	var user authmodels.User
	if err := h.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	var req struct {
		Email           string `json:"email"`
		Nickname       string `json:"nickname"`
		IsActive       *bool  `json:"is_active"`
		IsAdmin        *bool  `json:"is_admin"`
		DNSRecordQuota int    `json:"dns_record_quota"`
		Status         string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新字段
	if req.Email != "" {
		// 验证邮箱
		if err := authmodels.ValidateEmail(req.Email); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// 检查邮箱是否已存在（排除当前用户）
		var existingUser authmodels.User
		if err := h.db.Where("email = ? AND id != ?", req.Email, id).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "邮箱已被注册"})
			return
		}
		user.Email = req.Email
	}
	if req.Nickname != "" {
		// 验证昵称
		if err := authmodels.ValidateNickname(req.Nickname); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		user.Nickname = req.Nickname
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	if req.IsAdmin != nil {
		user.IsAdmin = *req.IsAdmin
	}
	if req.DNSRecordQuota > 0 {
		user.DNSRecordQuota = req.DNSRecordQuota
	}
	if req.Status != "" {
		validStatuses := []string{"normal", "suspended", "banned"}
		valid := false
		for _, status := range validStatuses {
			if req.Status == status {
				valid = true
				break
			}
		}
		if !valid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户状态"})
			return
		}
		user.Status = req.Status
	}

	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"user":    user,
	})
}

// DeleteUser 删除用户（管理员功能）
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// 检查管理员权限
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	id := c.Param("id")
	
	// 检查用户是否存在
	var user authmodels.User
	if err := h.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 不能删除自己
	currentUserID, exists := c.Get("user_id")
	if exists && currentUserID.(uint) == user.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能删除自己的账户"})
		return
	}

	// 检查是否有关联的DNS记录
	var count int64
	if err := h.db.Model(&dnsmodels.DNSRecord{}).Where("user_id = ?", id).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该用户下还有DNS记录，无法删除"})
		return
	}

	if err := h.db.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
	})
}

// ResetUserPassword 重置用户密码（管理员功能）
func (h *UserHandler) ResetUserPassword(c *gin.Context) {
	// 检查管理员权限
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	id := c.Param("id")
	
	// 检查用户是否存在
	var user authmodels.User
	if err := h.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	var req struct {
		NewPassword string `json:"new_password" binding:"required,min=8,max=100"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证密码
	if err := authmodels.ValidatePassword(req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// 更新密码
	if err := h.db.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码重置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "密码重置成功",
	})
}

// GetUserStats 获取用户统计信息
func (h *UserHandler) GetUserStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	// 获取用户信息
	var user authmodels.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 统计DNS记录数量
	var dnsRecordCount int64
	if err := h.db.Model(&dnsmodels.DNSRecord{}).Where("user_id = ?", userID).Count(&dnsRecordCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 统计各种类型的DNS记录数量
	var typeStats []struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}

	if err := h.db.Model(&dnsmodels.DNSRecord{}).
		Where("user_id = ?", userID).
		Select("type, count(*) as count").
		Group("type").
		Find(&typeStats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":            user,
		"dns_record_count": dnsRecordCount,
		"type_stats":       typeStats,
		"quota_usage": gin.H{
			"used":  dnsRecordCount,
			"total": user.DNSRecordQuota,
			"percentage": float64(dnsRecordCount) / float64(user.DNSRecordQuota) * 100,
		},
	})
}

// GetSystemStats 获取系统统计信息（管理员功能）
func (h *UserHandler) GetSystemStats(c *gin.Context) {
	// 检查管理员权限
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	// 统计用户数量
	var userCount int64
	var activeUserCount int64
	var adminCount int64

	if err := h.db.Model(&authmodels.User{}).Count(&userCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	if err := h.db.Model(&authmodels.User{}).Where("is_active = ?", true).Count(&activeUserCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	if err := h.db.Model(&authmodels.User{}).Where("is_admin = ?", true).Count(&adminCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 统计域名数量
	var domainCount int64
	var activeDomainCount int64

	if err := h.db.Model(&dnsmodels.Domain{}).Count(&domainCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	if err := h.db.Model(&dnsmodels.Domain{}).Where("is_active = ?", true).Count(&activeDomainCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 统计DNS记录数量
	var dnsRecordCount int64

	if err := h.db.Model(&dnsmodels.DNSRecord{}).Count(&dnsRecordCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 统计DNS提供商数量
	var providerCount int64
	var activeProviderCount int64

	if err := h.db.Model(&dnsmodels.DNSProvider{}).Count(&providerCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	if err := h.db.Model(&dnsmodels.DNSProvider{}).Where("is_active = ?", true).Count(&activeProviderCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": gin.H{
			"total":  userCount,
			"active": activeUserCount,
			"admins": adminCount,
		},
		"domains": gin.H{
			"total":  domainCount,
			"active": activeDomainCount,
		},
		"dns_records": dnsRecordCount,
		"providers": gin.H{
			"total":  providerCount,
			"active": activeProviderCount,
		},
	})
}