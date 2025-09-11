package middleware

import (
	"domain-manager/internal/constants"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitEntry 速率限制条目
type RateLimitEntry struct {
	Count      int
	ResetTime  time.Time
	LastAccess time.Time
}

// RateLimiter 速率限制器
type RateLimiter struct {
	entries map[string]*RateLimitEntry
	mu      sync.RWMutex
}

// NewRateLimiter 创建新的速率限制器
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		entries: make(map[string]*RateLimitEntry),
	}
	
	// 启动清理goroutine
	go rl.cleanup()
	
	return rl
}

// cleanup 定期清理过期条目
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			for key, entry := range rl.entries {
				// 清理超过1小时未访问的条目
				if now.Sub(entry.LastAccess) > time.Hour {
					delete(rl.entries, key)
				}
			}
			rl.mu.Unlock()
		}
	}
}

// checkLimit 检查是否超过速率限制
func (rl *RateLimiter) checkLimit(key string, limit int, window time.Duration) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	now := time.Now()
	entry, exists := rl.entries[key]
	
	if !exists {
		// 创建新条目
		rl.entries[key] = &RateLimitEntry{
			Count:      1,
			ResetTime:  now.Add(window),
			LastAccess: now,
		}
		return true
	}
	
	entry.LastAccess = now
	
	// 检查是否需要重置计数器
	if now.After(entry.ResetTime) {
		entry.Count = 1
		entry.ResetTime = now.Add(window)
		return true
	}
	
	// 检查是否超过限制
	if entry.Count >= limit {
		return false
	}
	
	entry.Count++
	return true
}

// getRemainingTime 获取剩余重置时间
func (rl *RateLimiter) getRemainingTime(key string) time.Duration {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	
	entry, exists := rl.entries[key]
	if !exists {
		return 0
	}
	
	remaining := entry.ResetTime.Sub(time.Now())
	if remaining < 0 {
		return 0
	}
	
	return remaining
}

// 全局速率限制器实例
var (
	globalRateLimiter = NewRateLimiter()
)

// getClientIP 获取客户端IP地址
func getClientIP(c *gin.Context) string {
	// 检查 X-Forwarded-For 头
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		// 取第一个IP地址
		if idx := 0; idx < len(xff); idx++ {
			if xff[idx] == ',' {
				return xff[:idx]
			}
		}
		return xff
	}
	
	// 检查 X-Real-IP 头
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}
	
	// 使用 RemoteAddr
	return c.ClientIP()
}

// APIRateLimit API通用速率限制中间件
func APIRateLimit() gin.HandlerFunc {
	return RateLimit(constants.APIRateLimit, constants.APIRateWindow, "api")
}

// LoginRateLimit 登录速率限制中间件
func LoginRateLimit() gin.HandlerFunc {
	return RateLimit(constants.LoginRateLimit, constants.LoginRateWindow, "login")
}

// RegisterRateLimit 注册速率限制中间件
func RegisterRateLimit() gin.HandlerFunc {
	return RateLimit(constants.RegisterRateLimit, constants.RegisterRateWindow, "register")
}

// DNSOperationRateLimit DNS操作速率限制中间件
func DNSOperationRateLimit() gin.HandlerFunc {
	return RateLimit(constants.DNSOperationRateLimit, constants.DNSOperationRateWindow, "dns")
}

// RateLimit 通用速率限制中间件
func RateLimit(limit int, window time.Duration, prefix string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		clientIP := getClientIP(c)
		key := prefix + ":" + clientIP
		
		if !globalRateLimiter.checkLimit(key, limit, window) {
			remainingTime := globalRateLimiter.getRemainingTime(key)
			
			// 设置速率限制响应头
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(remainingTime).Unix()))
			c.Header("Retry-After", fmt.Sprintf("%.0f", remainingTime.Seconds()))
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   constants.ErrMsgTooManyRequests,
				"code":    "TOO_MANY_REQUESTS",
				"retry_after": int(remainingTime.Seconds()),
			})
			c.Abort()
			return
		}
		
		c.Next()
	})
}

// PerUserRateLimit 基于用户的速率限制中间件
func PerUserRateLimit(limit int, window time.Duration, prefix string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 尝试从上下文获取用户信息
		var userKey string
		if user, exists := c.Get("user"); exists {
			if userObj, ok := user.(interface{ GetID() uint }); ok {
				userKey = fmt.Sprintf("%s:user:%d", prefix, userObj.GetID())
			}
		}
		
		// 如果没有用户信息，使用IP作为key
		if userKey == "" {
			clientIP := getClientIP(c)
			userKey = fmt.Sprintf("%s:ip:%s", prefix, clientIP)
		}
		
		if !globalRateLimiter.checkLimit(userKey, limit, window) {
			remainingTime := globalRateLimiter.getRemainingTime(userKey)
			
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(remainingTime).Unix()))
			c.Header("Retry-After", fmt.Sprintf("%.0f", remainingTime.Seconds()))
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       constants.ErrMsgTooManyRequests,
				"code":        "TOO_MANY_REQUESTS",
				"retry_after": int(remainingTime.Seconds()),
			})
			c.Abort()
			return
		}
		
		c.Next()
	})
}

// AdminRateLimit 管理员操作速率限制
func AdminRateLimit() gin.HandlerFunc {
	// 管理员操作允许更高的速率限制
	return RateLimit(200, 1*time.Minute, "admin")
}

// GetRateLimitStatus 获取速率限制状态（用于调试）
func GetRateLimitStatus() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		clientIP := getClientIP(c)
		
		status := make(map[string]interface{})
		globalRateLimiter.mu.RLock()
		for key, entry := range globalRateLimiter.entries {
			if clientIP != "" && key != "" {
				status[key] = gin.H{
					"count":       entry.Count,
					"reset_time":  entry.ResetTime.Unix(),
					"last_access": entry.LastAccess.Unix(),
				}
			}
		}
		globalRateLimiter.mu.RUnlock()
		
		c.JSON(http.StatusOK, gin.H{
			"client_ip": clientIP,
			"limits":    status,
		})
	})
}
