package services

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"image-host/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

type R2Service struct {
	s3Client *s3.S3
	bucket   string
	publicURL string
}

var R2 *R2Service

// InitR2Service 初始化 R2 服务
func InitR2Service() {
	cfg := config.AppConfig

	// 创建 AWS 会话
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.R2Region),
		Credentials: credentials.NewStaticCredentials(
			cfg.R2AccessKeyID,
			cfg.R2SecretAccessKey,
			"",
		),
		Endpoint:         aws.String(cfg.R2Endpoint),
		S3ForcePathStyle: aws.Bool(true),
	})

	if err != nil {
		panic(fmt.Sprintf("Failed to create R2 session: %v", err))
	}

	R2 = &R2Service{
		s3Client:  s3.New(sess),
		bucket:    cfg.R2BucketName,
		publicURL: cfg.R2PublicURL,
	}
}

// UploadFile 上传文件到 R2 (如果失败则使用本地存储)
func (r *R2Service) UploadFile(file multipart.File, header *multipart.FileHeader) (string, string, error) {
	// 生成唯一文件名
	uuid := uuid.New().String()
	ext := filepath.Ext(header.Filename)
	fileName := fmt.Sprintf("%s%s", uuid, ext)

	// 按日期组织文件路径
	now := time.Now()
	key := fmt.Sprintf("images/%d/%02d/%02d/%s", now.Year(), now.Month(), now.Day(), fileName)

	// 读取文件内容
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", "", fmt.Errorf("failed to read file: %v", err)
	}

	// 重置文件指针
	file.Seek(0, 0)

	// 尝试上传到 R2
	_, err = r.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(r.bucket),
		Key:           aws.String(key),
		Body:          bytes.NewReader(fileBytes),
		ContentType:   aws.String(header.Header.Get("Content-Type")),
		ContentLength: aws.Int64(header.Size),
		ACL:           aws.String("public-read"),
	})

	if err != nil {
		// R2上传失败，使用本地存储作为备用方案
		fmt.Printf("R2 upload failed, using local storage: %v\n", err)
		return r.uploadToLocal(fileBytes, key)
	}

	// 生成公共 URL
	publicURL := fmt.Sprintf("%s/%s", strings.TrimRight(r.publicURL, "/"), key)

	return key, publicURL, nil
}

// DeleteFile 从 R2 删除文件
func (r *R2Service) DeleteFile(key string) error {
	_, err := r.s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("failed to delete from R2: %v", err)
	}

	return nil
}

// GetFileURL 获取文件的公共 URL
func (r *R2Service) GetFileURL(key string) string {
	return fmt.Sprintf("%s/%s", strings.TrimRight(r.publicURL, "/"), key)
}

// uploadToLocal 本地存储备用方案
func (r *R2Service) uploadToLocal(fileBytes []byte, key string) (string, string, error) {
	// 创建本地存储目录
	localPath := filepath.Join("uploads", key)
	localDir := filepath.Dir(localPath)
	
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create local directory: %v", err)
	}
	
	// 写入文件到本地
	if err := os.WriteFile(localPath, fileBytes, 0644); err != nil {
		return "", "", fmt.Errorf("failed to write local file: %v", err)
	}
	
	// 生成本地访问URL
	localURL := fmt.Sprintf("http://localhost:8080/uploads/%s", key)
	
	return key, localURL, nil
}