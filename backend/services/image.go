package services

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"strings"

	"github.com/disintegration/imaging"
)

type ImageService struct{}

var ImageSvc *ImageService

// InitImageService 初始化图片服务
func InitImageService() {
	ImageSvc = &ImageService{}
}

// ProcessImage 处理图片（压缩、获取尺寸等）
func (s *ImageService) ProcessImage(file multipart.File, header *multipart.FileHeader) (*ProcessedImage, error) {
	// 读取文件内容
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	// 重置文件指针
	file.Seek(0, 0)

	// 解码图片
	img, format, err := image.Decode(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	// 获取图片尺寸
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 压缩图片（如果需要）
	compressedBytes := fileBytes
	if header.Size > 1024*1024 { // 大于1MB时压缩
		compressedBytes, err = s.compressImage(img, format, 85) // 85%质量
		if err != nil {
			return nil, fmt.Errorf("failed to compress image: %v", err)
		}
	}

	// 生成缩略图
	thumbnailBytes, err := s.generateThumbnail(img, format, 300, 300)
	if err != nil {
		return nil, fmt.Errorf("failed to generate thumbnail: %v", err)
	}

	return &ProcessedImage{
		OriginalBytes:  fileBytes,
		CompressedBytes: compressedBytes,
		ThumbnailBytes: thumbnailBytes,
		Width:          width,
		Height:         height,
		Format:         format,
		MimeType:       header.Header.Get("Content-Type"),
	}, nil
}

// compressImage 压缩图片
func (s *ImageService) compressImage(img image.Image, format string, quality int) ([]byte, error) {
	var buf bytes.Buffer

	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
		if err != nil {
			return nil, err
		}
	case "png":
		// PNG 使用无损压缩，通过调整压缩级别
		encoder := &png.Encoder{CompressionLevel: png.BestCompression}
		err := encoder.Encode(&buf, img)
		if err != nil {
			return nil, err
		}
	default:
		// 默认使用 JPEG 格式
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// generateThumbnail 生成缩略图
func (s *ImageService) generateThumbnail(img image.Image, format string, maxWidth, maxHeight int) ([]byte, error) {
	// 调整图片大小，保持宽高比
	thumbnail := imaging.Fit(img, maxWidth, maxHeight, imaging.Lanczos)

	var buf bytes.Buffer

	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		err := jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: 80})
		if err != nil {
			return nil, err
		}
	case "png":
		err := png.Encode(&buf, thumbnail)
		if err != nil {
			return nil, err
		}
	default:
		// 默认使用 JPEG 格式
		err := jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: 80})
		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// ValidateImage 验证图片格式和大小
func (s *ImageService) ValidateImage(header *multipart.FileHeader, allowedTypes []string, maxSize int64) error {
	// 检查文件大小
	if header.Size > maxSize {
		return fmt.Errorf("file size exceeds limit: %d bytes", maxSize)
	}

	// 检查文件类型
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		return fmt.Errorf("missing content type")
	}

	for _, allowedType := range allowedTypes {
		if mimeType == allowedType {
			return nil
		}
	}

	return fmt.Errorf("unsupported file type: %s", mimeType)
}

// ProcessedImage 处理后的图片数据
type ProcessedImage struct {
	OriginalBytes   []byte
	CompressedBytes []byte
	ThumbnailBytes  []byte
	Width           int
	Height          int
	Format          string
	MimeType        string
}