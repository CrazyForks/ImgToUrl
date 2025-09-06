package database

import (
	"log"
	"time"

	"image-host/config"
	"image-host/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func ensureDefaultAdmin(db *gorm.DB) {
	var count int64
	if err := db.Model(&models.User{}).Where("username = ?", config.AppConfig.DefaultAdmin).Count(&count).Error; err != nil {
		log.Printf("failed to query users: %v", err)
		return
	}
	if count > 0 {
		return
	}
	// 创建默认管理员
	hashed, err := bcrypt.GenerateFromPassword([]byte(config.AppConfig.DefaultPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("failed to hash default password: %v", err)
		return
	}
	u := &models.User{
		Username:     config.AppConfig.DefaultAdmin,
		PasswordHash: string(hashed),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := db.Create(u).Error; err != nil {
		log.Printf("failed to create default admin: %v", err)
	} else {
		log.Printf("default admin user created: %s", u.Username)
	}
}
