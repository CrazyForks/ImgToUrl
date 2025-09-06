package services

import (
	"fmt"
	"math/rand"
	"time"

	"image-host/database"
	"image-host/models"
)

type GuestService struct{}

var Guest = &GuestService{}

func randomCode(n int) string {
	const letters = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// GenerateCode 生成游客码
func (s *GuestService) GenerateCode(creator string, expiresAt *time.Time) (*models.GuestCode, error) {
	code := randomCode(10)
	g := &models.GuestCode{
		Code:      code,
		ExpiresAt: expiresAt,
		CreatedBy: creator,
	}
	if err := database.DB.Create(g).Error; err != nil {
		return nil, err
	}
	return g, nil
}

// DeleteCodeAndImages 删除指定游客码及其图片
func (s *GuestService) DeleteCodeAndImages(id string) error {
	var gc models.GuestCode
	if err := database.DB.Where("id = ?", id).First(&gc).Error; err != nil {
		return err
	}
	uploader := fmt.Sprintf("guest:%d", gc.ID)

	// 删除文件与记录
	var images []models.Image
	if err := database.DB.Where("uploader = ?", uploader).Find(&images).Error; err == nil {
		for _, img := range images {
			if img.R2Key != "" {
				_ = R2.DeleteFile(img.R2Key)
			}
			_ = database.DB.Unscoped().Delete(&img).Error
		}
	}
	// 删除游客码
	return database.DB.Unscoped().Delete(&gc).Error
}

// CleanupExpired 定时清理过期游客码与其图片
func (s *GuestService) CleanupExpired() {
	now := time.Now()
	var expired []models.GuestCode
	if err := database.DB.Where("expires_at IS NOT NULL AND expires_at < ?", now).Find(&expired).Error; err != nil {
		return
	}
	for _, gc := range expired {
		_ = s.DeleteCodeAndImages(fmt.Sprintf("%d", gc.ID))
	}
}

// StartCleanupJob 每小时清理一次
func (s *GuestService) StartCleanupJob() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			s.CleanupExpired()
		}
	}()
}
