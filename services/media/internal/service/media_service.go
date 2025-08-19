package service

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/flick/backend/services/media/internal/repository"
	"github.com/flick/backend/services/media/proto"
	"github.com/google/uuid"
)

// mediaService 媒体服务实现
type mediaService struct {
	mediaRepo repository.MediaRepository
}

// NewMediaService 创建媒体服务实例
func NewMediaService(mediaRepo repository.MediaRepository) MediaService {
	return &mediaService{
		mediaRepo: mediaRepo,
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
func (s *mediaService) UploadFile(ctx context.Context, req *proto.UploadFileRequest) (*proto.UploadFileResponse, error) {
	// 生成文件ID
	fileID := uuid.New().String()

	// 确定内容类型
	contentType := getContentTypeFromFilename(req.Filename)

	// 保存文件到存储
	url, err := s.mediaRepo.SaveFileToStorage(ctx, fileID, req.FileData, contentType, req.Type, req.UserId)
	if err != nil {
		return &proto.UploadFileResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to save file to storage: " + err.Error(),
			},
		}, err
	}

	// 创建文件记录
	file := &proto.MediaFile{
		Id:        fileID,
		UserId:    req.UserId,
		Filename:  req.Filename,
		Url:       url,
		Type:      req.Type,
		Size:      int64(len(req.FileData)),
		AltText:   req.AltText,
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

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

	return &proto.UploadFileResponse{
		File: file,
	}, nil
}

// GetFile 获取文件信息
func (s *mediaService) GetFile(ctx context.Context, req *proto.GetFileRequest) (*proto.GetFileResponse, error) {
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
func (s *mediaService) DeleteFile(ctx context.Context, req *proto.DeleteFileRequest) (*proto.DeleteFileResponse, error) {
	// 首先获取文件信息
	file, err := s.mediaRepo.GetFileByID(ctx, req.FileId)
	if err != nil {
		return &proto.DeleteFileResponse{
			Success: false,
			Error: &proto.Error{
				Code:    404,
				Message: "File not found: " + err.Error(),
			},
		}, err
	}

	// 从存储中删除文件
	err = s.mediaRepo.DeleteFileFromStorage(ctx, file.Url)
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
	err = s.mediaRepo.DeleteFile(ctx, req.FileId)
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
func (s *mediaService) ListFiles(ctx context.Context, req *proto.ListFilesRequest) (*proto.ListFilesResponse, error) {
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
