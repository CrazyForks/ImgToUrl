package controllers

import (
	"net/http"
	"path/filepath"
	"time"

	"image-host/config"
	"image-host/database"
	"image-host/models"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/disk"
)

type SystemController struct{}

var System *SystemController
var serviceStartTime time.Time

func InitSystemController() {
	System = &SystemController{}
	serviceStartTime = time.Now()
}

// Status 返回系统运行状态与容量信息
// GET /api/v1/system/status
func (sc *SystemController) Status(c *gin.Context) {
	// 计算进程运行时长
	uptimeSeconds := int64(time.Since(serviceStartTime).Seconds())

	// 统计上传目录所在分区的磁盘容量
	uploadPath := config.AppConfig.UploadPath
	if abs, err := filepath.Abs(uploadPath); err == nil {
		uploadPath = abs
	}
	usage, err := disk.Usage(uploadPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to stat filesystem: " + err.Error(),
		})
		return
	}
	diskTotal := int64(usage.Total)
	diskFree := int64(usage.Free)
	diskUsed := int64(usage.Used)
	if diskTotal < 0 {
		diskTotal = 0
	}
	if diskFree < 0 {
		diskFree = 0
	}
	if diskUsed < 0 {
		diskUsed = 0
	}

	// 数据库汇总
	var totalImages int64
	database.DB.Model(&models.Image{}).Count(&totalImages)

	var totalSize int64
	database.DB.Model(&models.Image{}).Select("COALESCE(SUM(file_size), 0)").Scan(&totalSize)

	todayStr := time.Now().Format("2006-01-02")
	var todayImages int64
	database.DB.Model(&models.Image{}).Where("DATE(created_at) = ?", todayStr).Count(&todayImages)

	var todaySize int64
	database.DB.Model(&models.Image{}).Where("DATE(created_at) = ?", todayStr).Select("COALESCE(SUM(file_size), 0)").Scan(&todaySize)

	// 平均/最大/最小
	var avgSizeFloat float64
	database.DB.Model(&models.Image{}).Select("COALESCE(AVG(file_size), 0)").Scan(&avgSizeFloat)
	averageSize := int64(avgSizeFloat)

	var maxSize int64
	database.DB.Model(&models.Image{}).Select("COALESCE(MAX(file_size), 0)").Scan(&maxSize)

	var minSize int64
	database.DB.Model(&models.Image{}).Select("COALESCE(MIN(file_size), 0)").Scan(&minSize)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"uptime_seconds":    uptimeSeconds,
			"disk_total":        diskTotal,
			"disk_free":         diskFree,
			"disk_used":         diskUsed,
			"total_images":      totalImages,
			"total_size":        totalSize,
			"today_images":      todayImages,
			"today_size":        todaySize,
			"average_file_size": averageSize,
			"max_file_size":     maxSize,
			"min_file_size":     minSize,
		},
	})
}
