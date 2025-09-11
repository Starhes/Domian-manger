package api

import (
	"domain-manager/internal/models"
	"domain-manager/internal/services"
	"domain-manager/internal/utils"
	"domain-manager/internal/middleware"
	"net/http"
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleValidationError(c, err)
		return
	}

	if err := h.authService.RegisterWithContext(c, req); err != nil {
		utils.HandleBadRequest(c, err.Error(), err)
		return
	}

	utils.SuccessWithMessage(c, "注册成功，请检查邮箱激活账户", nil)
}

// 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleValidationError(c, err)
		return
	}

	resp, err := h.authService.Login(req)
	if err != nil {
		utils.HandleUnauthorized(c, "邮箱或密码错误")
		utils.LogSensitiveOperation("LOGIN_FAILED", req.Email, "Invalid credentials", c)
		return
	}

	// 生成CSRF令牌
	csrfToken, err := generateCSRFToken()
	if err != nil {
		utils.HandleInternalError(c, err)
		return
	}

	// 获取Cookie配置
	cookieConfig := middleware.GetCookieConfig(
		c.GetHeader("X-Development") == "true", // 开发环境标识
		c.Request.Host,
	)

	// 设置认证Cookie
	middleware.SetAuthCookie(c, resp.Token, cookieConfig)
	if resp.RefreshToken != "" {
		middleware.SetRefreshCookie(c, resp.RefreshToken, cookieConfig)
	}
	middleware.SetCSRFCookie(c, csrfToken, cookieConfig)

	utils.LogSensitiveOperation("LOGIN_SUCCESS", req.Email, "User logged in", c)
	
	// 返回响应（不包含敏感令牌）
	utils.Success(c, gin.H{
		"user":       resp.User,
		"csrf_token": csrfToken, // 前端需要在请求头中包含此令牌
		"message":    "登录成功",
	})
}

// 邮箱验证
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		utils.HandleBadRequest(c, "验证令牌不能为空", nil)
		return
	}

	if err := h.authService.VerifyEmail(token); err != nil {
		utils.HandleBadRequest(c, "验证令牌无效或已过期", err)
		return
	}

	utils.LogSensitiveOperation("EMAIL_VERIFIED", "unknown", "Email verification successful", c)
	utils.SuccessWithMessage(c, "邮箱验证成功，账户已激活", nil)
}

// 忘记密码
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleValidationError(c, err)
		return
	}

	if err := h.authService.ForgotPassword(req); err != nil {
		utils.HandleInternalError(c, err)
		return
	}

	utils.LogSensitiveOperation("PASSWORD_RESET_REQUESTED", req.Email, "Password reset requested", c)
	utils.SuccessWithMessage(c, "如果邮箱存在，重置链接已发送", nil)
}

// 重置密码
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleValidationError(c, err)
		return
	}

	if err := h.authService.ResetPassword(req); err != nil {
		utils.HandleBadRequest(c, "密码重置失败，令牌可能无效或已过期", err)
		return
	}

	utils.LogSensitiveOperation("PASSWORD_RESET_SUCCESS", "unknown", "Password reset successful", c)
	utils.SuccessWithMessage(c, "密码重置成功", nil)
}

// 获取用户资料
func (h *AuthHandler) GetProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		utils.HandleUnauthorized(c, "用户信息不存在")
		return
	}

	utils.Success(c, gin.H{"user": user.(models.User)})
}

// 更新用户资料
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		utils.HandleUnauthorized(c, "用户信息不存在")
		return
	}

	userObj := user.(models.User)

	var req struct {
		Email    string `json:"email" binding:"omitempty,email"`
		Password string `json:"password" binding:"omitempty,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleValidationError(c, err)
		return
	}

	// 调用服务层更新用户资料
	updatedUser, err := h.authService.UpdateProfile(userObj.ID, req.Email, req.Password)
	if err != nil {
		utils.HandleBadRequest(c, err.Error(), err)
		return
	}

	utils.LogSensitiveOperation("PROFILE_UPDATED", string(rune(userObj.ID)), "Profile updated", c)
	utils.SuccessWithMessage(c, "资料更新成功", gin.H{"user": updatedUser})
}

// 用户登出
func (h *AuthHandler) Logout(c *gin.Context) {
	// 从上下文获取token
	tokenString, exists := c.Get("token")
	if !exists {
		utils.HandleBadRequest(c, "无法获取令牌", nil)
		return
	}

	// 获取用户信息用于日志记录
	user, _ := c.Get("user")
	userID := "unknown"
	if user != nil {
		userObj := user.(models.User)
		userID = string(rune(userObj.ID))
	}

	// 撤销token
	if err := h.authService.Logout(tokenString.(string)); err != nil {
		utils.HandleInternalError(c, err)
		return
	}

	// 获取Cookie配置并清除Cookie
	cookieConfig := middleware.GetCookieConfig(
		c.GetHeader("X-Development") == "true",
		c.Request.Host,
	)
	middleware.ClearAuthCookies(c, cookieConfig)

	utils.LogSensitiveOperation("LOGOUT_SUCCESS", userID, "User logged out", c)
	utils.SuccessWithMessage(c, "登出成功", nil)
}

// generateCSRFToken 生成CSRF令牌
func generateCSRFToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
