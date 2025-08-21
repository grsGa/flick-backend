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

// VideoProcessor 视频处理器
type VideoProcessor struct {
	tempDir string
}

// NewVideoProcessor 创建视频处理器
func NewVideoProcessor(tempDir string) *VideoProcessor {
	return &VideoProcessor{
		tempDir: tempDir,
	}
}

// ProcessVideo 处理视频生成多版本和预览图
func (p *VideoProcessor) ProcessVideo(inputPath string, variants []string) (*models.MediaVariants, error) {
	result := &models.MediaVariants{}

	// 获取视频信息
	videoInfo, err := p.getVideoInfo(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get video info: %w", err)
	}

	// 生成预览图（第1秒的帧）
	previewPath := p.getVariantPath(inputPath, "preview", ".jpg")
	err = p.extractVideoFrame(inputPath, previewPath, "00:00:01")
	if err != nil {
		return nil, fmt.Errorf("failed to extract video frame: %w", err)
	}

	previewInfo, err := p.getImageInfo(previewPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get preview info: %w", err)
	}

	result.Preview = &models.MediaVariant{
		URL:    previewPath,
		Width:  previewInfo.Width,
		Height: previewInfo.Height,
		Size:   previewInfo.Size,
	}

	// 生成不同分辨率的视频版本
	for _, variant := range variants {
		var targetHeight int
		var bitrate string

		switch variant {
		case "low_res":
			targetHeight = 240
			bitrate = "500k"
		case "mid_res":
			targetHeight = 480
			bitrate = "1000k"
		case "high_res":
			targetHeight = 720
			bitrate = "2000k"
		default:
			continue
		}

		// 如果原视频分辨率小于目标分辨率，跳过
		if videoInfo.Height <= int32(targetHeight) {
			continue
		}

		outputPath := p.getVariantPath(inputPath, variant, ".mp4")
		err := p.transcodeVideo(inputPath, outputPath, targetHeight, bitrate)
		if err != nil {
			return nil, fmt.Errorf("failed to transcode video for %s: %w", variant, err)
		}

		// 获取转码后视频信息
		variantInfo, err := p.getVideoInfo(outputPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get variant video info: %w", err)
		}

		mediaVariant := &models.MediaVariant{
			URL:    outputPath,
			Width:  variantInfo.Width,
			Height: variantInfo.Height,
			Size:   variantInfo.Size,
		}

		switch variant {
		case "low_res":
			result.LowRes = mediaVariant
		case "mid_res":
			result.MidRes = mediaVariant
		case "high_res":
			result.HighRes = mediaVariant
		}
	}

	// 设置原视频信息
	result.Original = &models.MediaVariant{
		URL:    inputPath,
		Width:  videoInfo.Width,
		Height: videoInfo.Height,
		Size:   videoInfo.Size,
	}

	return result, nil
}

// transcodeVideo 转码视频到指定分辨率和码率
func (p *VideoProcessor) transcodeVideo(inputPath, outputPath string, height int, bitrate string) error {
	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 使用FFmpeg转码
	// -vf scale=-2:height 保持宽高比，高度设为指定值
	// -c:v libx264 使用H.264编码
	// -b:v bitrate 设置视频码率
	// -c:a aac 音频使用AAC编码
	// -movflags +faststart 优化web播放
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-vf", fmt.Sprintf("scale=-2:%d", height),
		"-c:v", "libx264",
		"-b:v", bitrate,
		"-c:a", "aac",
		"-movflags", "+faststart",
		"-y", outputPath,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg transcode failed: %w", err)
	}

	return nil
}

// extractVideoFrame 提取视频帧作为预览图
func (p *VideoProcessor) extractVideoFrame(inputPath, outputPath, timestamp string) error {
	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 使用FFmpeg提取帧
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-ss", timestamp,
		"-vframes", "1",
		"-q:v", "2", // 高质量
		"-y", outputPath,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg frame extraction failed: %w", err)
	}

	return nil
}

// getVideoInfo 获取视频信息
func (p *VideoProcessor) getVideoInfo(videoPath string) (*models.MediaVariant, error) {
	// 使用ffprobe获取视频信息
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "csv=p=0",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height,duration",
		videoPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	parts := strings.Split(strings.TrimSpace(string(output)), ",")
	if len(parts) != 3 {
		return nil, fmt.Errorf("unexpected ffprobe output: %s", string(output))
	}

	width, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid width: %s", parts[0])
	}

	height, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid height: %s", parts[1])
	}

	// 获取文件大小
	fileInfo, err := os.Stat(videoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &models.MediaVariant{
		Width:  int32(width),
		Height: int32(height),
		Size:   fileInfo.Size(),
	}, nil
}

// getImageInfo 获取图片信息（用于预览图）
func (p *VideoProcessor) getImageInfo(imagePath string) (*models.MediaVariant, error) {
	// 使用ImageMagick的identify命令获取图片信息
	cmd := exec.Command("identify", "-format", "%w %h %B", imagePath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("imagemagick identify failed: %w", err)
	}

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
func (p *VideoProcessor) getVariantPath(originalPath, variant, ext string) string {
	dir := filepath.Dir(originalPath)
	name := strings.TrimSuffix(filepath.Base(originalPath), filepath.Ext(originalPath))
	return filepath.Join(dir, fmt.Sprintf("%s_%s%s", name, variant, ext))
}

// GenerateHLSPlaylist 生成HLS播放列表（高级功能）
func (p *VideoProcessor) GenerateHLSPlaylist(inputPath, outputDir string) error {
	// 确保输出目录存在
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	playlistPath := filepath.Join(outputDir, "playlist.m3u8")
	segmentPath := filepath.Join(outputDir, "segment_%03d.ts")

	// 生成HLS分片
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-c:v", "libx264",
		"-c:a", "aac",
		"-hls_time", "10", // 每个分片10秒
		"-hls_list_size", "0", // 保留所有分片
		"-hls_segment_filename", segmentPath,
		"-y", playlistPath,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg HLS generation failed: %w", err)
	}

	return nil
}
