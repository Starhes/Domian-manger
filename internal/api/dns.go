package api

import (
	"domain-manager/internal/models"
	"domain-manager/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DNSHandler struct {
	dnsService *services.DNSService
}

func NewDNSHandler(dnsService *services.DNSService) *DNSHandler {
	return &DNSHandler{dnsService: dnsService}
}

// 获取用户的DNS记录
func (h *DNSHandler) GetUserDNSRecords(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户信息不存在"})
		return
	}

	userObj := user.(models.User)
	records, err := h.dnsService.GetUserDNSRecords(userObj.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取DNS记录失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"records": records,
	})
}

// 创建DNS记录
func (h *DNSHandler) CreateDNSRecord(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户信息不存在"})
		return
	}

	userObj := user.(models.User)

	var req models.CreateDNSRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	record, err := h.dnsService.CreateDNSRecord(userObj.ID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "DNS记录创建成功",
		"record":  record,
	})
}

// 更新DNS记录
func (h *DNSHandler) UpdateDNSRecord(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户信息不存在"})
		return
	}

	userObj := user.(models.User)

	recordIDStr := c.Param("id")
	recordID, err := strconv.ParseUint(recordIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的记录ID"})
		return
	}

	var req models.UpdateDNSRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	record, err := h.dnsService.UpdateDNSRecord(userObj.ID, uint(recordID), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "DNS记录更新成功",
		"record":  record,
	})
}

// 删除DNS记录
func (h *DNSHandler) DeleteDNSRecord(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户信息不存在"})
		return
	}

	userObj := user.(models.User)

	recordIDStr := c.Param("id")
	recordID, err := strconv.ParseUint(recordIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的记录ID"})
		return
	}

	if err := h.dnsService.DeleteDNSRecord(userObj.ID, uint(recordID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "DNS记录删除成功",
	})
}

// 获取可用域名列表
func (h *DNSHandler) GetAvailableDomains(c *gin.Context) {
	domains, err := h.dnsService.GetAvailableDomains()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取域名列表失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"domains": domains,
	})
}
