package models

import (
	"time"

	"gorm.io/gorm"
)

// Image 图片模型
type Image struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	UUID        string         `json:"uuid" gorm:"type:varchar(36);uniqueIndex;not null"`
	OriginalName string        `json:"original_name" gorm:"not null"`
	FileName     string        `json:"file_name" gorm:"not null"`
	FileSize     int64         `json:"file_size" gorm:"not null"`
	MimeType     string        `json:"mime_type" gorm:"not null"`
	Width        int           `json:"width"`
	Height       int           `json:"height"`
	R2Key        string        `json:"r2_key" gorm:"not null"`
	PublicURL    string        `json:"public_url" gorm:"not null"`
	ThumbnailURL string        `json:"thumbnail_url"`
	UploadIP     string        `json:"upload_ip"`
	UserAgent    string        `json:"user_agent"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (Image) TableName() string {
	return "images"
}

// ImageStats 图片统计模型
type ImageStats struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Date        time.Time `json:"date" gorm:"uniqueIndex;not null"`
	TotalImages int64     `json:"total_images" gorm:"default:0"`
	TotalSize   int64     `json:"total_size" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ImageStats) TableName() string {
	return "image_stats"
}