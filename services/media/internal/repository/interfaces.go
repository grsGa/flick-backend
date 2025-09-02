package repository

import (
	"context"

	"github.com/flick/backend/services/media/proto"
)

// MediaRepository 定义媒体仓储接口
type MediaRepository interface {
	// CreateFile 创建文件记录
	CreateFile(ctx context.Context, file *proto.MediaFile) error

	// GetFileByID 根据ID获取文件
	GetFileByID(ctx context.Context, id string) (*proto.MediaFile, error)

	// GetFileByURL 根据URL获取文件
	GetFileByURL(ctx context.Context, url string) (*proto.MediaFile, error)

	// UpdateFile 更新文件记录
	UpdateFile(ctx context.Context, file *proto.MediaFile) error

	// DeleteFile 删除文件记录
	DeleteFile(ctx context.Context, id string) error

	// ListFiles 列出文件
	ListFiles(ctx context.Context, userID string, page, pageSize int32) ([]*proto.MediaFile, int32, error)

	// SaveFileToStorage 保存文件到存储
	SaveFileToStorage(ctx context.Context, fileID string, fileData []byte, contentType, category, userID string) (string, error)

	// DeleteFileFromStorage 从存储中删除文件
	DeleteFileFromStorage(ctx context.Context, url string) error
}
