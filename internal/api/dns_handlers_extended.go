package api

import (
	"domain-manager/internal/models"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// BatchCreateDNSRecords 批量创建DNS记录
func (h *DNSHandler) BatchCreateDNSRecords(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户信息不存在"})
		return
	}

	userObj := user.(models.User)

	var req models.BatchDNSRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	successRecords, errorList := h.dnsService.BatchCreateDNSRecords(userObj.ID, req)

	response := gin.H{
		"success_count": len(successRecords),
		"error_count":   len(errorList),
		"records":       successRecords,
	}

	if len(errorList) > 0 {
		var errorMessages []string
		for _, err := range errorList {
			errorMessages = append(errorMessages, err.Error())
		}
		response["errors"] = errorMessages
	}

	if len(successRecords) > 0 {
		c.JSON(http.StatusCreated, response)
	} else {
		c.JSON(http.StatusBadRequest, response)
	}
}

// ExportDNSRecords 导出DNS记录
func (h *DNSHandler) ExportDNSRecords(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户信息不存在"})
		return
	}

	userObj := user.(models.User)

	exportData, err := h.dnsService.ExportDNSRecords(userObj.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "导出DNS记录失败", "details": err.Error()})
		return
	}

	// 检查是否请求下载文件
	format := c.Query("format")
	if format == "file" {
		// 生成JSON文件
		jsonData, err := json.MarshalIndent(exportData, "", "  ")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成导出文件失败"})
			return
		}

		c.Header("Content-Disposition", "attachment; filename=dns_records_export.json")
		c.Header("Content-Type", "application/json")
		c.Data(http.StatusOK, "application/json", jsonData)
		return
	}

	c.JSON(http.StatusOK, exportData)
}

// ImportDNSRecords 导入DNS记录
func (h *DNSHandler) ImportDNSRecords(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户信息不存在"})
		return
	}

	userObj := user.(models.User)

	var req struct {
		Records []models.DNSRecordExport `json:"records" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	successRecords, errorList := h.dnsService.ImportDNSRecords(userObj.ID, req.Records)

	response := gin.H{
		"success_count": len(successRecords),
		"error_count":   len(errorList),
		"records":       successRecords,
	}

	if len(errorList) > 0 {
		var errorMessages []string
		for _, err := range errorList {
			errorMessages = append(errorMessages, err.Error())
		}
		response["errors"] = errorMessages
	}

	if len(successRecords) > 0 {
		c.JSON(http.StatusCreated, response)
	} else {
		c.JSON(http.StatusBadRequest, response)
	}
}

// GetDNSRecordTypes 获取DNS记录类型列表
func (h *DNSHandler) GetDNSRecordTypes(c *gin.Context) {
	types := []map[string]interface{}{
		{
			"value":       "A",
			"label":       "A记录",
			"description": models.GetDNSRecordTypeDescription("A"),
			"fields":      []string{"value"},
		},
		{
			"value":       "AAAA",
			"label":       "AAAA记录",
			"description": models.GetDNSRecordTypeDescription("AAAA"),
			"fields":      []string{"value"},
		},
		{
			"value":       "CNAME",
			"label":       "CNAME记录",
			"description": models.GetDNSRecordTypeDescription("CNAME"),
			"fields":      []string{"value"},
		},
		{
			"value":       "MX",
			"label":       "MX记录",
			"description": models.GetDNSRecordTypeDescription("MX"),
			"fields":      []string{"value", "priority"},
		},
		{
			"value":       "TXT",
			"label":       "TXT记录",
			"description": models.GetDNSRecordTypeDescription("TXT"),
			"fields":      []string{"value"},
		},
		{
			"value":       "NS",
			"label":       "NS记录",
			"description": models.GetDNSRecordTypeDescription("NS"),
			"fields":      []string{"value"},
		},
		{
			"value":       "PTR",
			"label":       "PTR记录",
			"description": models.GetDNSRecordTypeDescription("PTR"),
			"fields":      []string{"value"},
		},
		{
			"value":       "SRV",
			"label":       "SRV记录",
			"description": models.GetDNSRecordTypeDescription("SRV"),
			"fields":      []string{"value", "priority", "weight", "port"},
		},
		{
			"value":       "CAA",
			"label":       "CAA记录",
			"description": models.GetDNSRecordTypeDescription("CAA"),
			"fields":      []string{"value"},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"types": types,
	})
}

// GetTTLOptions 获取TTL选项列表
func (h *DNSHandler) GetTTLOptions(c *gin.Context) {
	options := models.GetTTLOptions()
	c.JSON(http.StatusOK, gin.H{
		"options": options,
	})
}

// ValidateDNSRecordFile 验证DNS记录文件
func (h *DNSHandler) ValidateDNSRecordFile(c *gin.Context) {
	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	records, err := h.dnsService.ValidateDNSRecordFile(req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件验证失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":   true,
		"records": records,
		"count":   len(records),
	})
}

// GetDNSRecordStats 获取DNS记录统计信息
func (h *DNSHandler) GetDNSRecordStats(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户信息不存在"})
		return
	}

	userObj := user.(models.User)

	// 获取用户的DNS记录统计
	var totalRecords int64
	if err := h.dnsService.GetDB().Model(&models.DNSRecord{}).Where("user_id = ?", userObj.ID).Count(&totalRecords).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取记录统计失败"})
		return
	}

	// 按类型统计
	var typeStats []struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}
	if err := h.dnsService.GetDB().Model(&models.DNSRecord{}).
		Select("type, count(*) as count").
		Where("user_id = ?", userObj.ID).
		Group("type").
		Scan(&typeStats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取类型统计失败"})
		return
	}

	// 按域名统计
	var domainStats []struct {
		DomainName string `json:"domain_name"`
		Count      int64  `json:"count"`
	}
	if err := h.dnsService.GetDB().Table("dns_records").
		Select("domains.name as domain_name, count(*) as count").
		Joins("JOIN domains ON dns_records.domain_id = domains.id").
		Where("dns_records.user_id = ?", userObj.ID).
		Group("domains.name").
		Scan(&domainStats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取域名统计失败"})
		return
	}

	quotaUsed := 0.0
	if userObj.DNSRecordQuota > 0 {
		quotaUsed = float64(totalRecords) / float64(userObj.DNSRecordQuota) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"total_records": totalRecords,
		"quota":         userObj.DNSRecordQuota,
		"quota_used":    quotaUsed,
		"type_stats":    typeStats,
		"domain_stats":  domainStats,
	})
}