package models

import (
	"time"

	"gorm.io/gorm"
)

// GuestCode 游客码
type GuestCode struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Code      string         `json:"code" gorm:"type:varchar(64);uniqueIndex;not null"`
	ExpiresAt *time.Time     `json:"expires_at" gorm:"index"` // nil 表示永久
	CreatedBy string         `json:"created_by" gorm:"type:varchar(64);not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (GuestCode) TableName() string {
	return "guest_codes"
}
