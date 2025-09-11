package utils

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ErrorCode 错误代码枚举
type ErrorCode string

const (
	ErrInvalidRequest    ErrorCode = "INVALID_REQUEST"
	ErrUnauthorized      ErrorCode = "UNAUTHORIZED" 
	ErrForbidden         ErrorCode = "FORBIDDEN"
	ErrNotFound          ErrorCode = "NOT_FOUND"
	ErrConflict          ErrorCode = "CONFLICT"
	ErrValidationFailed  ErrorCode = "VALIDATION_FAILED"
	ErrInternalError     ErrorCode = "INTERNAL_ERROR"
	ErrServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
)

// ApiError API错误结构
type ApiError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"` // 仅在开发环境返回
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Error ApiError `json:"error"`
}

// isDevelopment 判断是否为开发环境
func isDevelopment(c *gin.Context) bool {
	mode := gin.Mode()
	return mode == gin.DebugMode || mode == gin.TestMode
}

// HandleError 统一错误处理函数
func HandleError(c *gin.Context, statusCode int, errorCode ErrorCode, userMessage string, internalError error) {
	// 记录详细错误到日志（生产环境重要）
	if internalError != nil {
		log.Printf("[ERROR] %s - %v - Path: %s, Method: %s, IP: %s", 
			string(errorCode), internalError, c.Request.URL.Path, c.Request.Method, c.ClientIP())
	}

	// 构建错误响应
	apiErr := ApiError{
		Code:    errorCode,
		Message: userMessage,
	}

	// 仅在开发环境提供详细错误信息
	if isDevelopment(c) && internalError != nil {
		apiErr.Details = internalError.Error()
	}

	c.JSON(statusCode, ErrorResponse{Error: apiErr})
}

// HandleValidationError 处理验证错误
func HandleValidationError(c *gin.Context, err error) {
	// 清理敏感信息的验证错误
	message := sanitizeValidationError(err.Error())
	HandleError(c, http.StatusBadRequest, ErrValidationFailed, message, err)
}

// HandleInternalError 处理内部服务器错误
func HandleInternalError(c *gin.Context, err error) {
	HandleError(c, http.StatusInternalServerError, ErrInternalError, "服务器内部错误，请稍后再试", err)
}

// HandleUnauthorized 处理未授权错误
func HandleUnauthorized(c *gin.Context, message string) {
	HandleError(c, http.StatusUnauthorized, ErrUnauthorized, message, nil)
}

// HandleForbidden 处理禁止访问错误
func HandleForbidden(c *gin.Context, message string) {
	HandleError(c, http.StatusForbidden, ErrForbidden, message, nil)
}

// HandleNotFound 处理资源未找到错误
func HandleNotFound(c *gin.Context, resource string) {
	message := resource + "不存在"
	HandleError(c, http.StatusNotFound, ErrNotFound, message, nil)
}

// HandleConflict 处理资源冲突错误
func HandleConflict(c *gin.Context, message string) {
	HandleError(c, http.StatusConflict, ErrConflict, message, nil)
}

// HandleBadRequest 处理请求参数错误
func HandleBadRequest(c *gin.Context, message string, err error) {
	HandleError(c, http.StatusBadRequest, ErrInvalidRequest, message, err)
}

// sanitizeValidationError 清理验证错误信息，移除敏感信息
func sanitizeValidationError(errMsg string) string {
	// 移除可能包含敏感信息的字段名
	sensitiveFields := []string{
		"password", "token", "secret", "key", "credential",
		"authorization", "session", "cookie",
	}
	
	lowerMsg := strings.ToLower(errMsg)
	for _, field := range sensitiveFields {
		if strings.Contains(lowerMsg, field) {
			return "请求参数验证失败"
		}
	}
	
	// 移除路径信息和技术细节
	if strings.Contains(errMsg, "/") || strings.Contains(errMsg, "\\") {
		return "请求参数格式错误"
	}
	
	// 限制错误消息长度
	if len(errMsg) > 100 {
		return "请求参数验证失败"
	}
	
	return errMsg
}

// LogSensitiveOperation 记录敏感操作日志
func LogSensitiveOperation(operation, userID, details string, c *gin.Context) {
	log.Printf("[SECURITY] Operation: %s, User: %s, IP: %s, UserAgent: %s, Details: %s",
		operation, userID, c.ClientIP(), c.GetHeader("User-Agent"), details)
}

// Success 返回成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// SuccessWithMessage 返回带消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}
