package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
	"log"
	"net/url"

	"image-host/config"
	"image-host/database"
	"image-host/models"

	alioss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// R2Service 同时支持本地与 Cloudflare R2（S3 兼容）
type R2Service struct {
	provider        string           // local / r2 / oss
	minioClient     *minio.Client    // R2 客户端（S3 兼容）
	ossClient       *alioss.Client   // 阿里云 OSS 客户端
	bucketName      string           // 存储桶名（R2/OSS 复用）
	publicURLPrefix string           // 公共 URL 前缀（例如 https://cdn.example.com）
	endpoint        string           // 端点：R2 为 host（无协议）；OSS 为完整 endpoint（含协议）
}

var R2 *R2Service

// Provider 返回当前存储渠道（local 或 r2）
func (r *R2Service) Provider() string {
	return r.provider
}

// InitR2Service 初始化存储服务（默认创建本地目录），并尝试加载数据库/环境中的 R2 配置
func InitR2Service() {
	_ = os.MkdirAll(config.AppConfig.UploadPath, 0755)
	R2 = &R2Service{}
	R2.ReloadConfig()
}

// ReloadConfig 尝试从数据库的 storage_config 读取配置；若无则回退到环境变量；最终默认为 local
func (r *R2Service) ReloadConfig() {
	// 数据库优先
	if database.DB != nil {
		var cfg models.StorageConfig
		if err := database.DB.First(&cfg).Error; err == nil {
			r.applyCfgFromModel(&cfg)
			log.Printf("storage reload (db): provider=%s bucket=%s endpoint=%s public_prefix=%s", r.provider, r.bucketName, r.endpoint, r.publicURLPrefix)
			return
		}
	}

	// 环境变量回退（优先 R2，再尝试 OSS；否则使用本地）
	// R2
	epR2 := strings.TrimSpace(config.AppConfig.R2Endpoint)
	akR2 := strings.TrimSpace(config.AppConfig.R2AccessKeyID)
	skR2 := strings.TrimSpace(config.AppConfig.R2SecretAccessKey)
	bnR2 := strings.TrimSpace(config.AppConfig.R2BucketName)
	pubR2 := strings.TrimSpace(config.AppConfig.R2PublicURL)

	if epR2 != "" && akR2 != "" && skR2 != "" && bnR2 != "" {
		r.provider = "r2"
		r.endpoint = normalizeEndpoint(epR2)
		r.bucketName = bnR2
		r.publicURLPrefix = pubR2
		client, err := minio.New(r.endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(akR2, skR2, ""),
			Secure: true, // R2 统一使用 https
			Region: config.AppConfig.R2Region,
		})
		if err != nil {
			log.Printf("R2 init failed with env (endpoint=%s): %v; continue to try OSS or fallback local", r.endpoint, err)
			r.provider = "local"
			r.minioClient = nil
		} else {
			r.minioClient = client
			log.Printf("storage reload (env): provider=%s bucket=%s endpoint=%s public_prefix=%s", r.provider, r.bucketName, r.endpoint, r.publicURLPrefix)
			return
		}
	}

	// OSS
	epOSS := strings.TrimSpace(config.AppConfig.OSSEndpoint)
	akOSS := strings.TrimSpace(config.AppConfig.OSSAccessKeyID)
	skOSS := strings.TrimSpace(config.AppConfig.OSSSecretAccessKey)
	bnOSS := strings.TrimSpace(config.AppConfig.OSSBucketName)
	pubOSS := strings.TrimSpace(config.AppConfig.OSSPublicURL)

	if epOSS != "" && akOSS != "" && skOSS != "" && bnOSS != "" {
		r.provider = "oss"
		r.endpoint = ensureScheme(epOSS)
		r.bucketName = bnOSS
		r.publicURLPrefix = pubOSS
		if client, err := alioss.New(r.endpoint, akOSS, skOSS); err == nil {
			r.ossClient = client
			log.Printf("storage reload (env): provider=%s bucket=%s endpoint=%s public_prefix=%s", r.provider, r.bucketName, r.endpoint, r.publicURLPrefix)
			return
		} else {
			log.Printf("OSS init failed with env (endpoint=%s): %v; fallback local", r.endpoint, err)
		}
	}

	// 默认本地
	r.provider = "local"
	r.minioClient = nil
	r.ossClient = nil
	r.bucketName = ""
	r.endpoint = ""
	r.publicURLPrefix = ""
	log.Printf("storage reload (default): provider=%s", r.provider)
}

// applyCfgFromModel 应用数据库模型配置并尝试初始化存储客户端
func (r *R2Service) applyCfgFromModel(cfg *models.StorageConfig) {
	r.provider = strings.ToLower(strings.TrimSpace(cfg.Provider))
	if r.provider == "" {
		r.provider = "local"
	}

	// reset clients
	r.minioClient = nil
	r.ossClient = nil
	r.bucketName = ""
	r.publicURLPrefix = ""
	r.endpoint = ""

	switch r.provider {
	case "r2":
		r.bucketName = strings.TrimSpace(cfg.R2BucketName)
		r.publicURLPrefix = strings.TrimSpace(cfg.R2PublicURL)
		epRaw := strings.TrimSpace(cfg.R2Endpoint)
		if epRaw != "" && cfg.R2AccessKeyID != "" && cfg.R2SecretAccessKey != "" && r.bucketName != "" {
			r.endpoint = normalizeEndpoint(epRaw)
			client, err := minio.New(r.endpoint, &minio.Options{
				Creds:  credentials.NewStaticV4(cfg.R2AccessKeyID, cfg.R2SecretAccessKey, ""),
				Secure: true, // R2 使用 https
				Region: cfg.R2Region,
			})
			if err == nil {
				r.minioClient = client
				return
			}
			log.Printf("R2 init failed with DB config (endpoint=%s): %v; falling back to local", r.endpoint, err)
		}
	case "oss":
		r.bucketName = strings.TrimSpace(cfg.OSSBucketName)
		r.publicURLPrefix = strings.TrimSpace(cfg.OSSPublicURL)
		epRaw := strings.TrimSpace(cfg.OSSEndpoint)
		ak := strings.TrimSpace(cfg.OSSAccessKeyID)
		sk := strings.TrimSpace(cfg.OSSSecretAccessKey)
		if epRaw != "" && ak != "" && sk != "" && r.bucketName != "" {
			ep := ensureScheme(epRaw)
			r.endpoint = ep
			if client, err := alioss.New(ep, ak, sk); err == nil {
				r.ossClient = client
				return
			} else {
				log.Printf("OSS init failed with DB config (endpoint=%s): %v; falling back to local", ep, err)
			}
		}
	default:
		// local
	}

	// 客户端未就绪则回退到本地
	r.provider = "local"
	r.minioClient = nil
	r.ossClient = nil
}

// UploadFile 上传文件到当前配置的存储渠道（local 或 r2）
func (r *R2Service) UploadFile(file multipart.File, header *multipart.FileHeader) (string, string, error) {
	// 生成唯一文件名与路径
	u := uuid.New().String()
	ext := filepath.Ext(header.Filename)
	fileName := fmt.Sprintf("%s%s", u, ext)

	now := time.Now()
	key := fmt.Sprintf("images/%d/%02d/%02d/%s", now.Year(), now.Month(), now.Day(), fileName)

	// 读取文件内容（用于 local 与 r2 两种方式）
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", "", fmt.Errorf("failed to read file: %v")
	}
	_, _ = file.Seek(0, 0)

	// 根据渠道上传
	log.Printf("upload start: provider=%s bucket=%s endpoint=%s public_prefix=%s key=%s", r.provider, r.bucketName, r.endpoint, r.publicURLPrefix, key)
	if r.provider == "oss" && r.ossClient != nil && r.bucketName != "" {
		if err := r.uploadToOSS(fileBytes, key, header.Header.Get("Content-Type")); err != nil {
			// OSS 上传失败回退至本地
			log.Printf("OSS upload error: %v; falling back to local", err)
			return r.uploadToLocal(fileBytes, key)
		}
		url := r.buildOSSURL(key)
		log.Printf("upload done (oss): key=%s url=%s", key, url)
		return key, url, nil
	}
	if r.provider == "r2" && r.minioClient != nil && r.bucketName != "" {
		if err := r.uploadToR2(fileBytes, key, header.Header.Get("Content-Type")); err != nil {
			// R2 上传失败回退至本地
			log.Printf("R2 upload error: %v; falling back to local", err)
			return r.uploadToLocal(fileBytes, key)
		}
		url := r.buildPublicURL(key)
		log.Printf("upload done (r2): key=%s url=%s", key, url)
		return key, url, nil
	}
	
	// 默认本地存储
	k, u, err := r.uploadToLocal(fileBytes, key)
	if err == nil {
		log.Printf("upload done (local): key=%s url=%s", k, u)
	}
	return k, u, err
}

// DeleteFile 删除文件（根据渠道删除本地或对象存储中的对象）
func (r *R2Service) DeleteFile(key string) error {
	if r.provider == "oss" && r.ossClient != nil && r.bucketName != "" {
		if bucket, err := r.ossClient.Bucket(r.bucketName); err == nil {
			_ = bucket.DeleteObject(key)
		}
	}
	if r.provider == "r2" && r.minioClient != nil && r.bucketName != "" {
		ctx := context.Background()
		err := r.minioClient.RemoveObject(ctx, r.bucketName, key, minio.RemoveObjectOptions{})
		if err != nil {
			// 忽略错误；继续尝试删除本地（如果存在）
		}
	}

	// 本地删除（忽略不存在错误）
	localPath := filepath.Join(config.AppConfig.UploadPath, key)
	if _, err := os.Stat(localPath); err == nil {
		if err := os.Remove(localPath); err != nil {
			return fmt.Errorf("failed to delete local file: %v", err)
		}
	}
	return nil
}

// DeleteFileByProvider 根据图片的存储渠道删除文件（独立于当前运行时渠道）
func (r *R2Service) DeleteFileByProvider(key string, provider string) error {
	provider = strings.ToLower(strings.TrimSpace(provider))

	// R2：优先尝试删除 R2 对象
	if provider == "r2" {
		var cfg models.StorageConfig
		if err := database.DB.First(&cfg).Error; err == nil {
			ep := normalizeEndpoint(strings.TrimSpace(cfg.R2Endpoint))
			ak := strings.TrimSpace(cfg.R2AccessKeyID)
			sk := strings.TrimSpace(cfg.R2SecretAccessKey)
			bn := strings.TrimSpace(cfg.R2BucketName)
			if ep != "" && ak != "" && sk != "" && bn != "" {
				if client, err := minio.New(ep, &minio.Options{
					Creds:  credentials.NewStaticV4(ak, sk, ""),
					Secure: true,
					Region: strings.TrimSpace(cfg.R2Region),
				}); err == nil {
					ctx := context.Background()
					_ = client.RemoveObject(ctx, bn, key, minio.RemoveObjectOptions{})
				}
			}
		}
	}

	// OSS：尝试删除 OSS 对象
	if provider == "oss" {
		var cfg models.StorageConfig
		if err := database.DB.First(&cfg).Error; err == nil {
			ep := ensureScheme(strings.TrimSpace(cfg.OSSEndpoint))
			ak := strings.TrimSpace(cfg.OSSAccessKeyID)
			sk := strings.TrimSpace(cfg.OSSSecretAccessKey)
			bn := strings.TrimSpace(cfg.OSSBucketName)
			if ep != "" && ak != "" && sk != "" && bn != "" {
				if client, err := alioss.New(ep, ak, sk); err == nil {
					if bucket, err := client.Bucket(bn); err == nil {
						_ = bucket.DeleteObject(key)
					}
				}
			}
		}
	}

	// 始终尝试删除本地文件（忽略不存在错误）
	localPath := filepath.Join(config.AppConfig.UploadPath, key)
	if _, err := os.Stat(localPath); err == nil {
		if err := os.Remove(localPath); err != nil {
			return fmt.Errorf("failed to delete local file: %v", err)
		}
	}
	return nil
}

// GetFileURL 获取公共访问 URL（local: /uploads/...；r2/oss: 使用各自策略拼接）
func (r *R2Service) GetFileURL(key string) string {
	if r.provider == "oss" {
		return r.buildOSSURL(key)
	}
	if r.provider == "r2" {
		return r.buildPublicURL(key)
	}
	return fmt.Sprintf("/uploads/%s", key)
}

// 上传到本地
func (r *R2Service) uploadToLocal(fileBytes []byte, key string) (string, string, error) {
	localPath := filepath.Join(config.AppConfig.UploadPath, key)
	localDir := filepath.Dir(localPath)

	if err := os.MkdirAll(localDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create local directory: %v", err)
	}

	if err := os.WriteFile(localPath, fileBytes, 0644); err != nil {
		return "", "", fmt.Errorf("failed to write local file: %v", err)
	}

	localURL := fmt.Sprintf("/uploads/%s", key)
	return key, localURL, nil
}

// 上传到 R2（S3 兼容）
func (r *R2Service) uploadToR2(fileBytes []byte, key string, contentType string) error {
	if r.minioClient == nil || r.bucketName == "" {
		return fmt.Errorf("r2 client not initialized")
	}
	ctx := context.Background()

	reader := bytes.NewReader(fileBytes)
	_, err := r.minioClient.PutObject(ctx, r.bucketName, key, reader, int64(len(fileBytes)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

// 构建公共 URL（针对 R2）
func (r *R2Service) buildPublicURL(key string) string {
	prefix := strings.TrimSpace(r.publicURLPrefix)
	if prefix == "" {
		// 未配置公共前缀：优先生成临时签名链接（适用于私有桶）
		if r.minioClient != nil && strings.TrimSpace(r.bucketName) != "" {
			ctx := context.Background()
			expiry := 24 * time.Hour // 可按需调整有效期
			if u, err := r.minioClient.PresignedGetObject(ctx, r.bucketName, key, expiry, nil); err == nil {
				return u.String()
			}
		}
		// 签名失败则回退为 https://<endpoint>/<bucket>/<key>（要求桶或自定义域公开可读）
		ep := strings.TrimSpace(r.endpoint)
		bn := strings.TrimSpace(r.bucketName)
		if ep != "" && bn != "" {
			return fmt.Sprintf("https://%s/%s/%s", ep, bn, key)
		}
		// 参数缺失时返回对象键
		return key
	}
	if strings.HasSuffix(prefix, "/") {
		return prefix + key
	}
	return prefix + "/" + key
}

// 构建公共 URL（针对 OSS）
func (r *R2Service) buildOSSURL(key string) string {
	prefix := strings.TrimSpace(r.publicURLPrefix)
	if prefix != "" {
		if strings.HasSuffix(prefix, "/") {
			return prefix + key
		}
		return prefix + "/" + key
	}

	// 未配置公共前缀：优先生成临时签名链接（适用于私有桶）
	if r.ossClient != nil && strings.TrimSpace(r.bucketName) != "" {
		if bucket, err := r.ossClient.Bucket(r.bucketName); err == nil {
			expiry := 24 * time.Hour // 有效期可按需调整
			if u, err := bucket.SignURL(key, "GET", int64(expiry/time.Second)); err == nil && strings.TrimSpace(u) != "" {
				return u
			}
		}
	}

	// 签名失败则回退为 https://<bucket>.<host>/<key>（要求桶或自定义域公开可读）
	host := ossHostFromEndpoint(r.endpoint)
	if host != "" && strings.TrimSpace(r.bucketName) != "" {
		return fmt.Sprintf("https://%s.%s/%s", r.bucketName, host, key)
	}
	// 参数缺失时返回对象键
	return key
}

// 上传到 OSS
func (r *R2Service) uploadToOSS(fileBytes []byte, key string, contentType string) error {
	if r.ossClient == nil || r.bucketName == "" {
		return fmt.Errorf("oss client not initialized")
	}
	bucket, err := r.ossClient.Bucket(r.bucketName)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(fileBytes)
	opts := []alioss.Option{}
	if ct := strings.TrimSpace(contentType); ct != "" {
		opts = append(opts, alioss.ContentType(ct))
	}
	return bucket.PutObject(key, reader, opts...)
}

// 规范化 OSS endpoint，确保包含协议
func ensureScheme(ep string) string {
	ep = strings.TrimSpace(ep)
	if ep == "" {
		return ep
	}
	if strings.HasPrefix(ep, "http://") || strings.HasPrefix(ep, "https://") {
		return ep
	}
	return "https://" + ep
}

// 提取 endpoint 的 host
func ossHostFromEndpoint(ep string) string {
	ep = strings.TrimSpace(ep)
	if ep == "" {
		return ""
	}
	if !strings.Contains(ep, "://") {
		return ep
	}
	if u, err := url.Parse(ep); err == nil {
		return u.Host
	}
	return ep
}

// normalizeEndpoint 规范化 R2 端点（剥离协议，提取 host）
func normalizeEndpoint(ep string) string {
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
