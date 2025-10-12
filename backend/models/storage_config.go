package models

import (
	"time"

	"gorm.io/gorm"
)

// StorageConfig 存储配置（单行表）
type StorageConfig struct {
	ID uint `json:"id" gorm:"primaryKey"`

	// provider: local / r2 / oss
	Provider string `json:"provider" gorm:"type:varchar(16);not null;default:local"`

	// R2 配置（S3 兼容）
	R2AccountID       string `json:"r2_account_id" gorm:"type:varchar(128)"`
	R2AccessKeyID     string `json:"r2_access_key_id" gorm:"type:varchar(256)"`
	R2SecretAccessKey string `json:"r2_secret_access_key" gorm:"type:varchar(256)"`
	R2BucketName      string `json:"r2_bucket_name" gorm:"type:varchar(128)"`
	R2Endpoint        string `json:"r2_endpoint" gorm:"type:varchar(256)"`
	R2Region          string `json:"r2_region" gorm:"type:varchar(64);default:auto"`
	R2PublicURL       string `json:"r2_public_url" gorm:"type:varchar(512)"`

	// OSS 配置（阿里云）
	OSSAccessKeyID     string `json:"oss_access_key_id" gorm:"type:varchar(256)"`
	OSSSecretAccessKey string `json:"oss_secret_access_key" gorm:"type:varchar(256)"`
	OSSBucketName      string `json:"oss_bucket_name" gorm:"type:varchar(128)"`
	OSSEndpoint        string `json:"oss_endpoint" gorm:"type:varchar(256)"`
	OSSRegion          string `json:"oss_region" gorm:"type:varchar(64)"`
	OSSPublicURL       string `json:"oss_public_url" gorm:"type:varchar(512)"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (StorageConfig) TableName() string {
	return "storage_config"
}