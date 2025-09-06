package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	// 服务器配置
	Port string

	// 鉴权配置
	JWTSecret       string
	JWTExpireHours  int
	DefaultAdmin    string
	DefaultPassword string

	// 数据库配置
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Cloudflare R2 配置
	R2AccountID       string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2BucketName      string
	R2Endpoint        string
	R2Region          string
	R2PublicURL       string

	// Redis 配置
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	// 上传配置
	MaxFileSize  int64
	AllowedTypes []string
	UploadPath   string
}

var AppConfig *Config

func LoadConfig() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	maxFileSize, _ := strconv.ParseInt(getEnv("MAX_FILE_SIZE", "10485760"), 10, 64) // 10MB
	jwtExpireHours, _ := strconv.Atoi(getEnv("JWT_EXPIRE_HOURS", "72"))

	// 端口与允许类型（从环境变量解析）
	port := getEnv("SERVER_PORT", getEnv("PORT", "8080"))
	allowedTypesStr := getEnv("ALLOWED_TYPES", "image/jpeg,image/png,image/gif,image/webp")
	var allowedTypes []string
	for _, t := range strings.Split(allowedTypesStr, ",") {
		tt := strings.TrimSpace(t)
		if tt != "" {
			allowedTypes = append(allowedTypes, tt)
		}
	}

	AppConfig = &Config{
		// 服务器配置
		Port: port,

		// 鉴权配置
		JWTSecret:       getEnv("JWT_SECRET", "change_me_secret"),
		JWTExpireHours:  jwtExpireHours,
		DefaultAdmin:    getEnv("DEFAULT_ADMIN", "root"),
		DefaultPassword: getEnv("DEFAULT_PASSWORD", "123456"),

		// 数据库配置
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "image_host"),

		// Cloudflare R2 配置（保留结构以兼容，但不强制使用）
		R2AccountID:       getEnv("R2_ACCOUNT_ID", ""),
		R2AccessKeyID:     getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretAccessKey: getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2BucketName:      getEnv("R2_BUCKET_NAME", ""),
		R2Endpoint:        getEnv("R2_ENDPOINT", ""),
		R2Region:          getEnv("R2_REGION", "auto"),
		R2PublicURL:       getEnv("R2_PUBLIC_URL", ""),

		// Redis 配置
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,

		// 上传配置
		MaxFileSize:  maxFileSize,
		AllowedTypes: allowedTypes,
		UploadPath:   getEnv("UPLOAD_PATH", "./uploads"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
