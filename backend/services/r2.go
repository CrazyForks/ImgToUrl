package services

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"image-host/config"

	"github.com/google/uuid"
)

type R2Service struct{}

var R2 *R2Service

// InitR2Service 初始化本地存储服务
func InitR2Service() {
	_ = os.MkdirAll(config.AppConfig.UploadPath, 0755)
	R2 = &R2Service{}
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

	// 直接使用本地存储（移除对 R2 的依赖）
	return r.uploadToLocal(fileBytes, key)
}

// DeleteFile 删除本地文件
func (r *R2Service) DeleteFile(key string) error {
	localPath := filepath.Join(config.AppConfig.UploadPath, key)
	if _, err := os.Stat(localPath); err == nil {
		if err := os.Remove(localPath); err != nil {
			return fmt.Errorf("failed to delete local file: %v", err)
		}
	}
	return nil
}

// GetFileURL 获取文件的公共 URL（相对路径）
func (r *R2Service) GetFileURL(key string) string {
	return fmt.Sprintf("/uploads/%s", key)
}

// uploadToLocal 本地存储备用方案
func (r *R2Service) uploadToLocal(fileBytes []byte, key string) (string, string, error) {
	// 创建本地存储目录
	localPath := filepath.Join(config.AppConfig.UploadPath, key)
	localDir := filepath.Dir(localPath)

	if err := os.MkdirAll(localDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create local directory: %v", err)
	}

	// 写入文件到本地
	if err := os.WriteFile(localPath, fileBytes, 0644); err != nil {
		return "", "", fmt.Errorf("failed to write local file: %v", err)
	}

	// 生成相对访问 URL，便于在任意域名/端口下工作
	localURL := fmt.Sprintf("/uploads/%s", key)

	return key, localURL, nil
}
