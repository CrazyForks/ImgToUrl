package controllers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"strings"
	"path"

	"image-host/config"
	"image-host/database"
	"image-host/models"
	"image-host/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	alioss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"gorm.io/gorm"
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
	uploader := c.GetString("username")
	// 根据最终 URL 判断实际存储渠道（避免 OSS 失败回退本地但记录仍为 oss）
	providerUsed := services.R2.Provider()
	if strings.HasPrefix(publicURL, "/uploads/") {
		providerUsed = "local"
	}
	image := &models.Image{
		UUID:            uuid.New().String(),
		OriginalName:    header.Filename,
		FileName:        header.Filename,
		FileSize:        header.Size,
		MimeType:        processedImage.MimeType,
		Width:           processedImage.Width,
		Height:          processedImage.Height,
		R2Key:           r2Key,
		PublicURL:       publicURL,
		StorageProvider: providerUsed,
		UploadIP:        c.ClientIP(),
		UserAgent:       c.GetHeader("User-Agent"),
		Uploader:        uploader,
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
			"id":            image.ID,
			"uuid":          image.UUID,
			"original_name": image.OriginalName,
			"file_size":     image.FileSize,
			"mime_type":     image.MimeType,
			"width":         image.Width,
			"height":        image.Height,
			"public_url":    image.PublicURL,
			"created_at":    image.CreatedAt,
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
				"index":    i,
				"filename": header.Filename,
				"error":    "Failed to open file",
			})
			continue
		}

		// 验证图片
		if err := services.ImageSvc.ValidateImage(header, config.AppConfig.AllowedTypes, config.AppConfig.MaxFileSize); err != nil {
			errors = append(errors, gin.H{
				"index":    i,
				"filename": header.Filename,
				"error":    err.Error(),
			})
			file.Close()
			continue
		}

		// 处理图片
		processedImage, err := services.ImageSvc.ProcessImage(file, header)
		if err != nil {
			errors = append(errors, gin.H{
				"index":    i,
				"filename": header.Filename,
				"error":    "Failed to process image",
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
				"index":    i,
				"filename": header.Filename,
				"error":    "Failed to upload to storage",
			})
			file.Close()
			continue
		}

		// 保存到数据库（根据最终 URL 判断实际存储渠道）
		providerUsed := services.R2.Provider()
		if strings.HasPrefix(publicURL, "/uploads/") {
			providerUsed = "local"
		}
		image := &models.Image{
			UUID:            uuid.New().String(),
			OriginalName:    header.Filename,
			FileName:        header.Filename,
			FileSize:        header.Size,
			MimeType:        processedImage.MimeType,
			Width:           processedImage.Width,
			Height:          processedImage.Height,
			R2Key:           r2Key,
			PublicURL:       publicURL,
			StorageProvider: providerUsed,
			UploadIP:        c.ClientIP(),
			UserAgent:       c.GetHeader("User-Agent"),
			Uploader:        c.GetString("username"),
		}

		if err := database.DB.Create(image).Error; err != nil {
			// 如果数据库保存失败，删除已上传的文件
			services.R2.DeleteFile(r2Key)
			errors = append(errors, gin.H{
				"index":    i,
				"filename": header.Filename,
				"error":    "Failed to save image metadata",
			})
			file.Close()
			continue
		}

		// 添加到成功结果
		results = append(results, gin.H{
			"index":         i,
			"id":            image.ID,
			"uuid":          image.UUID,
			"original_name": image.OriginalName,
			"file_size":     image.FileSize,
			"mime_type":     image.MimeType,
			"width":         image.Width,
			"height":        image.Height,
			"public_url":    image.PublicURL,
			"created_at":    image.CreatedAt,
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

/*
*
ListImages 分页获取图片列表
GET /api/v1/images?page=1&page_size=20
响应：

	{
	  "success": true,
	  "data": { "items": [...], "total": 123, "page": 1, "page_size": 20 }
	}
*/
func (uc *UploadController) ListImages(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("page_size", "20")

	page, _ := strconv.Atoi(pageStr)
	size, _ := strconv.Atoi(sizeStr)
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	var total int64
	username := c.GetString("username")
	dbq := database.DB.Model(&models.Image{})
	// 非 root 仅查看自己的
	if username != config.AppConfig.DefaultAdmin {
		dbq = dbq.Where("uploader = ?", username)
	}
	if err := dbq.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to count images",
			"code":  "DATABASE_ERROR",
		})
		return
	}

	var images []models.Image
	query := database.DB
	if username != config.AppConfig.DefaultAdmin {
		query = query.Where("uploader = ?", username)
	}
	if err := query.
		Order("created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&images).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch images",
			"code":  "DATABASE_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"items":     images,
			"total":     total,
			"page":      page,
			"page_size": size,
		},
	})
}

// DeleteImage 删除图片（先删文件，再硬删记录）
func (uc *UploadController) DeleteImage(c *gin.Context) {
	u := c.Param("uuid")
	if u == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing image UUID",
			"code":  "INVALID_UUID",
		})
		return
	}

	var image models.Image
	username := c.GetString("username")
	q := database.DB
	// 非 root 只能删除自己的
	if username != config.AppConfig.DefaultAdmin {
		q = q.Where("uploader = ?", username)
	}
	if err := q.Where("uuid = ?", u).First(&image).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Image not found",
			"code":  "NOT_FOUND",
		})
		return
	}

	// 删除存储中的文件（根据图片的存储渠道决定删除本地或 R2）
	if image.R2Key != "" {
		// 对旧数据进行渠道推断：public_url 以 /uploads/ 开头视为 local，否则 r2
		provider := strings.ToLower(strings.TrimSpace(image.StorageProvider))
		if provider == "" {
			if strings.HasPrefix(image.PublicURL, "/uploads/") {
				provider = "local"
			} else {
				provider = "r2"
			}
		}
		_ = services.R2.DeleteFileByProvider(image.R2Key, provider)
	}

	// 硬删除数据库记录
	if err := database.DB.Unscoped().Delete(&image).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete image record",
			"code":  "DATABASE_ERROR",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
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

// ClearImages 一键清空图片（仅管理员）
// POST /api/v1/images/clear
// body: { "scope": "all" | "local" | "r2" }  默认 all
func (uc *UploadController) ClearImages(c *gin.Context) {
	username := c.GetString("username")
	if username != config.AppConfig.DefaultAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	var payload struct {
		Scope string `json:"scope"`
	}
	_ = c.ShouldBindJSON(&payload)
	scope := strings.ToLower(strings.TrimSpace(payload.Scope))
	if scope != "local" && scope != "r2" && scope != "oss" {
		scope = "all"
	}

	// 读取将要删除的记录，用于删除实际文件
	var images []models.Image
	q := database.DB
	switch scope {
	case "local":
		// 兼容旧数据：storage_provider=local 或 public_url 以 /uploads/ 开头 或 r2_key 为空
		q = q.Where("storage_provider = ? OR public_url LIKE ? OR r2_key = ''", "local", "/uploads/%")
	case "r2":
		// 仅清理 R2 渠道
		q = q.Where("storage_provider = ?", "r2")
	case "oss":
		// 仅清理 OSS 渠道
		q = q.Where("storage_provider = ?", "oss")
	default:
		// all：不加条件
	}
	if err := q.Find(&images).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch images for clearing"})
		return
	}

	for _, img := range images {
		if img.R2Key != "" {
			// 根据图片渠道删除对应位置的文件
			provider := strings.ToLower(strings.TrimSpace(img.StorageProvider))
			if provider == "" {
				if strings.HasPrefix(img.PublicURL, "/uploads/") {
					provider = "local"
				} else {
					provider = "r2"
				}
			}
			_ = services.R2.DeleteFileByProvider(img.R2Key, provider)
		}
	}

	// 硬删数据库记录
	dq := database.DB.Unscoped()
	var err error
	switch scope {
	case "local":
		err = dq.Where("storage_provider = ? OR public_url LIKE ? OR r2_key = ''", "local", "/uploads/%").Delete(&models.Image{}).Error
	case "r2":
		err = dq.Where("storage_provider = ?", "r2").Delete(&models.Image{}).Error
	case "oss":
		err = dq.Where("storage_provider = ?", "oss").Delete(&models.Image{}).Error
	default:
		// all：允许全表删除
		err = dq.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Image{}).Error
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete image records"})
		return
	}

	// 清空统计信息，避免展示不一致
	_ = database.DB.Exec("DELETE FROM image_stats").Error

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"scope": scope}})
}

// SyncImages 同步对象存储中的对象为站点图片记录（仅管理员）
// POST /api/v1/images/sync
// body: {
//   "provider": "r2" | "oss",
//   "prefix": "images/",           // 可选，限定前缀
//   "limit": 1000,                 // 可选，最多同步条数（默认 1000，-1 表示全部）
//   "overwrite": false             // 可选，已存在记录是否覆盖更新
// }
func (uc *UploadController) SyncImages(c *gin.Context) {
	username := c.GetString("username")
	if username != config.AppConfig.DefaultAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	var payload struct {
		Provider  string `json:"provider"`
		Prefix    string `json:"prefix"`
		Limit     int    `json:"limit"`
		Overwrite bool   `json:"overwrite"`
	}
	_ = c.ShouldBindJSON(&payload)
	pv := strings.ToLower(strings.TrimSpace(payload.Provider))
	if pv != "r2" && pv != "oss" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider"})
		return
	}
	prefix := strings.TrimSpace(payload.Prefix)
	limit := payload.Limit
	if limit == 0 {
		limit = 1000
	}

	// 读取存储配置
	var cfg models.StorageConfig
	if err := database.DB.First(&cfg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load storage config"})
		return
	}

	added := 0
	skipped := 0
	updated := 0

	if pv == "r2" {
		// 初始化临时 R2 客户端
		ep := normalizeEndpointLocal(strings.TrimSpace(cfg.R2Endpoint))
		ak := strings.TrimSpace(cfg.R2AccessKeyID)
		sk := strings.TrimSpace(cfg.R2SecretAccessKey)
		bn := strings.TrimSpace(cfg.R2BucketName)
		if ep == "" || ak == "" || sk == "" || bn == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "R2 config incomplete"})
			return
		}
		client, err := minio.New(ep, &minio.Options{
			Creds:  credentials.NewStaticV4(ak, sk, ""),
			Secure: true,
			Region: strings.TrimSpace(cfg.R2Region),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to init R2 client: " + err.Error()})
			return
		}

		ctx := context.Background()
		lo := minio.ListObjectsOptions{
			Prefix:    prefix,
			Recursive: true,
		}
		count := 0
		for obj := range client.ListObjects(ctx, bn, lo) {
			if obj.Err != nil {
				continue
			}
			key := obj.Key
			if limit > 0 && count >= limit {
				break
			}
			count++

			// 获取对象元数据（大小、类型）
			var contentType string
			var size int64 = obj.Size
			if st, err := client.StatObject(ctx, bn, key, minio.StatObjectOptions{}); err == nil {
				if st.Size > 0 {
					size = st.Size
				}
				if strings.TrimSpace(st.ContentType) != "" {
					contentType = st.ContentType
				}
			}
			if contentType == "" {
				contentType = inferMimeFromExt(path.Base(key))
			}

			publicURL := buildR2PublicURLFor(cfg, key, ep, bn, client)
			orig := path.Base(key)

			var exist models.Image
			err := database.DB.Where("r2_key = ? AND storage_provider = ?", key, "r2").First(&exist).Error
			if err == nil {
				if payload.Overwrite {
					exist.OriginalName = orig
					exist.FileName = orig
					exist.FileSize = size
					exist.MimeType = contentType
					exist.PublicURL = publicURL
					if e := database.DB.Save(&exist).Error; e == nil {
						updated++
					} else {
						// ignore update error
					}
				} else {
					skipped++
				}
				continue
			}
			if err != nil && !errorsIsRecordNotFound(err) {
				// 其他错误，跳过
				skipped++
				continue
			}

			img := &models.Image{
				UUID:            uuid.New().String(),
				OriginalName:    orig,
				FileName:        orig,
				FileSize:        size,
				MimeType:        contentType,
				Width:           0,
				Height:          0,
				R2Key:           key,
				PublicURL:       publicURL,
				StorageProvider: "r2",
				UploadIP:        "",
				UserAgent:       "sync",
				Uploader:        username,
			}
			if e := database.DB.Create(img).Error; e == nil {
				added++
			}
		}
	} else {
		// OSS
		ep := ensureSchemeLocal(strings.TrimSpace(cfg.OSSEndpoint))
		ak := strings.TrimSpace(cfg.OSSAccessKeyID)
		sk := strings.TrimSpace(cfg.OSSSecretAccessKey)
		bn := strings.TrimSpace(cfg.OSSBucketName)
		if ep == "" || ak == "" || sk == "" || bn == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "OSS config incomplete"})
			return
		}
		client, err := alioss.New(ep, ak, sk)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to init OSS client: " + err.Error()})
			return
		}
		bucket, err := client.Bucket(bn)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get OSS bucket: " + err.Error()})
			return
		}

		marker := ""
		count := 0
		for {
			lor, err := bucket.ListObjects(
				alioss.Prefix(prefix),
				alioss.Marker(marker),
				alioss.MaxKeys(1000),
			)
			if err != nil {
				break
			}
			for _, oi := range lor.Objects {
				if limit > 0 && count >= limit {
					break
				}
				count++

				key := oi.Key
				orig := path.Base(key)
				// 获取 meta
				var size int64 = int64(oi.Size)
				contentType := inferMimeFromExt(orig)
				if meta, err := bucket.GetObjectMeta(key); err == nil {
					if v := meta.Get("Content-Length"); v != "" {
						if sz, e := strconv.ParseInt(v, 10, 64); e == nil && sz > 0 {
							size = sz
						}
					}
					if v := meta.Get("Content-Type"); v != "" {
						contentType = v
					}
				}

				publicURL := buildOSSPublicURLFor(cfg, key, ep, bn, bucket)

				var exist models.Image
				err := database.DB.Where("r2_key = ? AND storage_provider = ?", key, "oss").First(&exist).Error
				if err == nil {
					if payload.Overwrite {
						exist.OriginalName = orig
						exist.FileName = orig
						exist.FileSize = size
						exist.MimeType = contentType
						exist.PublicURL = publicURL
						if e := database.DB.Save(&exist).Error; e == nil {
							updated++
						}
					} else {
						skipped++
					}
					continue
				}
				if err != nil && !errorsIsRecordNotFound(err) {
					skipped++
					continue
				}

				img := &models.Image{
					UUID:            uuid.New().String(),
					OriginalName:    orig,
					FileName:        orig,
					FileSize:        size,
					MimeType:        contentType,
					Width:           0,
					Height:          0,
					R2Key:           key,
					PublicURL:       publicURL,
					StorageProvider: "oss",
					UploadIP:        "",
					UserAgent:       "sync",
					Uploader:        username,
				}
				if e := database.DB.Create(img).Error; e == nil {
					added++
				}
			}
			if !lor.IsTruncated || (limit > 0 && count >= limit) {
				break
			}
			marker = lor.NextMarker
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"provider": pv,
			"added":    added,
			"updated":  updated,
			"skipped":  skipped,
		},
	})
}

// 辅助：判断是否为未找到
func errorsIsRecordNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound
}

// 辅助：根据扩展名推断 mime
func inferMimeFromExt(name string) string {
	n := strings.ToLower(name)
	switch {
	case strings.HasSuffix(n, ".jpg") || strings.HasSuffix(n, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(n, ".png"):
		return "image/png"
	case strings.HasSuffix(n, ".gif"):
		return "image/gif"
	case strings.HasSuffix(n, ".webp"):
		return "image/webp"
	case strings.HasSuffix(n, ".bmp"):
		return "image/bmp"
	case strings.HasSuffix(n, ".tiff") || strings.HasSuffix(n, ".tif"):
		return "image/tiff"
	default:
		return "application/octet-stream"
	}
}

// 辅助：规范化 R2 endpoint（剥离协议）
func normalizeEndpointLocal(ep string) string {
	ep = strings.TrimSpace(ep)
	if ep == "" {
		return ep
	}
	if strings.Contains(ep, "://") {
		if u, err := url.Parse(ep); err == nil && u.Host != "" {
			return u.Host
		}
	}
	return ep
}

// 辅助：确保 endpoint 带协议（OSS）
func ensureSchemeLocal(ep string) string {
	ep = strings.TrimSpace(ep)
	if ep == "" {
		return ep
	}
	if strings.HasPrefix(ep, "http://") || strings.HasPrefix(ep, "https://") {
		return ep
	}
	return "https://" + ep
}

// 构建 R2 公网 URL：
// - 若配置了 Public URL 前缀，直接拼接
// - 否则尝试生成临时签名链接（适用于私有桶）
// - 最后回退为 https://<endpoint>/<bucket>/<key>（要求公开可读）
func buildR2PublicURLFor(cfg models.StorageConfig, key, endpointHost, bucket string, client *minio.Client) string {
	prefix := strings.TrimSpace(cfg.R2PublicURL)
	if prefix != "" {
		if strings.HasSuffix(prefix, "/") {
			return prefix + key
		}
		return prefix + "/" + key
	}

	// 未配置公共前缀：尝试生成签名 URL（私有桶可访问）
	if client != nil && strings.TrimSpace(bucket) != "" {
		ctx := context.Background()
		expiry := 24 * time.Hour
		if u, err := client.PresignedGetObject(ctx, bucket, key, expiry, nil); err == nil {
			return u.String()
		}
	}

	// 回退直链（需公开可读）
	if endpointHost != "" && bucket != "" {
		return fmt.Sprintf("https://%s/%s/%s", endpointHost, bucket, key)
	}
	return key
}

// 构建 OSS 公网 URL：
// - 若配置了 Public URL 前缀，直接拼接
// - 否则尝试生成临时签名链接（适用于私有桶）
// - 最后回退 https://<bucket>.<host>/<key>（需公开可读）
func buildOSSPublicURLFor(cfg models.StorageConfig, key, endpoint, bucket string, bkt *alioss.Bucket) string {
	prefix := strings.TrimSpace(cfg.OSSPublicURL)
	if prefix != "" {
		if strings.HasSuffix(prefix, "/") {
			return prefix + key
		}
		return prefix + "/" + key
	}

	// 未配置公共前缀：尝试生成签名 URL（私有桶可访问）
	if bkt != nil {
		expiry := 24 * time.Hour
		if u, err := bkt.SignURL(key, "GET", int64(expiry/time.Second)); err == nil && strings.TrimSpace(u) != "" {
			return u
		}
	}

	// 回退直链（需公开可读）
	host := ""
	if u, err := url.Parse(endpoint); err == nil && u.Host != "" {
		host = u.Host
	} else {
		host = endpoint
	}
	if host != "" && strings.TrimSpace(bucket) != "" {
		return fmt.Sprintf("https://%s.%s/%s", bucket, host, key)
	}
	return key
}
