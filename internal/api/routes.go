package api

import (
	"domain-manager/internal/config"
	"domain-manager/internal/middleware"
	"domain-manager/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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
