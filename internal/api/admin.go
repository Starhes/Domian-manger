package api

import (
	"domain-manager/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminService *services.AdminService
}

func NewAdminHandler(adminService *services.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

// 获取用户列表
func (h *AdminHandler) GetUsers(c *gin.Context) {
	// 获取查询参数
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")
	search := c.Query("search")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	users, total, err := h.adminService.GetUsers(page, pageSize, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"pagination": gin.H{
			"page":     page,
			"pageSize": pageSize,
			"total":    total,
		},
	})
}

// 获取单个用户
func (h *AdminHandler) GetUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	user, err := h.adminService.GetUser(uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// 更新用户
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	user, err := h.adminService.UpdateUser(uint(userID), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "用户更新成功",
		"user":    user,
	})
}

// 删除用户
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	if err := h.adminService.DeleteUser(uint(userID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户删除成功"})
}

// 获取域名列表
func (h *AdminHandler) GetDomains(c *gin.Context) {
	domains, err := h.adminService.GetDomains()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取域名列表失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"domains": domains})
}

// 创建域名
func (h *AdminHandler) CreateDomain(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	domain, err := h.adminService.CreateDomain(req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "域名创建成功",
		"domain":  domain,
	})
}

// 更新域名
func (h *AdminHandler) UpdateDomain(c *gin.Context) {
	domainIDStr := c.Param("id")
	domainID, err := strconv.ParseUint(domainIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的域名ID"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	domain, err := h.adminService.UpdateDomain(uint(domainID), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "域名更新成功",
		"domain":  domain,
	})
}

// 删除域名
func (h *AdminHandler) DeleteDomain(c *gin.Context) {
	domainIDStr := c.Param("id")
	domainID, err := strconv.ParseUint(domainIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的域名ID"})
		return
	}

	if err := h.adminService.DeleteDomain(uint(domainID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "域名删除成功"})
}

// 获取DNS服务商列表
func (h *AdminHandler) GetDNSProviders(c *gin.Context) {
	providers, err := h.adminService.GetDNSProviders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取DNS服务商列表失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

// 创建DNS服务商
func (h *AdminHandler) CreateDNSProvider(c *gin.Context) {
	var req struct {
		Name   string `json:"name" binding:"required"`
		Type   string `json:"type" binding:"required"`
		Config string `json:"config" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	provider, err := h.adminService.CreateDNSProvider(req.Name, req.Type, req.Config)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "DNS服务商创建成功",
		"provider": provider,
	})
}

// 更新DNS服务商
func (h *AdminHandler) UpdateDNSProvider(c *gin.Context) {
	providerIDStr := c.Param("id")
	providerID, err := strconv.ParseUint(providerIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的服务商ID"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	provider, err := h.adminService.UpdateDNSProvider(uint(providerID), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "DNS服务商更新成功",
		"provider": provider,
	})
}

// 删除DNS服务商
func (h *AdminHandler) DeleteDNSProvider(c *gin.Context) {
	providerIDStr := c.Param("id")
	providerID, err := strconv.ParseUint(providerIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的服务商ID"})
		return
	}

	if err := h.adminService.DeleteDNSProvider(uint(providerID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "DNS服务商删除成功"})
}

// 获取系统统计
func (h *AdminHandler) GetStats(c *gin.Context) {
	stats, err := h.adminService.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计信息失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}
