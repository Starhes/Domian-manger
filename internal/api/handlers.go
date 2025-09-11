package api

import (
	"domain-manager/internal/config"
	"domain-manager/internal/middleware"
	"domain-manager/internal/models"
	"domain-manager/internal/services"
	"domain-manager/internal/utils"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ========================= 路由设置 =========================

func SetupRoutes(router *gin.RouterGroup, db *gorm.DB, cfg *config.Config) {
	// 创建服务实例
	authService := services.NewAuthService(db, cfg)
	dnsService := services.NewDNSService(db)
	adminService := services.NewAdminService(db, cfg)

	// 创建处理器实例
	authHandler := NewAuthHandler(authService)
	dnsHandler := NewDNSHandler(dnsService)
	adminHandler := NewAdminHandler(adminService)

	// 公开路由
	public := router.Group("/")
	{
		// 健康检查
		public.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"message": "服务运行正常",
			})
		})

	// 认证相关（添加速率限制）
	public.POST("/register", middleware.RegisterRateLimit(), authHandler.Register)
	public.POST("/login", middleware.LoginRateLimit(), authHandler.Login)
	public.GET("/verify-email/:token", authHandler.VerifyEmail)
	public.POST("/forgot-password", authHandler.ForgotPassword)
	public.POST("/reset-password", authHandler.ResetPassword)

	// 获取可用域名（无需认证）
	public.GET("/domains", dnsHandler.GetAvailableDomains)
	}

	// 需要认证的路由
	protected := router.Group("/")
	protected.Use(middleware.AuthRequired(db, cfg))
	{
		// 用户相关
		protected.GET("/profile", authHandler.GetProfile)
		protected.PUT("/profile", authHandler.UpdateProfile)

	// DNS记录管理（添加速率限制）
	protected.GET("/dns-records", dnsHandler.GetUserDNSRecords)
	protected.POST("/dns-records", middleware.DNSOperationRateLimit(), dnsHandler.CreateDNSRecord)
	protected.PUT("/dns-records/:id", middleware.DNSOperationRateLimit(), dnsHandler.UpdateDNSRecord)
	protected.DELETE("/dns-records/:id", middleware.DNSOperationRateLimit(), dnsHandler.DeleteDNSRecord)
	}

	// 需要认证且支持token撤销的路由
	protectedWithRevocation := router.Group("/")
	protectedWithRevocation.Use(middleware.AuthRequiredWithTokenManager(db, cfg, authService))
	{
		// 登出需要token撤销功能
		protectedWithRevocation.POST("/logout", authHandler.Logout)
	}

	// 管理员路由（添加速率限制）
	admin := router.Group("/admin")
	admin.Use(middleware.AdminRequired(db, cfg))
	admin.Use(middleware.AdminRateLimit())
	{
		// 用户管理
		admin.GET("/users", adminHandler.GetUsers)
		admin.GET("/users/:id", adminHandler.GetUser)
		admin.PUT("/users/:id", adminHandler.UpdateUser)
		admin.DELETE("/users/:id", adminHandler.DeleteUser)

		// 域名管理
		admin.GET("/domains", adminHandler.GetDomains)
		admin.POST("/domains", adminHandler.CreateDomain)
		admin.PUT("/domains/:id", adminHandler.UpdateDomain)
		admin.DELETE("/domains/:id", adminHandler.DeleteDomain)
		admin.POST("/domains/sync", adminHandler.SyncDomains)

		// DNS服务商管理
		admin.GET("/dns-providers", adminHandler.GetDNSProviders)
		admin.POST("/dns-providers", adminHandler.CreateDNSProvider)
		admin.PUT("/dns-providers/:id", adminHandler.UpdateDNSProvider)
		admin.DELETE("/dns-providers/:id", adminHandler.DeleteDNSProvider)

		// 系统统计
		admin.GET("/stats", adminHandler.GetStats)

		// SMTP配置管理
		admin.GET("/smtp-configs", adminHandler.GetSMTPConfigs)
		admin.GET("/smtp-configs/:id", adminHandler.GetSMTPConfig)
		admin.POST("/smtp-configs", adminHandler.CreateSMTPConfig)
		admin.PUT("/smtp-configs/:id", adminHandler.UpdateSMTPConfig)
		admin.DELETE("/smtp-configs/:id", adminHandler.DeleteSMTPConfig)
		admin.POST("/smtp-configs/:id/activate", adminHandler.ActivateSMTPConfig)
		admin.POST("/smtp-configs/:id/set-default", adminHandler.SetDefaultSMTPConfig)
		admin.POST("/smtp-configs/:id/test", adminHandler.TestSMTPConfig)
	}
}

// ========================= 认证处理器 =========================

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleValidationError(c, err)
		return
	}

	if err := h.authService.RegisterWithContext(c, req); err != nil {
		utils.HandleBadRequest(c, err.Error(), err)
		return
	}

	utils.SuccessWithMessage(c, "注册成功，请检查邮箱激活账户", nil)
}

// 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleValidationError(c, err)
		return
	}

	resp, err := h.authService.Login(req)
	if err != nil {
		utils.HandleUnauthorized(c, "邮箱或密码错误")
		utils.LogSensitiveOperation("LOGIN_FAILED", req.Email, "Invalid credentials", c)
		return
	}

	// 生成CSRF令牌
	csrfToken, err := generateCSRFToken()
	if err != nil {
		utils.HandleInternalError(c, err)
		return
	}

	// 获取Cookie配置
	cookieConfig := middleware.GetCookieConfig(
		c.GetHeader("X-Development") == "true", // 开发环境标识
		c.Request.Host,
	)

	// 设置认证Cookie
	middleware.SetAuthCookie(c, resp.Token, cookieConfig)
	if resp.RefreshToken != "" {
		middleware.SetRefreshCookie(c, resp.RefreshToken, cookieConfig)
	}
	middleware.SetCSRFCookie(c, csrfToken, cookieConfig)

	utils.LogSensitiveOperation("LOGIN_SUCCESS", req.Email, "User logged in", c)
	
	// 返回响应（不包含敏感令牌）
	utils.Success(c, gin.H{
		"user":       resp.User,
		"csrf_token": csrfToken, // 前端需要在请求头中包含此令牌
		"message":    "登录成功",
	})
}

// 邮箱验证
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		utils.HandleBadRequest(c, "验证令牌不能为空", nil)
		return
	}

	if err := h.authService.VerifyEmail(token); err != nil {
		utils.HandleBadRequest(c, "验证令牌无效或已过期", err)
		return
	}

	utils.LogSensitiveOperation("EMAIL_VERIFIED", "unknown", "Email verification successful", c)
	utils.SuccessWithMessage(c, "邮箱验证成功，账户已激活", nil)
}

// 忘记密码
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleValidationError(c, err)
		return
	}

	if err := h.authService.ForgotPassword(req); err != nil {
		utils.HandleInternalError(c, err)
		return
	}

	utils.LogSensitiveOperation("PASSWORD_RESET_REQUESTED", req.Email, "Password reset requested", c)
	utils.SuccessWithMessage(c, "如果邮箱存在，重置链接已发送", nil)
}

// 重置密码
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleValidationError(c, err)
		return
	}

	if err := h.authService.ResetPassword(req); err != nil {
		utils.HandleBadRequest(c, "密码重置失败，令牌可能无效或已过期", err)
		return
	}

	utils.LogSensitiveOperation("PASSWORD_RESET_SUCCESS", "unknown", "Password reset successful", c)
	utils.SuccessWithMessage(c, "密码重置成功", nil)
}

// 获取用户资料
func (h *AuthHandler) GetProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		utils.HandleUnauthorized(c, "用户信息不存在")
		return
	}

	utils.Success(c, gin.H{"user": user.(models.User)})
}

// 更新用户资料
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		utils.HandleUnauthorized(c, "用户信息不存在")
		return
	}

	userObj := user.(models.User)

	var req struct {
		Email    string `json:"email" binding:"omitempty,email"`
		Password string `json:"password" binding:"omitempty,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleValidationError(c, err)
		return
	}

	// 调用服务层更新用户资料
	updatedUser, err := h.authService.UpdateProfile(userObj.ID, req.Email, req.Password)
	if err != nil {
		utils.HandleBadRequest(c, err.Error(), err)
		return
	}

	utils.LogSensitiveOperation("PROFILE_UPDATED", string(rune(userObj.ID)), "Profile updated", c)
	utils.SuccessWithMessage(c, "资料更新成功", gin.H{"user": updatedUser})
}

// 用户登出
func (h *AuthHandler) Logout(c *gin.Context) {
	// 从上下文获取token
	tokenString, exists := c.Get("token")
	if !exists {
		utils.HandleBadRequest(c, "无法获取令牌", nil)
		return
	}

	// 获取用户信息用于日志记录
	user, _ := c.Get("user")
	userID := "unknown"
	if user != nil {
		userObj := user.(models.User)
		userID = string(rune(userObj.ID))
	}

	// 撤销token
	if err := h.authService.Logout(tokenString.(string)); err != nil {
		utils.HandleInternalError(c, err)
		return
	}

	// 获取Cookie配置并清除Cookie
	cookieConfig := middleware.GetCookieConfig(
		c.GetHeader("X-Development") == "true",
		c.Request.Host,
	)
	middleware.ClearAuthCookies(c, cookieConfig)

	utils.LogSensitiveOperation("LOGOUT_SUCCESS", userID, "User logged out", c)
	utils.SuccessWithMessage(c, "登出成功", nil)
}

// generateCSRFToken 生成CSRF令牌
func generateCSRFToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ========================= DNS处理器 =========================

type DNSHandler struct {
	dnsService *services.DNSService
}

func NewDNSHandler(dnsService *services.DNSService) *DNSHandler {
	return &DNSHandler{dnsService: dnsService}
}

// 获取用户的DNS记录
func (h *DNSHandler) GetUserDNSRecords(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户信息不存在"})
		return
	}

	userObj := user.(models.User)
	records, err := h.dnsService.GetUserDNSRecords(userObj.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取DNS记录失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"records": records,
	})
}

// 创建DNS记录
func (h *DNSHandler) CreateDNSRecord(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户信息不存在"})
		return
	}

	userObj := user.(models.User)

	var req models.CreateDNSRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	record, err := h.dnsService.CreateDNSRecord(userObj.ID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "DNS记录创建成功",
		"record":  record,
	})
}

// 更新DNS记录
func (h *DNSHandler) UpdateDNSRecord(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户信息不存在"})
		return
	}

	userObj := user.(models.User)

	recordIDStr := c.Param("id")
	recordID, err := strconv.ParseUint(recordIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的记录ID"})
		return
	}

	var req models.UpdateDNSRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	record, err := h.dnsService.UpdateDNSRecord(userObj.ID, uint(recordID), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "DNS记录更新成功",
		"record":  record,
	})
}

// 删除DNS记录
func (h *DNSHandler) DeleteDNSRecord(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户信息不存在"})
		return
	}

	userObj := user.(models.User)

	recordIDStr := c.Param("id")
	recordID, err := strconv.ParseUint(recordIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的记录ID"})
		return
	}

	if err := h.dnsService.DeleteDNSRecord(userObj.ID, uint(recordID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "DNS记录删除成功",
	})
}

// 获取可用域名列表
func (h *DNSHandler) GetAvailableDomains(c *gin.Context) {
	domains, err := h.dnsService.GetAvailableDomains()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取域名列表失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"domains": domains,
	})
}

// ========================= 管理员处理器 =========================

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

	// 验证分页参数
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		utils.HandleBadRequest(c, "页码格式不正确", err)
		return
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		utils.HandleBadRequest(c, "每页大小格式不正确", err)
		return
	}

	page, pageSize, err = utils.ValidatePageParams(page, pageSize)
	if err != nil {
		utils.HandleBadRequest(c, err.Error(), err)
		return
	}

	// 验证搜索查询
	if search != "" {
		if err := utils.ValidateSearchQuery(search); err != nil {
			utils.HandleBadRequest(c, err.Error(), err)
			return
		}
	}

	users, total, err := h.adminService.GetUsers(page, pageSize, search)
	if err != nil {
		utils.HandleInternalError(c, err)
		return
	}

	utils.Success(c, gin.H{
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
	
	// 使用增强的ID验证
	userIDUint64, err := utils.ValidateUserID(userIDStr)
	if err != nil {
		utils.HandleBadRequest(c, err.Error(), err)
		return
	}
	userID := uint(userIDUint64)

	user, err := h.adminService.GetUser(uint(userID))
	if err != nil {
		utils.HandleNotFound(c, "用户")
		return
	}

	utils.Success(c, gin.H{"user": user})
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

// SyncDomains 同步域名
func (h *AdminHandler) SyncDomains(c *gin.Context) {
	err := h.adminService.SyncDomains()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "同步域名失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "域名同步成功"})
}

// ========================= SMTP配置管理 =========================

// 获取所有SMTP配置
func (h *AdminHandler) GetSMTPConfigs(c *gin.Context) {
	configs, err := h.adminService.GetSMTPConfigs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取SMTP配置列表失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"configs": configs})
}

// 获取单个SMTP配置
func (h *AdminHandler) GetSMTPConfig(c *gin.Context) {
	configIDStr := c.Param("id")
	configID, err := strconv.ParseUint(configIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的配置ID"})
		return
	}

	config, err := h.adminService.GetSMTPConfig(uint(configID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"config": config})
}

// 创建SMTP配置
func (h *AdminHandler) CreateSMTPConfig(c *gin.Context) {
	var req models.CreateSMTPConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	config, err := h.adminService.CreateSMTPConfig(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "SMTP配置创建成功",
		"config":  config,
	})
}

// 更新SMTP配置
func (h *AdminHandler) UpdateSMTPConfig(c *gin.Context) {
	configIDStr := c.Param("id")
	configID, err := strconv.ParseUint(configIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的配置ID"})
		return
	}

	var req models.UpdateSMTPConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	config, err := h.adminService.UpdateSMTPConfig(uint(configID), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "SMTP配置更新成功",
		"config":  config,
	})
}

// 删除SMTP配置
func (h *AdminHandler) DeleteSMTPConfig(c *gin.Context) {
	configIDStr := c.Param("id")
	configID, err := strconv.ParseUint(configIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的配置ID"})
		return
	}

	if err := h.adminService.DeleteSMTPConfig(uint(configID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "SMTP配置删除成功"})
}

// 激活SMTP配置
func (h *AdminHandler) ActivateSMTPConfig(c *gin.Context) {
	configIDStr := c.Param("id")
	configID, err := strconv.ParseUint(configIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的配置ID"})
		return
	}

	if err := h.adminService.ActivateSMTPConfig(uint(configID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "SMTP配置激活成功"})
}

// 设置默认SMTP配置
func (h *AdminHandler) SetDefaultSMTPConfig(c *gin.Context) {
	configIDStr := c.Param("id")
	configID, err := strconv.ParseUint(configIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的配置ID"})
		return
	}

	if err := h.adminService.SetDefaultSMTPConfig(uint(configID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "默认SMTP配置设置成功"})
}

// 测试SMTP配置
func (h *AdminHandler) TestSMTPConfig(c *gin.Context) {
	configIDStr := c.Param("id")
	configID, err := strconv.ParseUint(configIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的配置ID"})
		return
	}

	var req models.TestSMTPConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	if err := h.adminService.TestSMTPConfig(uint(configID), req.ToEmail); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SMTP测试失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "SMTP测试邮件发送成功"})
}
