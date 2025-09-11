package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

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
