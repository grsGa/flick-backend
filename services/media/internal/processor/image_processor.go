package processor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/flick/backend/pkg/models"
)

// ImageProcessor 图片处理器
type ImageProcessor struct {
	tempDir string
}

// NewImageProcessor 创建图片处理器
func NewImageProcessor(tempDir string) *ImageProcessor {
	return &ImageProcessor{
		tempDir: tempDir,
	}
}

// ProcessImage 处理图片生成多版本
func (p *ImageProcessor) ProcessImage(inputPath string, variants []string) (*models.MediaVariants, error) {
	fmt.Printf("[IMAGE PROCESSOR] Starting ProcessImage for: %s, variants: %v\n", inputPath, variants)
	result := &models.MediaVariants{}
	
	// 获取原图信息
	fmt.Printf("[IMAGE PROCESSOR] Getting image info for: %s\n", inputPath)
	originalInfo, err := p.getImageInfo(inputPath)
	if err != nil {
		fmt.Printf("[IMAGE PROCESSOR] Failed to get image info: %v\n", err)
		return nil, fmt.Errorf("failed to get image info: %w", err)
	}
	fmt.Printf("[IMAGE PROCESSOR] Original image info: %dx%d, size: %d bytes\n", originalInfo.Width, originalInfo.Height, originalInfo.Size)

	// 生成各种尺寸版本
	for _, variant := range variants {
		fmt.Printf("[IMAGE PROCESSOR] Processing variant: %s\n", variant)
		var targetWidth int
		switch variant {
		case "thumbnail":
			targetWidth = 150
		case "small":
			targetWidth = 300
		case "medium":
			targetWidth = 600
		case "large":
			targetWidth = 1200
		default:
			fmt.Printf("[IMAGE PROCESSOR] Unknown variant: %s, skipping\n", variant)
			continue
		}

		// 如果原图小于目标尺寸，跳过
		if originalInfo.Width <= int32(targetWidth) {
			fmt.Printf("[IMAGE PROCESSOR] Original width (%d) <= target width (%d), skipping %s\n", originalInfo.Width, targetWidth, variant)
			continue
		}

		outputPath := p.getVariantPath(inputPath, variant)
		fmt.Printf("[IMAGE PROCESSOR] Resizing image for %s: %s -> %s (width: %d)\n", variant, inputPath, outputPath, targetWidth)
		err := p.resizeImage(inputPath, outputPath, targetWidth)
		if err != nil {
			fmt.Printf("[IMAGE PROCESSOR] Failed to resize image for %s: %v\n", variant, err)
			return nil, fmt.Errorf("failed to resize image for %s: %w", variant, err)
		}
		fmt.Printf("[IMAGE PROCESSOR] Successfully resized image for %s\n", variant)

		// 获取生成图片信息
		variantInfo, err := p.getImageInfo(outputPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get variant info: %w", err)
		}

		// 设置对应的variant
		mediaVariant := &models.MediaVariant{
			URL:    outputPath, // 这里应该是MinIO的URL，暂时用本地路径
			Width:  variantInfo.Width,
			Height: variantInfo.Height,
			Size:   variantInfo.Size,
		}

		switch variant {
		case "thumbnail":
			result.Thumbnail = mediaVariant
		case "small":
			result.Small = mediaVariant
		case "medium":
			result.Medium = mediaVariant
		case "large":
			result.Large = mediaVariant
		}
	}

	// 设置原图信息
	result.Original = &models.MediaVariant{
		URL:    inputPath,
		Width:  originalInfo.Width,
		Height: originalInfo.Height,
		Size:   originalInfo.Size,
	}

	return result, nil
}

// resizeImage 使用ImageMagick调整图片大小
func (p *ImageProcessor) resizeImage(inputPath, outputPath string, width int) error {
	fmt.Printf("[IMAGE PROCESSOR] Starting resizeImage: %s -> %s (width: %d)\n", inputPath, outputPath, width)
	
	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		fmt.Printf("[IMAGE PROCESSOR] Failed to create output directory: %v\n", err)
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 使用ImageMagick的convert命令
	// -resize 600x> 表示只在宽度大于600时才缩放，保持宽高比
	cmd := exec.Command("convert", inputPath, "-resize", fmt.Sprintf("%dx>", width), "-quality", "85", outputPath)
	
	fmt.Printf("[IMAGE PROCESSOR] Running ImageMagick command: %s\n", cmd.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[IMAGE PROCESSOR] ImageMagick command failed: %v\n", err)
		fmt.Printf("[IMAGE PROCESSOR] ImageMagick output: %s\n", string(output))
		return fmt.Errorf("imagemagick convert failed: %w (output: %s)", err, string(output))
	}
	
	fmt.Printf("[IMAGE PROCESSOR] ImageMagick command completed successfully\n")
	return nil
}

// getImageInfo 获取图片信息
func (p *ImageProcessor) getImageInfo(imagePath string) (*models.MediaVariant, error) {
	fmt.Printf("[IMAGE PROCESSOR] Getting image info for: %s\n", imagePath)
	// 使用ImageMagick的identify命令获取图片信息
	cmd := exec.Command("identify", "-format", "%w %h %B", imagePath)
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("[IMAGE PROCESSOR] ImageMagick identify failed: %v\n", err)
		return nil, fmt.Errorf("imagemagick identify failed: %w", err)
	}
	fmt.Printf("[IMAGE PROCESSOR] ImageMagick identify output: %s\n", string(output))

	parts := strings.Fields(string(output))
	if len(parts) != 3 {
		return nil, fmt.Errorf("unexpected identify output: %s", string(output))
	}

	width, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid width: %s", parts[0])
	}

	height, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid height: %s", parts[1])
	}

	size, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid size: %s", parts[2])
	}

	return &models.MediaVariant{
		Width:  int32(width),
		Height: int32(height),
		Size:   size,
	}, nil
}

// getVariantPath 生成变体文件路径
func (p *ImageProcessor) getVariantPath(originalPath, variant string) string {
	dir := filepath.Dir(originalPath)
	ext := filepath.Ext(originalPath)
	name := strings.TrimSuffix(filepath.Base(originalPath), ext)
	return filepath.Join(dir, fmt.Sprintf("%s_%s%s", name, variant, ext))
}

// ProcessGIF 处理GIF文件
func (p *ImageProcessor) ProcessGIF(inputPath string) (*models.MediaVariants, error) {
	result := &models.MediaVariants{}

	// 获取GIF信息
	originalInfo, err := p.getImageInfo(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get GIF info: %w", err)
	}

	// 生成首帧缩略图
	thumbnailPath := p.getVariantPath(inputPath, "thumbnail")
	err = p.extractGIFFrame(inputPath, thumbnailPath, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to extract GIF frame: %w", err)
	}

	// 调整缩略图大小
	err = p.resizeImage(thumbnailPath, thumbnailPath, 150)
	if err != nil {
		return nil, fmt.Errorf("failed to resize GIF thumbnail: %w", err)
	}

	thumbnailInfo, err := p.getImageInfo(thumbnailPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get thumbnail info: %w", err)
	}

	result.Thumbnail = &models.MediaVariant{
		URL:    thumbnailPath,
		Width:  thumbnailInfo.Width,
		Height: thumbnailInfo.Height,
		Size:   thumbnailInfo.Size,
	}

	// 转换为MP4（可选）
	mp4Path := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + ".mp4"
	err = p.convertGIFToMP4(inputPath, mp4Path)
	if err == nil {
		// 如果转换成功，设置为预览版本
		mp4Info, _ := p.getVideoInfo(mp4Path)
		if mp4Info != nil {
			result.Preview = mp4Info
		}
	}

	// 设置原GIF信息
	result.Original = &models.MediaVariant{
		URL:    inputPath,
		Width:  originalInfo.Width,
		Height: originalInfo.Height,
		Size:   originalInfo.Size,
	}

	return result, nil
}

// extractGIFFrame 提取GIF指定帧
func (p *ImageProcessor) extractGIFFrame(inputPath, outputPath string, frameIndex int) error {
	cmd := exec.Command("convert", fmt.Sprintf("%s[%d]", inputPath, frameIndex), outputPath)
	return cmd.Run()
}

// convertGIFToMP4 将GIF转换为MP4
func (p *ImageProcessor) convertGIFToMP4(inputPath, outputPath string) error {
	cmd := exec.Command("ffmpeg", "-i", inputPath, "-c:v", "libx264", "-pix_fmt", "yuv420p", "-y", outputPath)
	return cmd.Run()
}

// getVideoInfo 获取视频信息（简化版）
func (p *ImageProcessor) getVideoInfo(videoPath string) (*models.MediaVariant, error) {
	// 获取文件大小
	fileInfo, err := os.Stat(videoPath)
	if err != nil {
		return nil, err
	}

	return &models.MediaVariant{
		URL:  videoPath,
		Size: fileInfo.Size(),
		// Width和Height需要通过ffprobe获取，这里简化处理
	}, nil
}
