package main

import (
	"log"

	"image-host/config"
	"image-host/controllers"
	"image-host/database"
	"image-host/routes"
	"image-host/services"
)

func main() {
	// 加载配置
	config.LoadConfig()
	log.Println("Configuration loaded successfully")

	// 初始化数据库
	database.InitDatabase()
	log.Println("Database initialized successfully")

	// 初始化服务
	services.InitR2Service()
	log.Println("R2 service initialized successfully")

	services.InitImageService()
	log.Println("Image service initialized successfully")

	// 初始化控制器
	controllers.InitUploadController()
	controllers.InitSystemController()
	log.Println("Controllers initialized successfully")

	// 设置路由
	r := routes.SetupRoutes()
	log.Println("Routes configured successfully")

	// 启动游客码过期清理任务
	services.Guest.StartCleanupJob()

	// 启动服务器
	port := config.AppConfig.Port
	log.Printf("Server starting on port %s", port)
	log.Printf("Health check: http://localhost:%s/health", port)
	log.Printf("API endpoint: http://localhost:%s/api/v1", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
