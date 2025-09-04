package controllers

import (
	"fmt"
	"net/http"
	"time"

	"image-host/config"
	"image-host/database"
	"image-host/models"
	"image-host/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UploadController struct{}

var Upload *UploadController

// InitUploadController 初始化上传控制器
func InitUploadController() {
	Upload = &UploadController{}
}

// UploadImage 上传图片
func (uc *UploadController) UploadImage(c *gin.Context) {
	// 获取上传的文件
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to get uploaded file",
			"code":  "INVALID_FILE",
		})
		return
	}
	defer file.Close()

	// 验证图片
	if err := services.ImageSvc.ValidateImage(header, config.AppConfig.AllowedTypes, config.AppConfig.MaxFileSize); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "VALIDATION_FAILED",
		})
		return
	}

	// 处理图片
	processedImage, err := services.ImageSvc.ProcessImage(file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process image",
			"code":  "PROCESSING_FAILED",
		})
		return
	}

	// 重置文件指针用于上传
	file.Seek(0, 0)

	// 上传到 R2
	r2Key, publicURL, err := services.R2.UploadFile(file, header)
	if err != nil {
		// 记录详细错误信息
		fmt.Printf("R2 upload error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to upload to storage: " + err.Error(),
			"code":  "UPLOAD_FAILED",
		})
		return
	}

	// 保存到数据库
	image := &models.Image{
		UUID:         uuid.New().String(),
		OriginalName: header.Filename,
		FileName:     header.Filename,
		FileSize:     header.Size,
		MimeType:     processedImage.MimeType,
		Width:        processedImage.Width,
		Height:       processedImage.Height,
		R2Key:        r2Key,
		PublicURL:    publicURL,
		UploadIP:     c.ClientIP(),
		UserAgent:    c.GetHeader("User-Agent"),
	}

	if err := database.DB.Create(image).Error; err != nil {
		// 如果数据库保存失败，删除已上传的文件
		services.R2.DeleteFile(r2Key)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save image metadata",
			"code":  "DATABASE_ERROR",
		})
		return
	}

	// 更新统计信息
	go uc.updateStats(header.Size)

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":           image.ID,
			"uuid":         image.UUID,
			"original_name": image.OriginalName,
			"file_size":    image.FileSize,
			"mime_type":    image.MimeType,
			"width":        image.Width,
			"height":       image.Height,
			"public_url":   image.PublicURL,
			"created_at":   image.CreatedAt,
		},
	})
}

// GetImage 获取图片信息
func (uc *UploadController) GetImage(c *gin.Context) {
	uuid := c.Param("uuid")
	if uuid == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing image UUID",
			"code":  "INVALID_UUID",
		})
		return
	}

	var image models.Image
	if err := database.DB.Where("uuid = ?", uuid).First(&image).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Image not found",
			"code":  "NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    image,
	})
}

// GetStats 获取统计信息
func (uc *UploadController) GetStats(c *gin.Context) {
	// 获取总图片数量
	var totalImages int64
	database.DB.Model(&models.Image{}).Count(&totalImages)

	// 获取总文件大小
	var totalSize int64
	database.DB.Model(&models.Image{}).Select("COALESCE(SUM(file_size), 0)").Scan(&totalSize)

	// 获取今日上传数量
	var todayImages int64
	today := time.Now().Format("2006-01-02")
	database.DB.Model(&models.Image{}).Where("DATE(created_at) = ?", today).Count(&todayImages)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_images": totalImages,
			"total_size":   totalSize,
			"today_images": todayImages,
		},
	})
}

// BatchUpload 批量上传图片
func (uc *UploadController) BatchUpload(c *gin.Context) {
	// 解析多文件上传
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse multipart form",
			"code":  "INVALID_FORM",
		})
		return
	}

	files := form.File["images"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No files uploaded",
			"code":  "NO_FILES",
		})
		return
	}

	// 限制批量上传数量
	if len(files) > 10 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Too many files, maximum 10 files allowed",
			"code":  "TOO_MANY_FILES",
		})
		return
	}

	var results []gin.H
	var errors []gin.H

	// 处理每个文件
	for i, header := range files {
		file, err := header.Open()
		if err != nil {
			errors = append(errors, gin.H{
				"index": i,
				"filename": header.Filename,
				"error": "Failed to open file",
			})
			continue
		}

		// 验证图片
		if err := services.ImageSvc.ValidateImage(header, config.AppConfig.AllowedTypes, config.AppConfig.MaxFileSize); err != nil {
			errors = append(errors, gin.H{
				"index": i,
				"filename": header.Filename,
				"error": err.Error(),
			})
			file.Close()
			continue
		}

		// 处理图片
		processedImage, err := services.ImageSvc.ProcessImage(file, header)
		if err != nil {
			errors = append(errors, gin.H{
				"index": i,
				"filename": header.Filename,
				"error": "Failed to process image",
			})
			file.Close()
			continue
		}

		// 重置文件指针
		file.Seek(0, 0)

		// 上传到 R2
		r2Key, publicURL, err := services.R2.UploadFile(file, header)
		if err != nil {
			errors = append(errors, gin.H{
				"index": i,
				"filename": header.Filename,
				"error": "Failed to upload to storage",
			})
			file.Close()
			continue
		}

		// 保存到数据库
		image := &models.Image{
			UUID:         uuid.New().String(),
			OriginalName: header.Filename,
			FileName:     header.Filename,
			FileSize:     header.Size,
			MimeType:     processedImage.MimeType,
			Width:        processedImage.Width,
			Height:       processedImage.Height,
			R2Key:        r2Key,
			PublicURL:    publicURL,
			UploadIP:     c.ClientIP(),
			UserAgent:    c.GetHeader("User-Agent"),
		}

		if err := database.DB.Create(image).Error; err != nil {
			// 如果数据库保存失败，删除已上传的文件
			services.R2.DeleteFile(r2Key)
			errors = append(errors, gin.H{
				"index": i,
				"filename": header.Filename,
				"error": "Failed to save image metadata",
			})
			file.Close()
			continue
		}

		// 添加到成功结果
		results = append(results, gin.H{
			"index":        i,
			"id":           image.ID,
			"uuid":         image.UUID,
			"original_name": image.OriginalName,
			"file_size":    image.FileSize,
			"mime_type":    image.MimeType,
			"width":        image.Width,
			"height":       image.Height,
			"public_url":   image.PublicURL,
			"created_at":   image.CreatedAt,
		})

		// 更新统计信息
		go uc.updateStats(header.Size)

		file.Close()
	}

	// 返回批量上传结果
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"successful": len(results),
			"failed":     len(errors),
			"results":    results,
			"errors":     errors,
		},
	})
}

// updateStats 更新统计信息
func (uc *UploadController) updateStats(fileSize int64) {
	today := time.Now().Truncate(24 * time.Hour)

	// 查找或创建今日统计记录
	var stats models.ImageStats
	result := database.DB.Where("date = ?", today).First(&stats)

	if result.Error != nil {
		// 创建新记录
		stats = models.ImageStats{
			Date:        today,
			TotalImages: 1,
			TotalSize:   fileSize,
		}
		database.DB.Create(&stats)
	} else {
		// 更新现有记录
		stats.TotalImages++
		stats.TotalSize += fileSize
		database.DB.Save(&stats)
	}
}