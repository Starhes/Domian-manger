package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter 速率限制器
type RateLimiter struct {
	visitors map[string]*Visitor
	mutex    sync.RWMutex
	rate     time.Duration
	capacity int
}

// Visitor 访问者信息
type Visitor struct {
	limiter  chan struct{}
	lastSeen time.Time
}

// NewRateLimiter 创建新的速率限制器
func NewRateLimiter(rate time.Duration, capacity int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
		rate:     rate,
		capacity: capacity,
	}
	
	// 启动清理goroutine
	go rl.cleanupVisitors()
	
	return rl
}

// RateLimitMiddleware 速率限制中间件
func (rl *RateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		
		if !rl.allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// allow 检查是否允许请求
func (rl *RateLimiter) allow(ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	visitor, exists := rl.visitors[ip]
	if !exists {
		visitor = &Visitor{
			limiter:  make(chan struct{}, rl.capacity),
			lastSeen: time.Now(),
		}
		rl.visitors[ip] = visitor
		
		// 启动令牌补充goroutine
		go rl.refillTokens(visitor)
	}
	
	visitor.lastSeen = time.Now()
	
	select {
	case visitor.limiter <- struct{}{}:
		return true
	default:
		return false
	}
}

// refillTokens 补充令牌
func (rl *RateLimiter) refillTokens(visitor *Visitor) {
	ticker := time.NewTicker(rl.rate)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			select {
			case <-visitor.limiter:
				// 移除一个令牌（实际上是释放一个位置）
			default:
				// 令牌桶已满
			}
		}
	}
}

// cleanupVisitors 清理过期访问者
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			rl.mutex.Lock()
			for ip, visitor := range rl.visitors {
				if time.Since(visitor.lastSeen) > time.Hour {
					delete(rl.visitors, ip)
				}
			}
			rl.mutex.Unlock()
		}
	}
}