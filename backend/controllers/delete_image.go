package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"backend/database"
	"backend/models"
	"backend/services"
)

// DeleteImage 按 UUID 删除图片：先删本地文件，再硬删数据库记录
// DELETE /api/v1/images/:uuid
func (uc *UploadController) DeleteImage(c *gin.Context) {
	uuid := c.Param("uuid")
	if uuid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid"})
		return
	}

	var img models.Image
	if err := database.DB.Where("uuid = ?", uuid).First(&img).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "image not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// 删除本地文件（忽略错误但记录）
	if img.R2Key != "" {
		_ = services.R2.DeleteFile(img.R2Key)
	}

	// 硬删除数据库记录
	if err := database.DB.Unscoped().Delete(&img).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
