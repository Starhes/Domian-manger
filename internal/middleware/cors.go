package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORSConfig CORS配置结构
type CORSConfig struct {
	AllowedOrigins []string
	IsDevelopment  bool
}

func CORS() gin.HandlerFunc {
	return CORSWithConfig(CORSConfig{
		AllowedOrigins: []string{
			"http://localhost:3000",   // 开发环境前端
			"http://localhost:5173",   // Vite开发服务器
			"http://localhost:8080",   // 生产环境
			"https://your-domain.com", // 生产域名，需要替换为实际域名
		},
		IsDevelopment: false,
	})
}

func CORSWithConfig(config CORSConfig) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 检查请求的origin是否在允许列表中
		isAllowed := false
		for _, allowedOrigin := range config.AllowedOrigins {
			if origin == allowedOrigin {
				isAllowed = true
				break
			}
		}

		// 只为允许的origin设置CORS头
		if isAllowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		} else if origin == "" {
			// 对于没有Origin头的请求（如直接访问、Postman测试等）
			// 在开发环境允许，生产环境拒绝
			if config.IsDevelopment {
				c.Header("Access-Control-Allow-Origin", "*")
			} else {
				// 生产环境中，没有Origin的请求被拒绝CORS
				// 但允许同源请求继续处理
			}
		} else {
			// 不在允许列表中的origin，明确拒绝
			c.Header("Access-Control-Allow-Origin", "null")
		}

		// 设置标准CORS头
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		c.Header("Access-Control-Max-Age", "86400") // 预检请求缓存24小时

		// 添加安全响应头
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'")

		// 处理OPTIONS预检请求
		if c.Request.Method == "OPTIONS" {
			// 只有允许的origin才能进行预检
			if isAllowed {
				c.AbortWithStatus(204)
			} else {
				c.AbortWithStatus(403)
			}
			return
		}

		c.Next()
	})
}
