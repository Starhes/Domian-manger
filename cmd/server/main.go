package main

import (
	"domain-max/pkg/api"
	"domain-max/pkg/config"
	"domain-max/pkg/database"
	"domain-max/pkg/middleware"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 连接数据库
	var db *gorm.DB
	var err error
	
	if cfg.Environment == "development" {
		log.Println("开发环境：跳过数据库连接")
		db = nil
	} else {
		db, err = database.Connect(cfg)
		if err != nil {
			log.Fatal("数据库连接失败:", err)
		}

		// 自动迁移数据库表
		if err := database.Migrate(db); err != nil {
			log.Fatal("数据库迁移失败:", err)
		}
	}

	// 设置Gin模式
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// 添加环境感知的CORS中间件
	corsConfig := middleware.CORSConfig{
		AllowedOrigins: []string{
			"http://localhost:3000",   // React开发服务器
			"http://localhost:5173",   // Vite开发服务器
			"http://localhost:8080",   // 本地生产环境
			"https://your-domain.com", // 生产域名，需要替换为实际域名
		},
		IsDevelopment: cfg.Environment == "development",
	}
	router.Use(middleware.CORSWithConfig(corsConfig))

	// API路由
	setupAPIRoutes(router, db, cfg)

	// 前端静态文件服务
	setupWebRoutes(router)

	log.Printf("服务器启动在端口 %s", cfg.Port)
	log.Fatal(router.Run(":" + cfg.Port))
}

func setupWebRoutes(router *gin.Engine) {
	// 检查web/dist目录是否存在
	webDistPath := "web/dist"
	if _, err := os.Stat(webDistPath); os.IsNotExist(err) {
		log.Printf("警告: web/dist 目录不存在，跳过静态文件服务")
		return
	}

	// 静态文件服务 - 处理构建后的静态资源
	router.Static("/static", filepath.Join(webDistPath, "static"))

	// 处理所有其他路由，返回index.html (用于React Router)
	router.NoRoute(func(c *gin.Context) {
		// 如果是API请求，返回404
		if len(c.Request.URL.Path) > 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(404, gin.H{"error": "API endpoint not found"})
			return
		}

		// 尝试读取index.html
		indexPath := filepath.Join(webDistPath, "index.html")
		indexHTML, err := os.ReadFile(indexPath)
		if err != nil {
			c.String(500, "无法加载前端页面")
			return
		}

		c.Data(200, "text/html; charset=utf-8", indexHTML)
	})
}

func setupAPIRoutes(router *gin.Engine, db *gorm.DB, cfg *config.Config) {
	// 设置API路由
	api.SetupRoutes(router, db, cfg)
}