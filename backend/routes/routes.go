package routes

import (
	"net/http"
	"time"

	"image-host/config"
	"image-host/controllers"
	"image-host/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes() *gin.Engine {
	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	// 创建路由器
	r := gin.New()

	// 添加中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// CORS 配置
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 速率限制中间件
	r.Use(middleware.RateLimit())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"service":   "image-host",
		})
	})

	// API 路由组
	api := r.Group("/api/v1")
	{
		// 认证相关（公开）
		auth := api.Group("/auth")
		{
			auth.POST("/login", controllers.Auth.Login)
			auth.POST("/guest-login", controllers.Auth.GuestLogin)
		}

		// 受保护的路由
		protected := api.Group("")
		protected.Use(middleware.Auth())
		{
			// 用户信息与改密
			protected.GET("/auth/me", controllers.Auth.Me)
			protected.POST("/auth/change-password", controllers.Auth.ChangePassword)

			// 图片上传相关路由
			images := protected.Group("/images")
			{
				// 列表与删除
				images.GET("/", controllers.Upload.ListImages)
				images.DELETE("/:uuid", controllers.Upload.DeleteImage)

				// 上传图片
				images.POST("/upload", controllers.Upload.UploadImage)

				// 获取图片信息
				images.GET("/:uuid", controllers.Upload.GetImage)

				// 获取统计信息
				images.GET("/stats/summary", controllers.Upload.GetStats)
			}

			// 批量上传路由
			protected.POST("/batch-upload", controllers.Upload.BatchUpload)

			// 游客码管理（仅 root）
			guest := protected.Group("/guest-codes")
			{
				guest.POST("/", controllers.GuestCode.Create)
				guest.GET("/", controllers.GuestCode.List)
				guest.DELETE("/:id", controllers.GuestCode.Delete)
			}

			// 系统状态
			system := protected.Group("/system")
			{
				system.GET("/status", controllers.System.Status)
			}
		}
	}

	// 静态文件服务
	r.Static("/static", "./static")
	r.Static("/uploads", config.AppConfig.UploadPath)

	// 404 处理
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Endpoint not found",
			"code":  "NOT_FOUND",
		})
	})

	return r
}
