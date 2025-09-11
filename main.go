package main

import (
	"domain-manager/internal/api"
	"domain-manager/internal/config"
	"domain-manager/internal/database"
	"domain-manager/internal/middleware"
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed frontend/dist/*
var frontendFiles embed.FS

func main() {
	// 加载配置
	cfg := config.Load()

	// 连接数据库
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// 自动迁移数据库表
	if err := database.Migrate(db); err != nil {
		log.Fatal("数据库迁移失败:", err)
	}

	// 设置Gin模式
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// 添加CORS中间件
	router.Use(middleware.CORS())

	// API路由
	apiGroup := router.Group("/api")
	api.SetupRoutes(apiGroup, db, cfg)

	// 前端静态文件服务
	setupFrontendRoutes(router)

	log.Printf("服务器启动在端口 %s", cfg.Port)
	log.Fatal(router.Run(":" + cfg.Port))
}

func setupFrontendRoutes(router *gin.Engine) {
	// 获取嵌入的前端文件系统
	frontendFS, err := fs.Sub(frontendFiles, "frontend/dist")
	if err != nil {
		log.Fatal("无法加载前端文件:", err)
	}

	// 静态文件服务 - 处理构建后的静态资源
	staticFS, err := fs.Sub(frontendFS, "static")
	if err != nil {
		log.Fatal("无法加载静态文件:", err)
	}
	router.StaticFS("/static", http.FS(staticFS))

	// 处理所有其他路由，返回index.html (用于React Router)
	router.NoRoute(func(c *gin.Context) {
		// 如果是API请求，返回404
		if len(c.Request.URL.Path) > 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(404, gin.H{"error": "API endpoint not found"})
			return
		}

		// 尝试读取index.html
		indexHTML, err := frontendFiles.ReadFile("frontend/dist/index.html")
		if err != nil {
			c.String(500, "无法加载前端页面")
			return
		}

		c.Data(200, "text/html; charset=utf-8", indexHTML)
	})
}
