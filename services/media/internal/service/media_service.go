package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/flick/backend/pkg/models"
	"github.com/flick/backend/services/media/internal/processor"
	"github.com/flick/backend/services/media/internal/repository"
	"github.com/flick/backend/services/media/internal/storage"
	"github.com/flick/backend/services/media/internal/validator"
	"github.com/flick/backend/services/media/proto"
	userProto "github.com/flick/backend/services/user/proto"
	"github.com/google/uuid"
)

// ProcessingJob 处理任务
type ProcessingJob struct {
	ID       string
	FileID   string
	Status   string // processing, completed, failed
	Progress string // "50%"
	Error    error
	Result   *proto.MediaFile
}

// mediaService 媒体服务实现
type MediaService struct {
	mediaRepo *repository.MediaRepository
	validator      *validator.MediaValidator
	processor      *processor.ProcessorManager
	storage        storage.MediaStorage
	tempDir        string
	processingJobs map[string]*ProcessingJob
	jobsMutex      sync.RWMutex
	userClient     userProto.UserServiceClient
}

// NewMediaService 创建媒体服务实例
func NewMediaService(mediaRepo *repository.MediaRepository, storage storage.MediaStorage, tempDir string, userClient userProto.UserServiceClient) *MediaService {
	if tempDir == "" {
		tempDir = "/tmp/media-processing"
	}

	// 确保临时目录存在
	os.MkdirAll(tempDir, 0755)

	return &MediaService{
		mediaRepo:      mediaRepo,
		validator:      validator.NewMediaValidator(),
		processor:      processor.NewProcessorManager(tempDir),
		storage:        storage,
		tempDir:        tempDir,
		userClient:     userClient,
		processingJobs: make(map[string]*ProcessingJob),
	}
}

// getContentTypeFromFilename determines content type from filename extension
func getContentTypeFromFilename(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
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

// UploadFile 上传文件
func (s *MediaService) UploadFile(ctx context.Context, req *proto.UploadFileRequest) (*proto.UploadFileResponse, error) {
	// 生成文件ID
	fileID := uuid.New().String()

	// 确定内容类型 - 优先使用请求中的ContentType
	contentType := req.ContentType
	if contentType == "" {
		contentType = getContentTypeFromFilename(req.Filename)
	}

	// 确定媒体类别
	var category storage.MediaCategory
	switch req.FileType {
	case "avatar":
		category = storage.CategoryAvatar
	case "banner":
		category = storage.CategoryBanner
	case "post_media":
		category = storage.CategoryPost
	default:
		category = storage.CategoryPost // 默认为帖子媒体
	}

	// 验证文件
	fileSize := int64(len(req.Content))
	if err := s.validator.ValidateUpload(category, req.Filename, fileSize, contentType); err != nil {
		fmt.Printf("[MEDIA SERVICE] Validation failed: %v\n", err)
		return &proto.UploadFileResponse{
			Error: &proto.Error{
				Code:    400,
				Message: err.Error(),
			},
		}, err
	}

	fmt.Printf("[MEDIA SERVICE] File validation passed: %s, size: %d bytes, type: %s\n", req.Filename, fileSize, contentType)

	// 保存文件到存储
	var url string
	var err error

	// 如果是头像或横幅上传，使用版本化存储
	if req.FileType == "avatar" {
		url, err = s.saveVersionedAvatar(ctx, fileID, req.Content, contentType, req.UserId)
	} else if req.FileType == "banner" {
		url, err = s.saveVersionedBanner(ctx, fileID, req.Content, contentType, req.UserId)
	} else {
		url, err = s.mediaRepo.SaveFileToStorage(ctx, fileID, req.Content, contentType, category.String(), req.UserId)
	}

	if err != nil {
		return &proto.UploadFileResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to save file to storage: " + err.Error(),
			},
		}, err
	}

	// 从URL中提取实际的文件ID
	// URL格式: http://host/bucket/category/userID/filename
	// 对于版本化头像，文件名格式: avatar_v1.jpg, avatar_v2.jpg 等
	// 对于普通文件，文件名格式: category_uuid.ext (如 avatar_74bfbdb7-4ff1-430e-82af-29269accbf47.jpg)
	parts := strings.Split(url, "/")
	var actualFileID string
	if len(parts) >= 5 {
		filename := parts[len(parts)-1] // 获取文件名

		// 检查是否为版本化头像或横幅文件
		if req.FileType == "avatar" && strings.HasPrefix(filename, "avatar_v") {
			// 版本化头像使用原始生成的UUID作为文件ID
			actualFileID = fileID
			fmt.Printf("[MEDIA SERVICE] Versioned avatar detected: %s, using original UUID: %s\n", filename, actualFileID)
		} else if req.FileType == "banner" && strings.HasPrefix(filename, "banner_v") {
			// 版本化横幅使用原始生成的UUID作为文件ID
			actualFileID = fileID
			fmt.Printf("[MEDIA SERVICE] Versioned banner detected: %s, using original UUID: %s\n", filename, actualFileID)
		} else if strings.Contains(filename, "_") && strings.Contains(filename, ".") {
			// 普通文件从文件名中提取UUID部分 (category_uuid.ext -> uuid)
			nameParts := strings.Split(filename, "_")
			if len(nameParts) >= 2 {
				uuidWithExt := nameParts[1]
				actualFileID = strings.Split(uuidWithExt, ".")[0] // 移除扩展名
				fmt.Printf("[MEDIA SERVICE] Extracted file ID from filename: %s -> %s\n", filename, actualFileID)
			}
		}
	}

	// 如果无法从URL提取，使用原始生成的UUID
	if actualFileID == "" {
		actualFileID = fileID
		fmt.Printf("[MEDIA SERVICE] Could not extract file ID from URL, using original: %s\n", actualFileID)
	}

	// 创建文件记录
	file := &proto.MediaFile{
		Id:        actualFileID,
		UserId:    req.UserId,
		Filename:  req.Filename,
		Url:       url,
		Type:      req.FileType,
		MimeType:  contentType, // 设置MIME类型
		Size:      int64(len(req.Content)),
		AltText:   req.AltText,
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	// Debug logging for URL generation
	fmt.Printf("[MEDIA SERVICE] Generated file URL: %s\n", url)
	fmt.Printf("[MEDIA SERVICE] File type: %s, User ID: %s\n", req.FileType, req.UserId)

	err = s.mediaRepo.CreateFile(ctx, file)
	if err != nil {
		// 如果创建记录失败，尝试删除已保存的文件
		s.mediaRepo.DeleteFileFromStorage(ctx, url)

		return &proto.UploadFileResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to create file record: " + err.Error(),
			},
		}, err
	}

	// 自动触发媒体处理生成多版本（图片和视频）
	// 跳过版本化头像和横幅的自动处理，因为它们已经是处理后的版本
	if strings.HasPrefix(contentType, "image/") && req.FileType != "avatar" && req.FileType != "banner" {
		fmt.Printf("[MEDIA SERVICE] Triggering automatic media processing for image: %s\n", actualFileID)
		go func() {
			processReq := &proto.ProcessMediaRequest{
				FileId:   actualFileID,
				Variants: []string{"thumbnail", "small", "medium", "large"}, // 生成标准版本
			}
			_, processErr := s.ProcessMedia(context.Background(), processReq)
			if processErr != nil {
				fmt.Printf("[MEDIA SERVICE] Auto-processing failed for %s: %v\n", actualFileID, processErr)
			} else {
				fmt.Printf("[MEDIA SERVICE] Auto-processing started for %s\n", actualFileID)
			}
		}()
	} else if req.FileType == "avatar" {
		fmt.Printf("[MEDIA SERVICE] Skipping auto-processing for versioned avatar: %s\n", actualFileID)
	} else if req.FileType == "banner" {
		fmt.Printf("[MEDIA SERVICE] Skipping auto-processing for versioned banner: %s\n", actualFileID)
	} else if strings.HasPrefix(contentType, "video/") {
		fmt.Printf("[MEDIA SERVICE] Triggering automatic media processing for video: %s\n", actualFileID)
		go func() {
			processReq := &proto.ProcessMediaRequest{
				FileId:   actualFileID,
				Variants: []string{"low_res", "mid_res", "high_res"}, // 生成视频版本和预览图
			}
			_, processErr := s.ProcessMedia(context.Background(), processReq)
			if processErr != nil {
				fmt.Printf("[MEDIA SERVICE] Auto-processing failed for video %s: %v\n", actualFileID, processErr)
			} else {
				fmt.Printf("[MEDIA SERVICE] Auto-processing started for video %s\n", actualFileID)
			}
		}()
	}

	return &proto.UploadFileResponse{
		FileId:      actualFileID,
		FileUrl:     url,
		Filename:    req.Filename,
		ContentType: contentType,
		Size:        int64(len(req.Content)),
	}, nil
}

// GetFile 获取文件信息
func (s *MediaService) GetFile(ctx context.Context, req *proto.GetFileRequest) (*proto.GetFileResponse, error) {
	file, err := s.mediaRepo.GetFileByID(ctx, req.FileId)
	if err != nil {
		return &proto.GetFileResponse{
			Error: &proto.Error{
				Code:    404,
				Message: "File not found: " + err.Error(),
			},
		}, err
	}

	return &proto.GetFileResponse{
		File: file,
	}, nil
}

// DeleteFile 删除文件
func (s *MediaService) DeleteFile(ctx context.Context, req *proto.DeleteFileRequest) (*proto.DeleteFileResponse, error) {
	// 首先获取文件信息 - 通过URL查找文件
	file, err := s.mediaRepo.GetFileByURL(ctx, req.FileUrl)
	if err != nil {
		return &proto.DeleteFileResponse{
			Success: false,
			Error: &proto.Error{
				Code:    404,
				Message: "File not found: " + err.Error(),
			},
		}, err
	}

	// 验证用户权限 - 只有文件所有者可以删除
	if file.UserId != req.UserId {
		return &proto.DeleteFileResponse{
			Success: false,
			Error: &proto.Error{
				Code:    403,
				Message: "Permission denied: you can only delete your own files",
			},
		}, fmt.Errorf("permission denied")
	}

	// 从存储中删除文件
	err = s.mediaRepo.DeleteFileFromStorage(ctx, req.FileUrl)
	if err != nil {
		return &proto.DeleteFileResponse{
			Success: false,
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to delete file from storage: " + err.Error(),
			},
		}, err
	}

	// 删除文件记录
	err = s.mediaRepo.DeleteFile(ctx, file.Id)
	if err != nil {
		return &proto.DeleteFileResponse{
			Success: false,
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to delete file record: " + err.Error(),
			},
		}, err
	}

	return &proto.DeleteFileResponse{
		Success: true,
	}, nil
}

// ListFiles 获取文件列表
func (s *MediaService) ListFiles(ctx context.Context, req *proto.ListFilesRequest) (*proto.ListFilesResponse, error) {
	files, total, err := s.mediaRepo.ListFiles(ctx, req.UserId, req.Page, req.PageSize)
	if err != nil {
		return &proto.ListFilesResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to list files: " + err.Error(),
			},
		}, err
	}

	return &proto.ListFilesResponse{
		Files: files,
		Total: total,
	}, nil
}

// ProcessMedia 处理媒体文件生成多版本
func (s *MediaService) ProcessMedia(ctx context.Context, req *proto.ProcessMediaRequest) (*proto.ProcessMediaResponse, error) {
	// 获取文件信息
	file, err := s.mediaRepo.GetFileByID(ctx, req.FileId)
	if err != nil {
		return &proto.ProcessMediaResponse{
			Error: &proto.Error{
				Code:    404,
				Message: "File not found: " + err.Error(),
			},
		}, err
	}

	// 生成任务ID
	jobID := uuid.New().String()

	// 创建处理任务
	job := &ProcessingJob{
		ID:       jobID,
		FileID:   req.FileId,
		Status:   "processing",
		Progress: "0%",
	}

	// 存储任务
	s.jobsMutex.Lock()
	s.processingJobs[jobID] = job
	s.jobsMutex.Unlock()

	// 异步处理
	go s.processMediaAsync(ctx, job, file, req)

	return &proto.ProcessMediaResponse{
		JobId:  jobID,
		Status: "processing",
	}, nil
}

// GetProcessStatus 获取媒体处理状态
func (s *MediaService) GetProcessStatus(ctx context.Context, req *proto.GetProcessStatusRequest) (*proto.GetProcessStatusResponse, error) {
	s.jobsMutex.RLock()
	job, exists := s.processingJobs[req.JobId]
	s.jobsMutex.RUnlock()

	if !exists {
		return &proto.GetProcessStatusResponse{
			Error: &proto.Error{
				Code:    404,
				Message: "Job not found",
			},
		}, nil
	}

	response := &proto.GetProcessStatusResponse{
		Status:   job.Status,
		Progress: job.Progress,
	}

	if job.Result != nil {
		response.File = job.Result
	}

	if job.Error != nil {
		response.Error = &proto.Error{
			Code:    500,
			Message: job.Error.Error(),
		}
	}

	return response, nil
}

// processMediaAsync 异步处理媒体文件
func (s *MediaService) processMediaAsync(ctx context.Context, job *ProcessingJob, file *proto.MediaFile, req *proto.ProcessMediaRequest) {
	defer func() {
		if r := recover(); r != nil {
			job.Status = "failed"
			job.Error = fmt.Errorf("processing panic: %v", r)
			fmt.Printf("[MEDIA PROCESSOR] Panic during processing: %v\n", r)
		}
	}()

	fmt.Printf("[MEDIA PROCESSOR] Starting processing for file %s\n", file.Id)

	// 更新进度
	job.Progress = "10%"

	// 下载文件到临时目录
	tempFilePath := filepath.Join(s.tempDir, file.Id+"_"+file.Filename)
	fmt.Printf("[MEDIA PROCESSOR] About to download file to: %s\n", tempFilePath)
	err := s.downloadFileToTemp(ctx, file.Url, tempFilePath)
	if err != nil {
		job.Status = "failed"
		job.Error = fmt.Errorf("failed to download file: %w", err)
		fmt.Printf("[MEDIA PROCESSOR] Download failed: %v\n", err)
		return
	}
	defer os.Remove(tempFilePath) // 清理临时文件
	fmt.Printf("[MEDIA PROCESSOR] Download completed, continuing processing\n")

	job.Progress = "30%"

	// 获取媒体类型
	fmt.Printf("[MEDIA PROCESSOR] Getting media type for filename: %s\n", file.Filename)
	mediaType := s.processor.GetMediaType(file.Filename)
	fmt.Printf("[MEDIA PROCESSOR] Detected media type: %s\n", mediaType)

	if req.Variants == nil || len(req.Variants) == 0 {
		req.Variants = s.processor.GetDefaultVariants(mediaType)
		fmt.Printf("[MEDIA PROCESSOR] Using default variants: %v\n", req.Variants)
	}

	// 验证媒体文件
	fmt.Printf("[MEDIA PROCESSOR] Validating media file: %s, size: %d, type: %s\n", file.Filename, file.Size, mediaType)
	err = s.processor.ValidateMediaFile(file.Filename, file.Size, mediaType)
	if err != nil {
		job.Status = "failed"
		job.Error = fmt.Errorf("media validation failed: %w", err)
		fmt.Printf("[MEDIA PROCESSOR] Validation failed: %v\n", err)
		return
	}
	fmt.Printf("[MEDIA PROCESSOR] Media file validation passed\n")

	job.Progress = "50%"

	// 处理媒体文件
	fmt.Printf("[MEDIA PROCESSOR] Starting ProcessMedia for file: %s, mediaType: %s, variants: %v\n", tempFilePath, mediaType, req.Variants)
	variants, err := s.processor.ProcessMedia(tempFilePath, mediaType, req.Variants)
	if err != nil {
		job.Status = "failed"
		job.Error = fmt.Errorf("media processing failed: %w", err)
		fmt.Printf("[MEDIA PROCESSOR] ProcessMedia failed: %v\n", err)
		return
	}

	// 修复 original variant 的 URL 为正确的 MinIO URL
	if variants.Original != nil {
		variants.Original.URL = file.Url
	}

	fmt.Printf("[MEDIA PROCESSOR] ProcessMedia completed successfully\n")

	job.Progress = "70%"

	// 上传处理后的文件到存储
	err = s.uploadVariantsToStorage(ctx, file, variants)
	if err != nil {
		job.Status = "failed"
		job.Error = fmt.Errorf("failed to upload variants: %w", err)
		return
	}

	job.Progress = "90%"

	// 更新数据库记录
	updatedFile := *file
	updatedFile.Status = "ready"
	updatedFile.MimeType = mediaType
	updatedFile.ProcessedAt = time.Now().Format(time.RFC3339)

	// 将models.MediaVariants转换为数据库存储格式
	// 需要转换为proto格式以匹配MediaFile结构
	updatedFile.Variants = s.convertVariantsToProto(variants)

	err = s.mediaRepo.UpdateFile(ctx, &updatedFile)
	if err != nil {
		job.Status = "failed"
		job.Error = fmt.Errorf("failed to update file record: %w", err)
		return
	}

	// 完成处理
	job.Status = "completed"
	job.Progress = "100%"
	job.Result = &updatedFile

	fmt.Printf("[MEDIA PROCESSOR] Successfully processed file %s\n", file.Id)
}

// downloadFileToTemp 下载文件到临时目录
func (s *MediaService) downloadFileToTemp(ctx context.Context, url, tempPath string) error {
	// 确保临时目录存在
	if err := os.MkdirAll(filepath.Dir(tempPath), 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	// 从URL中提取bucket和object路径
	// URL格式: http://127.0.0.1:9000/social-media/posts/user-id/file-id/filename
	parts := strings.Split(url, "/")
	if len(parts) < 4 {
		return fmt.Errorf("invalid MinIO URL format: %s", url)
	}

	// 提取bucket和object路径
	bucketName := parts[3]                     // social-media
	objectPath := strings.Join(parts[4:], "/") // posts/user-id/file-id/filename

	fmt.Printf("[MEDIA PROCESSOR] Downloading from MinIO: bucket=%s, object=%s\n", bucketName, objectPath)

	// 使用storage接口下载文件
	reader, err := s.storage.GetFile(ctx, bucketName, objectPath)
	if err != nil {
		return fmt.Errorf("failed to get file from storage: %w", err)
	}
	defer reader.Close()

	// 创建临时文件
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	// 复制文件内容
	_, err = io.Copy(tempFile, reader)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	fmt.Printf("[MEDIA PROCESSOR] File downloaded to temp path: %s\n", tempPath)
	return nil
}

// uploadVariantsToStorage 上传处理后的文件版本到存储
func (s *MediaService) uploadVariantsToStorage(ctx context.Context, originalFile *proto.MediaFile, variants *models.MediaVariants) error {
	// 从原文件URL中提取路径信息
	parts := strings.Split(originalFile.Url, "/")
	if len(parts) < 6 {
		return fmt.Errorf("invalid original file URL format: %s", originalFile.Url)
	}

	bucketName := parts[3] // social-media
	category := parts[4]   // avatars/banners/posts
	userID := parts[5]     // user-id
	filename := parts[6]   // complete filename with extension

	// 从文件名中提取fileID
	// 文件名格式: avatar_fileID.ext 或 banner_fileID.ext 或 img_fileID.ext
	var fileID string
	if strings.HasPrefix(filename, "avatar_") {
		fileID = strings.TrimPrefix(filename, "avatar_")
	} else if strings.HasPrefix(filename, "banner_") {
		fileID = strings.TrimPrefix(filename, "banner_")
	} else if strings.HasPrefix(filename, "img_") {
		fileID = strings.TrimPrefix(filename, "img_")
	} else {
		fileID = filename
	}
	// 移除扩展名
	fileID = strings.TrimSuffix(fileID, filepath.Ext(fileID))

	originalExt := filepath.Ext(originalFile.Filename)

	fmt.Printf("[MEDIA PROCESSOR] Uploading variants for file %s (category: %s)\n", fileID, category)

	// 上传各个版本 - 包括图片和视频variants
	variantTypes := []struct {
		variant  *models.MediaVariant
		suffix   string
		mimeType string
	}{
		// 图片variants
		{variants.Thumbnail, "thumbnail", "image/jpeg"},
		{variants.Small, "small", "image/jpeg"},
		{variants.Medium, "medium", "image/jpeg"},
		{variants.Large, "large", "image/jpeg"},
		// 视频variants
		{variants.Preview, "preview", "image/jpeg"},
		{variants.LowRes, "low_res", "video/mp4"},
		{variants.MidRes, "mid_res", "video/mp4"},
		{variants.HighRes, "high_res", "video/mp4"},
	}

	for _, vt := range variantTypes {
		if vt.variant == nil {
			continue
		}

		// 构建MinIO对象路径 - 使用原始文件的category而不是硬编码posts
		// 修复variant URL路径重复问题 - 不在文件名中重复fileID
		var variantFilename string
		if vt.suffix == "preview" {
			variantFilename = fmt.Sprintf("%s.jpg", vt.suffix) // 预览图使用.jpg扩展名
		} else if strings.Contains(vt.mimeType, "video") {
			variantFilename = fmt.Sprintf("%s.mp4", vt.suffix) // 视频variants使用.mp4扩展名
		} else {
			variantFilename = fmt.Sprintf("%s%s", vt.suffix, originalExt) // 图片variants使用原扩展名
		}
		objectPath := fmt.Sprintf("%s/%s/%s/%s", category, userID, fileID, variantFilename)

		// 打开本地处理后的文件
		file, err := os.Open(vt.variant.URL) // 这里URL实际是本地临时文件路径
		if err != nil {
			fmt.Printf("[MEDIA PROCESSOR] Warning: failed to open variant file %s: %v\n", vt.variant.URL, err)
			continue
		}

		// 获取文件信息
		fileInfo, err := file.Stat()
		if err != nil {
			file.Close()
			continue
		}

		// 上传到MinIO - 使用正确的MIME类型
		url, err := s.storage.UploadFileWithPath(ctx, bucketName, objectPath, file, fileInfo.Size(), vt.mimeType)
		file.Close()

		if err != nil {
			fmt.Printf("[MEDIA PROCESSOR] Warning: failed to upload variant %s: %v\n", vt.suffix, err)
			continue
		}

		// 保存临时文件路径用于清理
		tempPath := vt.variant.URL

		// 更新variant的URL为MinIO URL
		vt.variant.URL = url
		vt.variant.Size = fileInfo.Size()

		fmt.Printf("[MEDIA PROCESSOR] Uploaded %s variant: %s\n", vt.suffix, url)

		// 清理临时文件
		os.Remove(tempPath) // 删除本地临时文件
	}

	return nil
}

// convertVariantsToProto 转换variants到proto格式
func (s *MediaService) convertVariantsToProto(variants *models.MediaVariants) *proto.MediaVariants {
	result := &proto.MediaVariants{}

	if variants.Thumbnail != nil {
		result.Thumbnail = &proto.MediaVariant{
			Url:    variants.Thumbnail.URL,
			Width:  variants.Thumbnail.Width,
			Height: variants.Thumbnail.Height,
			Size:   variants.Thumbnail.Size,
		}
	}

	if variants.Small != nil {
		result.Small = &proto.MediaVariant{
			Url:    variants.Small.URL,
			Width:  variants.Small.Width,
			Height: variants.Small.Height,
			Size:   variants.Small.Size,
		}
	}

	if variants.Medium != nil {
		result.Medium = &proto.MediaVariant{
			Url:    variants.Medium.URL,
			Width:  variants.Medium.Width,
			Height: variants.Medium.Height,
			Size:   variants.Medium.Size,
		}
	}

	if variants.Large != nil {
		result.Large = &proto.MediaVariant{
			Url:    variants.Large.URL,
			Width:  variants.Large.Width,
			Height: variants.Large.Height,
			Size:   variants.Large.Size,
		}
	}

	if variants.Original != nil {
		result.Original = &proto.MediaVariant{
			Url:    variants.Original.URL,
			Width:  variants.Original.Width,
			Height: variants.Original.Height,
			Size:   variants.Original.Size,
		}
	}

	// 视频特有版本
	if variants.Preview != nil {
		result.Preview = &proto.MediaVariant{
			Url:    variants.Preview.URL,
			Width:  variants.Preview.Width,
			Height: variants.Preview.Height,
			Size:   variants.Preview.Size,
		}
	}

	if variants.LowRes != nil {
		result.LowRes = &proto.MediaVariant{
			Url:    variants.LowRes.URL,
			Width:  variants.LowRes.Width,
			Height: variants.LowRes.Height,
			Size:   variants.LowRes.Size,
		}
	}

	if variants.MidRes != nil {
		result.MidRes = &proto.MediaVariant{
			Url:    variants.MidRes.URL,
			Width:  variants.MidRes.Width,
			Height: variants.MidRes.Height,
			Size:   variants.MidRes.Size,
		}
	}

	if variants.HighRes != nil {
		result.HighRes = &proto.MediaVariant{
			Url:    variants.HighRes.URL,
			Width:  variants.HighRes.Width,
			Height: variants.HighRes.Height,
			Size:   variants.HighRes.Size,
		}
	}

	return result
}

// saveVersionedAvatar 保存版本化头像到MinIO存储
func (s *MediaService) saveVersionedAvatar(ctx context.Context, fileID string, content []byte, contentType, userID string) (string, error) {
	fmt.Printf("[MEDIA SERVICE] Starting saveVersionedAvatar for user: %s\n", userID)

	// 获取用户当前头像版本号
	currentVersion, err := s.getUserAvatarVersion(ctx, userID)
	if err != nil {
		fmt.Printf("[MEDIA SERVICE] Failed to get user avatar version: %v\n", err)
		currentVersion = 0 // 默认从版本0开始
	}

	fmt.Printf("[MEDIA SERVICE] Retrieved current version: %d for user: %s\n", currentVersion, userID)

	// 增加版本号
	newVersion := currentVersion + 1
	fmt.Printf("[MEDIA SERVICE] New version will be: %d for user: %s\n", newVersion, userID)

	// 生成版本化文件名
	fileExt := filepath.Ext(fileID)
	if fileExt == "" {
		// 从content type推断扩展名
		switch contentType {
		case "image/jpeg":
			fileExt = ".jpg"
		case "image/png":
			fileExt = ".png"
		case "image/webp":
			fileExt = ".webp"
		default:
			fileExt = ".jpg"
		}
	}

	versionedFileName := fmt.Sprintf("avatar_v%d%s", newVersion, fileExt)
	objectPath := fmt.Sprintf("avatars/%s/%s", userID, versionedFileName)

	fmt.Printf("[MEDIA SERVICE] Saving versioned avatar: %s (version %d)\n", objectPath, newVersion)

	// 上传到MinIO
	reader := bytes.NewReader(content)
	url, err := s.storage.UploadFileWithPath(ctx, "social-media", objectPath, reader, int64(len(content)), contentType)
	if err != nil {
		return "", fmt.Errorf("failed to upload versioned avatar: %w", err)
	}

	// 更新用户头像版本号（这里需要调用User Service）
	err = s.updateUserAvatarVersion(ctx, userID, newVersion, url)
	if err != nil {
		fmt.Printf("[MEDIA SERVICE] Warning: Failed to update user avatar version: %v\n", err)
		// 不返回错误，因为文件已经上传成功
	}

	// 异步清理旧头像文件
	go s.scheduleOldAvatarCleanup(ctx, userID, newVersion)

	return url, nil
}

// getUserAvatarVersion 获取用户当前头像版本号
func (s *MediaService) getUserAvatarVersion(ctx context.Context, userID string) (int, error) {
	fmt.Printf("[MEDIA SERVICE] Getting avatar version for user: %s\n", userID)

	if s.userClient == nil {
		fmt.Printf("[MEDIA SERVICE] User client not available, defaulting to version 0\n")
		return 0, nil
	}

	resp, err := s.userClient.GetUser(ctx, &userProto.GetUserRequest{
		UserId: userID,
	})
	if err != nil {
		fmt.Printf("[MEDIA SERVICE] Failed to get user %s: %v\n", userID, err)
		return 0, err
	}

	if resp.Error != nil {
		fmt.Printf("[MEDIA SERVICE] User service error: %s\n", resp.Error.Message)
		return 0, fmt.Errorf("user service error: %s", resp.Error.Message)
	}

	currentVersion := int(resp.User.AvatarVersion)
	fmt.Printf("[MEDIA SERVICE] Current avatar version for user %s: %d\n", userID, currentVersion)
	return currentVersion, nil
}

// updateUserAvatarVersion 更新用户头像版本号和URL
func (s *MediaService) updateUserAvatarVersion(ctx context.Context, userID string, version int, avatarURL string) error {
	if s.userClient == nil {
		fmt.Printf("[MEDIA SERVICE] User client not available, cannot update avatar version\n")
		return fmt.Errorf("user client not available")
	}

	resp, err := s.userClient.UpdateUserAvatar(ctx, &userProto.UpdateUserAvatarRequest{
		UserId:        userID,
		AvatarUrl:     avatarURL,
		AvatarVersion: int32(version),
	})
	if err != nil {
		fmt.Printf("[MEDIA SERVICE] Failed to update user avatar %s: %v\n", userID, err)
		return err
	}

	if resp.Error != nil {
		fmt.Printf("[MEDIA SERVICE] User service error updating avatar: %s\n", resp.Error.Message)
		return fmt.Errorf("user service error: %s", resp.Error.Message)
	}

	fmt.Printf("[MEDIA SERVICE] Successfully updated user %s avatar version to %d\n", userID, version)
	return nil
}

// scheduleOldAvatarCleanup 安排旧头像文件清理
func (s *MediaService) scheduleOldAvatarCleanup(ctx context.Context, userID string, currentVersion int) {
	fmt.Printf("[MEDIA SERVICE] Scheduling cleanup for user %s, current version: %d\n", userID, currentVersion)

	// 创建一个新的context，避免原context被取消
	cleanupCtx := context.Background()

	// 延迟30秒后开始清理，避免客户端仍在使用旧URL时出现404（缩短测试时间）
	time.Sleep(30 * time.Second)

	fmt.Printf("[MEDIA SERVICE] Starting cleanup of old avatars for user %s, keeping version %d\n", userID, currentVersion)

	// 保留最近2个版本，删除更老的版本（降低阈值便于测试）
	keepVersions := 2
	fmt.Printf("[MEDIA SERVICE] Cleanup policy: keep %d versions, current version: %d\n", keepVersions, currentVersion)

	if currentVersion > keepVersions {
		fmt.Printf("[MEDIA SERVICE] Will delete versions 1 to %d\n", currentVersion-keepVersions)
		for version := 1; version <= currentVersion-keepVersions; version++ {
			// 构造完整的URL格式，与存储时使用的格式一致
			oldObjectPath := fmt.Sprintf("avatars/%s/avatar_v%d.jpg", userID, version)
			oldURL := fmt.Sprintf("http://127.0.0.1:9000/social-media/%s", oldObjectPath)

			fmt.Printf("[MEDIA SERVICE] Attempting to delete old avatar: %s\n", oldURL)
			err := s.storage.DeleteFile(cleanupCtx, oldURL)
			if err != nil {
				fmt.Printf("[MEDIA SERVICE] Failed to delete old avatar %s: %v\n", oldURL, err)
			} else {
				fmt.Printf("[MEDIA SERVICE] Successfully deleted old avatar: %s\n", oldURL)
			}

			// 也尝试删除其他可能的扩展名
			for _, ext := range []string{".png", ".webp"} {
				altObjectPath := fmt.Sprintf("avatars/%s/avatar_v%d%s", userID, version, ext)
				altURL := fmt.Sprintf("http://127.0.0.1:9000/social-media/%s", altObjectPath)
				fmt.Printf("[MEDIA SERVICE] Attempting to delete alternative format: %s\n", altURL)
				err := s.storage.DeleteFile(cleanupCtx, altURL)
				if err != nil {
					fmt.Printf("[MEDIA SERVICE] Failed to delete alternative format %s: %v\n", altURL, err)
				} else {
					fmt.Printf("[MEDIA SERVICE] Successfully deleted alternative format: %s\n", altURL)
				}
			}
		}
	} else {
		fmt.Printf("[MEDIA SERVICE] No cleanup needed - current version (%d) <= keep versions (%d)\n", currentVersion, keepVersions)
	}

	fmt.Printf("[MEDIA SERVICE] Cleanup completed for user %s\n", userID)
}

// saveVersionedBanner 保存版本化横幅文件
func (s *MediaService) saveVersionedBanner(ctx context.Context, fileID string, content []byte, contentType, userID string) (string, error) {
	fmt.Printf("[MEDIA SERVICE] Saving versioned banner for user: %s\n", userID)

	// 获取用户当前横幅版本号
	currentVersion, err := s.getUserBannerVersion(ctx, userID)
	if err != nil {
		fmt.Printf("[MEDIA SERVICE] Failed to get user banner version: %v\n", err)
		currentVersion = 0 // 默认从版本0开始
	}

	// 递增版本号
	newVersion := currentVersion + 1

	// 确定文件扩展名
	fileExt := ".jpg" // 默认扩展名
	switch contentType {
	case "image/png":
		fileExt = ".png"
	case "image/webp":
		fileExt = ".webp"
	case "image/gif":
		fileExt = ".gif"
	}

	versionedFileName := fmt.Sprintf("banner_v%d%s", newVersion, fileExt)
	objectPath := fmt.Sprintf("banners/%s/%s", userID, versionedFileName)

	fmt.Printf("[MEDIA SERVICE] Saving versioned banner: %s (version %d)\n", objectPath, newVersion)

	// 上传到MinIO
	reader := bytes.NewReader(content)
	url, err := s.storage.UploadFileWithPath(ctx, "social-media", objectPath, reader, int64(len(content)), contentType)
	if err != nil {
		return "", fmt.Errorf("failed to upload versioned banner: %w", err)
	}

	// 更新用户横幅版本号（这里需要调用User Service）
	err = s.updateUserBannerVersion(ctx, userID, newVersion, url)
	if err != nil {
		fmt.Printf("[MEDIA SERVICE] Warning: Failed to update user banner version: %v\n", err)
		// 不返回错误，因为文件已经上传成功
	}

	// 异步清理旧版本横幅文件
	go s.cleanupOldBanners(userID, newVersion)

	return url, nil
}

// getUserBannerVersion 获取用户当前横幅版本号
func (s *MediaService) getUserBannerVersion(ctx context.Context, userID string) (int, error) {
	fmt.Printf("[MEDIA SERVICE] Getting banner version for user: %s\n", userID)

	if s.userClient == nil {
		fmt.Printf("[MEDIA SERVICE] User client not available, defaulting to version 0\n")
		return 0, nil
	}

	resp, err := s.userClient.GetUser(ctx, &userProto.GetUserRequest{
		UserId: userID,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get user: %w", err)
	}

	if resp.Error != nil {
		return 0, fmt.Errorf("user service error: %s", resp.Error.Message)
	}

	currentVersion := int(resp.User.BannerVersion)
	fmt.Printf("[MEDIA SERVICE] Current banner version for user %s: %d\n", userID, currentVersion)
	return currentVersion, nil
}

// updateUserBannerVersion 更新用户横幅版本号和URL
func (s *MediaService) updateUserBannerVersion(ctx context.Context, userID string, version int, bannerURL string) error {
	if s.userClient == nil {
		fmt.Printf("[MEDIA SERVICE] User client not available, cannot update banner version\n")
		return fmt.Errorf("user client not available")
	}

	resp, err := s.userClient.UpdateUserBanner(ctx, &userProto.UpdateUserBannerRequest{
		UserId:        userID,
		BannerUrl:     bannerURL,
		BannerVersion: int32(version),
	})
	if err != nil {
		return fmt.Errorf("failed to update user banner: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("user service error: %s", resp.Error.Message)
	}

	fmt.Printf("[MEDIA SERVICE] Successfully updated user %s banner version to %d\n", userID, version)
	return nil
}

// cleanupOldBanners 清理用户的旧横幅版本文件
func (s *MediaService) cleanupOldBanners(userID string, currentVersion int) {
	// 创建新的context避免原context被取消
	cleanupCtx := context.Background()

	// 延迟30秒后开始清理，避免客户端仍在使用旧URL时出现404
	time.Sleep(30 * time.Second)

	fmt.Printf("[MEDIA SERVICE] Starting cleanup of old banners for user %s, keeping version %d\n", userID, currentVersion)

	// 保留最近2个版本，删除更老的版本
	keepVersions := 2
	if currentVersion > keepVersions {
		// 删除版本1到(currentVersion - keepVersions)
		fmt.Printf("[MEDIA SERVICE] Will delete versions 1 to %d\n", currentVersion-keepVersions)
		for version := 1; version <= currentVersion-keepVersions; version++ {
			// 构造完整的URL格式，与存储时使用的格式一致
			oldObjectPath := fmt.Sprintf("banners/%s/banner_v%d.jpg", userID, version)
			oldURL := fmt.Sprintf("http://127.0.0.1:9000/social-media/%s", oldObjectPath)

			fmt.Printf("[MEDIA SERVICE] Attempting to delete old banner: %s\n", oldURL)
			err := s.storage.DeleteFile(cleanupCtx, oldURL)
			if err != nil {
				fmt.Printf("[MEDIA SERVICE] Failed to delete old banner %s: %v\n", oldURL, err)
			} else {
				fmt.Printf("[MEDIA SERVICE] Successfully deleted old banner: %s\n", oldURL)
			}

			// 也尝试删除其他可能的扩展名
			for _, ext := range []string{".png", ".webp"} {
				altObjectPath := fmt.Sprintf("banners/%s/banner_v%d%s", userID, version, ext)
				altURL := fmt.Sprintf("http://127.0.0.1:9000/social-media/%s", altObjectPath)
				fmt.Printf("[MEDIA SERVICE] Attempting to delete alternative format: %s\n", altURL)
				err := s.storage.DeleteFile(cleanupCtx, altURL)
				if err != nil {
					fmt.Printf("[MEDIA SERVICE] Failed to delete alternative format %s: %v\n", altURL, err)
				} else {
					fmt.Printf("[MEDIA SERVICE] Successfully deleted alternative format: %s\n", altURL)
				}
			}
		}
	} else {
		fmt.Printf("[MEDIA SERVICE] No cleanup needed - current version (%d) <= keep versions (%d)\n", currentVersion, keepVersions)
	}

	fmt.Printf("[MEDIA SERVICE] Banner cleanup completed for user %s\n", userID)
}

