package controllers

import (
	"net/http"
	"time"

	"image-host/config"
	"image-host/database"
	"image-host/models"
	"image-host/services"

	"github.com/gin-gonic/gin"
)

type GuestCodeController struct{}

var GuestCode = &GuestCodeController{}

// Create 生成游客码
// 请求: { days?: number, expires_at?: number(unix秒), permanent?: boolean }
func (g *GuestCodeController) Create(c *gin.Context) {
	creator := c.GetString("username")
	if creator != config.AppConfig.DefaultAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	var payload struct {
		Days      *int   `json:"days"`
		ExpiresAt *int64 `json:"expires_at"`
		Permanent bool   `json:"permanent"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	var expPtr *time.Time
	if payload.Permanent {
		expPtr = nil
	} else if payload.ExpiresAt != nil {
		t := time.Unix(*payload.ExpiresAt, 0)
		expPtr = &t
	} else if payload.Days != nil && *payload.Days > 0 {
		t := time.Now().Add(time.Duration(*payload.Days) * 24 * time.Hour)
		expPtr = &t
	} else {
		// 默认 1 天
		t := time.Now().Add(24 * time.Hour)
		expPtr = &t
	}

	code, err := services.Guest.GenerateCode(creator, expPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate code"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": code})
}

// List 列出游客码
func (g *GuestCodeController) List(c *gin.Context) {
	creator := c.GetString("username")
	if creator != config.AppConfig.DefaultAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	var list []models.GuestCode
	if err := database.DB.Order("id DESC").Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": list})
}

// Delete 删除游客码并清理其图片
func (g *GuestCodeController) Delete(c *gin.Context) {
	creator := c.GetString("username")
	if creator != config.AppConfig.DefaultAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	id := c.Param("id")
	if err := services.Guest.DeleteCodeAndImages(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
