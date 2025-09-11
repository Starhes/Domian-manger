package middleware

import (
	"domain-manager/internal/config"
	"domain-manager/internal/constants"
	"domain-manager/internal/models"
	"domain-manager/internal/services"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// ========================= 认证中间件 =========================

// AuthRequiredWithTokenManager 需要认证的中间件（支持token撤销检查）
func AuthRequiredWithTokenManager(db *gorm.DB, cfg *config.Config, authService *services.AuthService) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "需要授权令牌"})
			c.Abort()
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 验证签名算法，防止算法替换攻击
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的授权令牌"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的令牌声明"})
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的用户ID"})
			c.Abort()
			return
		}

		// 获取token的签发时间
		var issuedAt time.Time
		if iat, exists := claims["iat"]; exists {
			if iatFloat, ok := iat.(float64); ok {
				issuedAt = time.Unix(int64(iatFloat), 0)
			}
		}

		// 检查token是否被撤销
		if authService != nil {
			if err := authService.ValidateTokenRevocation(tokenString, uint(userID), issuedAt); err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				c.Abort()
				return
			}
		}

		var user models.User
		if err := db.First(&user, uint(userID)).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
			c.Abort()
			return
		}

		if !user.IsActive {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户账户未激活"})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Set("token", tokenString) // 保存token用于logout等操作
		c.Next()
	})
}

func AuthRequired(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		var tokenString string
		
		// 优先从Cookie获取token
		if cookieToken, err := GetAuthToken(c); err == nil && cookieToken != "" {
			tokenString = cookieToken
		} else {
			// 回退到Authorization头
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "需要授权令牌"})
				c.Abort()
				return
			}
			tokenString = strings.Replace(authHeader, "Bearer ", "", 1)
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 验证签名算法，防止算法替换攻击
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的授权令牌"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的令牌声明"})
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的用户ID"})
			c.Abort()
			return
		}

		var user models.User
		if err := db.First(&user, uint(userID)).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
			c.Abort()
			return
		}

		if !user.IsActive {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户账户未激活"})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	})
}

func AdminRequired(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 首先验证用户身份
		AuthRequired(db, cfg)(c)

		if c.IsAborted() {
			return
		}

		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户信息不存在"})
			c.Abort()
			return
		}

		userObj := user.(models.User)
		if !userObj.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
			c.Abort()
			return
		}

		c.Next()
	})
}

// ========================= CORS中间件 =========================

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

// ========================= Cookie管理 =========================

const (
	// Cookie名称常量
	AuthCookieName     = "auth_token"
	RefreshCookieName  = "refresh_token"
	CSRFCookieName     = "csrf_token"
	
	// Cookie配置常量
	CookieMaxAge       = 7 * 24 * 60 * 60 // 7天
	RefreshCookieMaxAge = 30 * 24 * 60 * 60 // 30天
	CSRFCookieMaxAge   = 24 * 60 * 60 // 24小时
)

// CookieConfig Cookie配置
type CookieConfig struct {
	Domain     string
	Secure     bool
	SameSite   http.SameSite
	HttpOnly   bool
	Path       string
}

// GetCookieConfig 根据环境获取Cookie配置
func GetCookieConfig(isDevelopment bool, domain string) CookieConfig {
	config := CookieConfig{
		Domain:   domain,
		HttpOnly: true,
		Path:     "/",
	}
	
	if isDevelopment {
		// 开发环境配置
		config.Secure = false // 允许HTTP
		config.SameSite = http.SameSiteLaxMode
	} else {
		// 生产环境配置
		config.Secure = true // 强制HTTPS
		config.SameSite = http.SameSiteStrictMode
	}
	
	return config
}

// SetAuthCookie 设置认证Cookie
func SetAuthCookie(c *gin.Context, token string, config CookieConfig) {
	c.SetCookie(
		AuthCookieName,
		token,
		CookieMaxAge,
		config.Path,
		config.Domain,
		config.Secure,
		config.HttpOnly,
	)
	
	// 设置SameSite属性（Gin版本兼容处理）
	if cookie := findCookie(c.Writer.Header(), AuthCookieName); cookie != nil {
		cookie.SameSite = config.SameSite
	}
}

// SetRefreshCookie 设置刷新令牌Cookie
func SetRefreshCookie(c *gin.Context, refreshToken string, config CookieConfig) {
	c.SetCookie(
		RefreshCookieName,
		refreshToken,
		RefreshCookieMaxAge,
		config.Path,
		config.Domain,
		config.Secure,
		config.HttpOnly,
	)
	
	if cookie := findCookie(c.Writer.Header(), RefreshCookieName); cookie != nil {
		cookie.SameSite = config.SameSite
	}
}

// SetCSRFCookie 设置CSRF令牌Cookie
func SetCSRFCookie(c *gin.Context, csrfToken string, config CookieConfig) {
	// CSRF token需要被JavaScript访问，所以不能设置HttpOnly
	csrfConfig := config
	csrfConfig.HttpOnly = false
	
	c.SetCookie(
		CSRFCookieName,
		csrfToken,
		CSRFCookieMaxAge,
		csrfConfig.Path,
		csrfConfig.Domain,
		csrfConfig.Secure,
		csrfConfig.HttpOnly,
	)
	
	if cookie := findCookie(c.Writer.Header(), CSRFCookieName); cookie != nil {
		cookie.SameSite = csrfConfig.SameSite
	}
}

// GetAuthToken 从Cookie获取认证令牌
func GetAuthToken(c *gin.Context) (string, error) {
	return c.Cookie(AuthCookieName)
}

// GetRefreshToken 从Cookie获取刷新令牌
func GetRefreshToken(c *gin.Context) (string, error) {
	return c.Cookie(RefreshCookieName)
}

// GetCSRFToken 从Cookie获取CSRF令牌
func GetCSRFToken(c *gin.Context) (string, error) {
	return c.Cookie(CSRFCookieName)
}

// ClearAuthCookies 清除所有认证相关的Cookie
func ClearAuthCookies(c *gin.Context, config CookieConfig) {
	// 设置过期时间为过去的时间来删除Cookie
	c.SetCookie(AuthCookieName, "", -1, config.Path, config.Domain, config.Secure, config.HttpOnly)
	c.SetCookie(RefreshCookieName, "", -1, config.Path, config.Domain, config.Secure, config.HttpOnly)
	c.SetCookie(CSRFCookieName, "", -1, config.Path, config.Domain, config.Secure, false)
}

// findCookie 在响应头中查找指定名称的Cookie
func findCookie(header http.Header, name string) *http.Cookie {
	cookies := header["Set-Cookie"]
	for _, cookieStr := range cookies {
		if cookie := parseCookieHeader(cookieStr); cookie != nil && cookie.Name == name {
			return cookie
		}
	}
	return nil
}

// parseCookieHeader 解析Cookie头字符串
func parseCookieHeader(cookieStr string) *http.Cookie {
	req := &http.Request{Header: http.Header{"Cookie": []string{cookieStr}}}
	cookies := req.Cookies()
	if len(cookies) > 0 {
		return cookies[0]
	}
	return nil
}

// CSRF Protection Middleware
func CSRFProtection() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 只对状态改变的请求进行CSRF保护
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || 
		   c.Request.Method == "DELETE" || c.Request.Method == "PATCH" {
			
			// 从Cookie获取CSRF token
			csrfCookie, err := GetCSRFToken(c)
			if err != nil {
				c.JSON(http.StatusForbidden, gin.H{"error": "CSRF令牌缺失"})
				c.Abort()
				return
			}
			
			// 从请求头获取CSRF token
			csrfHeader := c.GetHeader("X-CSRF-Token")
			if csrfHeader == "" {
				c.JSON(http.StatusForbidden, gin.H{"error": "CSRF令牌缺失"})
				c.Abort()
				return
			}
			
			// 验证CSRF token
			if csrfCookie != csrfHeader {
				c.JSON(http.StatusForbidden, gin.H{"error": "CSRF令牌无效"})
				c.Abort()
				return
			}
		}
		
		c.Next()
	})
}

// ========================= 速率限制中间件 =========================

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
		for idx := 0; idx < len(xff); idx++ {
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
