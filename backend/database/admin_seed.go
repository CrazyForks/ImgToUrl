package database

import (
	"log"

	"image-host/config"
	"image-host/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ensureDefaultAdmin 与 database.InitDatabase 中的调用一致
func ensureDefaultAdmin(db *gorm.DB) {
	var count int64
	if err := db.Model(&models.User{}).Where("username = ?", config.AppConfig.DefaultAdmin).Count(&count).Error; err != nil {
		log.Printf("check admin failed: %v", err)
		return
	}
	if count > 0 {
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(config.AppConfig.DefaultPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("hash default password failed: %v", err)
		return
	}
	admin := &models.User{
		Username:     config.AppConfig.DefaultAdmin,
		PasswordHash: string(hash),
	}
	if err := db.Create(admin).Error; err != nil {
		log.Printf("create default admin failed: %v", err)
		return
	}
	log.Printf("Default admin user created: %s", admin.Username)
}
