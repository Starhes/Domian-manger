package api

import (
	"domain-max/pkg/config"
	"domain-max/pkg/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRoutes 设置API路由
func SetupRoutes(router *gin.Engine, db *gorm.DB, cfg *config.Config) {
	// 创建API处理器
	authHandler := NewAuthHandler(db, cfg)
	dnsHandler := NewDNSHandler(db)
	domainHandler := NewDomainHandler(db)
	userHandler := NewUserHandler(db)
	smtpHandler := NewSMTPHandler(db)
	providerHandler := NewProviderHandler(db)

	// API路由组
	apiGroup := router.Group("/api")

	// 健康检查
	apiGroup.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 认证相关路由
	authGroup := apiGroup.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/forgot-password", authHandler.ForgotPassword)
		authGroup.POST("/reset-password", authHandler.ResetPassword)
	}

	// 需要认证的路由
	authRequiredGroup := apiGroup.Group("")
	authRequiredGroup.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		// 用户资料相关路由
		authRequiredGroup.GET("/profile", authHandler.GetProfile)
		authRequiredGroup.PUT("/profile", authHandler.UpdateProfile)
		authRequiredGroup.PUT("/change-password", authHandler.ChangePassword)
		authRequiredGroup.GET("/user/stats", userHandler.GetUserStats)

		// DNS记录相关路由
		authRequiredGroup.GET("/dns-records", dnsHandler.ListDNSRecords)
		authRequiredGroup.GET("/dns-records/:id", dnsHandler.GetDNSRecord)
		authRequiredGroup.POST("/dns-records", dnsHandler.CreateDNSRecord)
		authRequiredGroup.PUT("/dns-records/:id", dnsHandler.UpdateDNSRecord)
		authRequiredGroup.DELETE("/dns-records/:id", dnsHandler.DeleteDNSRecord)
		authRequiredGroup.POST("/dns-records/batch", dnsHandler.BatchCreateDNSRecords)
		authRequiredGroup.GET("/dns-records/export", dnsHandler.ExportDNSRecords)

		// 域名相关路由
		authRequiredGroup.GET("/domains", domainHandler.ListDomains)
		authRequiredGroup.GET("/domains/:id", domainHandler.GetDomain)
		authRequiredGroup.POST("/domains", domainHandler.CreateDomain)
		authRequiredGroup.PUT("/domains/:id", domainHandler.UpdateDomain)
		authRequiredGroup.DELETE("/domains/:id", domainHandler.DeleteDomain)
		authRequiredGroup.GET("/domains/:id/dns-records", domainHandler.GetDomainDNSRecords)
		authRequiredGroup.GET("/domains/:id/stats", domainHandler.GetDomainStats)

		// 需要管理员权限的路由
		adminGroup := authRequiredGroup.Group("")
		adminGroup.Use(middleware.AdminMiddleware())
		{
			// 用户管理路由
			adminGroup.GET("/users", userHandler.ListUsers)
			adminGroup.GET("/users/:id", userHandler.GetUser)
			adminGroup.POST("/users", userHandler.CreateUser)
			adminGroup.PUT("/users/:id", userHandler.UpdateUser)
			adminGroup.DELETE("/users/:id", userHandler.DeleteUser)
			adminGroup.POST("/users/:id/reset-password", userHandler.ResetUserPassword)
			adminGroup.GET("/system/stats", userHandler.GetSystemStats)

			// SMTP配置管理路由
			adminGroup.GET("/smtp-configs", smtpHandler.ListSMTPConfigs)
			adminGroup.GET("/smtp-configs/:id", smtpHandler.GetSMTPConfig)
			adminGroup.POST("/smtp-configs", smtpHandler.CreateSMTPConfig)
			adminGroup.PUT("/smtp-configs/:id", smtpHandler.UpdateSMTPConfig)
			adminGroup.DELETE("/smtp-configs/:id", smtpHandler.DeleteSMTPConfig)
			adminGroup.POST("/smtp-configs/:id/test", smtpHandler.TestSMTPConfig)
			adminGroup.PUT("/smtp-configs/:id/set-default", smtpHandler.SetDefaultSMTPConfig)

			// DNS提供商管理路由
			adminGroup.GET("/providers", providerHandler.ListProviders)
			adminGroup.GET("/providers/:id", providerHandler.GetProvider)
			adminGroup.POST("/providers", providerHandler.CreateProvider)
			adminGroup.PUT("/providers/:id", providerHandler.UpdateProvider)
			adminGroup.DELETE("/providers/:id", providerHandler.DeleteProvider)
			adminGroup.POST("/providers/:id/test", providerHandler.TestProvider)
			adminGroup.PUT("/providers/:id/toggle-status", providerHandler.ToggleProviderStatus)
			adminGroup.GET("/providers/types", providerHandler.GetProviderTypes)
		}
	}
}