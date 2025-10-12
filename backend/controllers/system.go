package controllers

import (
	"errors"
	"net/http"
	"path/filepath"
	"time"

	"image-host/config"
	"image-host/database"
	"image-host/models"
	"image-host/services"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"gorm.io/gorm"
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

	// 内存与 CPU
	vm, _ := mem.VirtualMemory()
	cpuPercents, _ := cpu.Percent(200*time.Millisecond, false)
	var cpuUsage float64
	if len(cpuPercents) > 0 {
		cpuUsage = cpuPercents[0]
	}

	// 规范化内存与磁盘占比
	memTotal := int64(0)
	memUsed := int64(0)
	memFree := int64(0)
	memUsedPercent := float64(0)
	if vm != nil {
		memTotal = int64(vm.Total)
		memUsed = int64(vm.Used)
		// 选择 Available 作为“可用”，更贴近日常理解
		memFree = int64(vm.Available)
		memUsedPercent = vm.UsedPercent
	}
	diskUsedPercent := usage.UsedPercent

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

			// 资源占用（实时）
			"disk_used_percent": diskUsedPercent,
			"mem_total":         memTotal,
			"mem_used":          memUsed,
			"mem_free":          memFree,
			"mem_used_percent":  memUsedPercent,
			"cpu_usage":         cpuUsage,
		},
	})
}

// GetStorageConfig 获取存储配置（仅 root）
// GET /api/v1/system/storage-config
func (sc *SystemController) GetStorageConfig(c *gin.Context) {
	username := c.GetString("username")
	if username != config.AppConfig.DefaultAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	var cfg models.StorageConfig
	result := database.DB.First(&cfg)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			cfg = models.StorageConfig{Provider: "local"}
			if err := database.DB.Create(&cfg).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to init storage config"})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load storage config"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
		"provider": cfg.Provider,
		"r2": gin.H{
			"account_id":        cfg.R2AccountID,
			"access_key_id":     cfg.R2AccessKeyID,
			"secret_access_key": cfg.R2SecretAccessKey,
			"bucket_name":       cfg.R2BucketName,
			"endpoint":          cfg.R2Endpoint,
			"region":            cfg.R2Region,
			"public_url":        cfg.R2PublicURL,
		},
		"oss": gin.H{
			"access_key_id":     cfg.OSSAccessKeyID,
			"secret_access_key": cfg.OSSSecretAccessKey,
			"bucket_name":       cfg.OSSBucketName,
			"endpoint":          cfg.OSSEndpoint,
			"region":            cfg.OSSRegion,
			"public_url":        cfg.OSSPublicURL,
		},
	}})
}

// UpdateStorageConfig 更新存储配置（仅 root）
// POST /api/v1/system/storage-config
func (sc *SystemController) UpdateStorageConfig(c *gin.Context) {
	username := c.GetString("username")
	if username != config.AppConfig.DefaultAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	var payload struct {
		Provider string `json:"provider"` // local / r2 / oss
		R2       struct {
			AccountID       string `json:"account_id"`
			AccessKeyID     string `json:"access_key_id"`
			SecretAccessKey string `json:"secret_access_key"`
			BucketName      string `json:"bucket_name"`
			Endpoint        string `json:"endpoint"`
			Region          string `json:"region"`
			PublicURL       string `json:"public_url"`
		} `json:"r2"`
		OSS struct {
			AccessKeyID     string `json:"access_key_id"`
			SecretAccessKey string `json:"secret_access_key"`
			BucketName      string `json:"bucket_name"`
			Endpoint        string `json:"endpoint"`
			Region          string `json:"region"`
			PublicURL       string `json:"public_url"`
		} `json:"oss"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	if payload.Provider != "local" && payload.Provider != "r2" && payload.Provider != "oss" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider"})
		return
	}

	var cfg models.StorageConfig
	result := database.DB.First(&cfg)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load storage config"})
		return
	}
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		cfg = models.StorageConfig{Provider: "local"}
	}

	// 更新字段
	cfg.Provider = payload.Provider

	// R2
	cfg.R2AccountID = payload.R2.AccountID
	cfg.R2AccessKeyID = payload.R2.AccessKeyID
	cfg.R2SecretAccessKey = payload.R2.SecretAccessKey
	cfg.R2BucketName = payload.R2.BucketName
	cfg.R2Endpoint = payload.R2.Endpoint
	cfg.R2Region = payload.R2.Region
	if cfg.R2Region == "" {
		cfg.R2Region = "auto"
	}
	cfg.R2PublicURL = payload.R2.PublicURL

	// OSS
	cfg.OSSAccessKeyID = payload.OSS.AccessKeyID
	cfg.OSSSecretAccessKey = payload.OSS.SecretAccessKey
	cfg.OSSBucketName = payload.OSS.BucketName
	cfg.OSSEndpoint = payload.OSS.Endpoint
	cfg.OSSRegion = payload.OSS.Region
	cfg.OSSPublicURL = payload.OSS.PublicURL

	// 保存
	if cfg.ID == 0 {
		if err := database.DB.Create(&cfg).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create storage config"})
			return
		}
	} else {
		if err := database.DB.Save(&cfg).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update storage config"})
			return
		}
	}

	// 运行时生效：更新内存中的存储客户端配置（local/r2）
	if services.R2 != nil {
		services.R2.ReloadConfig()
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
