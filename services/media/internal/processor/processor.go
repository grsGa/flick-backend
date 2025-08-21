package processor

import (
	"fmt"
	"mime"
	"path/filepath"
	"strings"

	"github.com/flick/backend/pkg/models"
)

// MediaProcessor 媒体处理器接口
type MediaProcessor interface {
	ProcessMedia(inputPath string, mediaType string, variants []string) (*models.MediaVariants, error)
}

// ProcessorManager 处理器管理器
type ProcessorManager struct {
	imageProcessor *ImageProcessor
	videoProcessor *VideoProcessor
	tempDir        string
}

// NewProcessorManager 创建处理器管理器
func NewProcessorManager(tempDir string) *ProcessorManager {
	return &ProcessorManager{
		imageProcessor: NewImageProcessor(tempDir),
		videoProcessor: NewVideoProcessor(tempDir),
		tempDir:        tempDir,
	}
}

// ProcessMedia 处理媒体文件
func (pm *ProcessorManager) ProcessMedia(inputPath string, mediaType string, variants []string) (*models.MediaVariants, error) {
	// 根据文件类型选择处理器
	switch {
	case strings.HasPrefix(mediaType, "image/"):
		if strings.Contains(mediaType, "gif") {
			return pm.imageProcessor.ProcessGIF(inputPath)
		}
		return pm.imageProcessor.ProcessImage(inputPath, variants)
	case strings.HasPrefix(mediaType, "video/"):
		return pm.videoProcessor.ProcessVideo(inputPath, variants)
	default:
		return nil, fmt.Errorf("unsupported media type: %s", mediaType)
	}
}

// GetMediaType 根据文件扩展名获取媒体类型
func (pm *ProcessorManager) GetMediaType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	mimeType := mime.TypeByExtension(ext)
	
	if mimeType == "" {
		// 手动处理一些常见类型
		switch ext {
		case ".jpg", ".jpeg":
			return "image/jpeg"
		case ".png":
			return "image/png"
		case ".gif":
			return "image/gif"
		case ".webp":
			return "image/webp"
		case ".mp4":
			return "video/mp4"
		case ".webm":
			return "video/webm"
		case ".mov":
			return "video/quicktime"
		case ".avi":
			return "video/x-msvideo"
		default:
			return "application/octet-stream"
		}
	}
	
	return mimeType
}

// GetDefaultVariants 根据媒体类型获取默认需要生成的版本
func (pm *ProcessorManager) GetDefaultVariants(mediaType string) []string {
	switch {
	case strings.HasPrefix(mediaType, "image/"):
		if strings.Contains(mediaType, "gif") {
			return []string{"thumbnail"} // GIF只生成缩略图
		}
		return []string{"thumbnail", "small", "medium", "large"}
	case strings.HasPrefix(mediaType, "video/"):
		return []string{"low_res", "mid_res", "high_res"}
	default:
		return []string{}
	}
}

// ValidateMediaFile 验证媒体文件
func (pm *ProcessorManager) ValidateMediaFile(filename string, size int64, mediaType string) error {
	// 文件大小限制
	const (
		MaxImageSize = 100 * 1024 * 1024  // 100MB
		MaxVideoSize = 500 * 1024 * 1024  // 500MB
	)

	switch {
	case strings.HasPrefix(mediaType, "image/"):
		if size > MaxImageSize {
			return fmt.Errorf("image file too large: %d bytes (max: %d)", size, MaxImageSize)
		}
	case strings.HasPrefix(mediaType, "video/"):
		if size > MaxVideoSize {
			return fmt.Errorf("video file too large: %d bytes (max: %d)", size, MaxVideoSize)
		}
	default:
		return fmt.Errorf("unsupported media type: %s", mediaType)
	}

	// 支持的文件类型
	supportedTypes := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"image/gif":       true,
		"image/webp":      true,
		"video/mp4":       true,
		"video/webm":      true,
		"video/quicktime": true,
		"video/x-msvideo": true,
	}

	if !supportedTypes[mediaType] {
		return fmt.Errorf("unsupported media type: %s", mediaType)
	}

	return nil
}

// EstimateProcessingTime 估算处理时间（秒）
func (pm *ProcessorManager) EstimateProcessingTime(mediaType string, size int64, variants []string) int {
	baseTime := 5 // 基础处理时间5秒

	switch {
	case strings.HasPrefix(mediaType, "image/"):
		// 图片处理相对较快
		return baseTime + len(variants)*2
	case strings.HasPrefix(mediaType, "video/"):
		// 视频处理较慢，根据文件大小估算
		sizeInMB := size / (1024 * 1024)
		return baseTime + int(sizeInMB)*3 + len(variants)*10
	default:
		return baseTime
	}
}
