package middleware

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestResponseLogger 记录请求和响应的中间件
func RequestResponseLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		var statusColor, methodColor, resetColor string
		if param.IsOutputColor() {
			statusColor = param.StatusCodeColor()
			methodColor = param.MethodColor()
			resetColor = param.ResetColor()
		}

		return fmt.Sprintf("[%s] %s |%s %3d %s| %13v | %15s |%s %-7s %s %#v %s\n",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			statusColor, param.StatusCode, resetColor,
			param.Latency,
			param.ClientIP,
			methodColor, param.Method, resetColor,
			param.Path,
			param.ErrorMessage,
		)
	})
}

// ResponseTimeMiddleware 记录API响应时间
func ResponseTimeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// 记录慢查询（超过1秒）
		if duration > time.Second {
			log.Printf("SLOW_REQUEST: %s %s took %v (status: %d)", method, path, duration, statusCode)
		}

		// 记录错误请求
		if statusCode >= 400 {
			log.Printf("ERROR_REQUEST: %s %s returned %d in %v", method, path, statusCode, duration)
		}

		// 设置响应头
		c.Header("X-Response-Time", duration.String())
	}
}

// HealthCheckMiddleware 健康检查中间件
func HealthCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/api/health" {
			c.JSON(200, gin.H{
				"status":    "ok",
				"timestamp": time.Now().Unix(),
				"uptime":    time.Since(startTime).String(),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

var startTime = time.Now()

// RequestSizeMiddleware 限制请求体大小
func RequestSizeMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.JSON(413, gin.H{"error": "请求体过大"})
			c.Abort()
			return
		}

		// 限制请求体读取大小
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}

// ErrorRecoveryMiddleware 错误恢复中间件
func ErrorRecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err interface{}) {
		log.Printf("PANIC: %v", err)
		c.JSON(500, gin.H{
			"error": "服务器内部错误",
			"code":  "INTERNAL_SERVER_ERROR",
		})
	})
}