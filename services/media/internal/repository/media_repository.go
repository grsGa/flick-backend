package repository

import (
	"context"
	"errors"
	"time"

	"backend/pkg/database"
	"backend/pkg/models"
	"backend/services/media/proto"

	"gorm.io/gorm"
)

// mediaRepository 媒体仓储实现
type mediaRepository struct {
	db *gorm.DB
}

// NewMediaRepository 创建媒体仓储实例
func NewMediaRepository() MediaRepository {
	return &mediaRepository{
		db: database.GetDB(),
	}
}

// CreateFile 创建文件记录
func (r *mediaRepository) CreateFile(ctx context.Context, file *proto.MediaFile) error {
	createdAt, _ := time.Parse(time.RFC3339, file.CreatedAt)
	media := &models.MediaAttachment{
		ID:        file.Id,
		PostID:    "", // 需要与内容服务关联
		URL:       file.Url,
		Type:      file.Type,
		AltText:   &file.AltText,
		CreatedAt: createdAt,
	}

	return r.db.Create(media).Error
}

// GetFileByID 根据ID获取文件
func (r *mediaRepository) GetFileByID(ctx context.Context, id string) (*proto.MediaFile, error) {
	var media models.MediaAttachment
	if err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&media).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("file not found")
		}
		return nil, err
	}

	return &proto.MediaFile{
		Id:        media.ID,
		Url:       media.URL,
		Type:      media.Type,
		AltText:   toString(media.AltText),
		CreatedAt: media.CreatedAt.Format(time.RFC3339),
		UpdatedAt: media.CreatedAt.Format(time.RFC3339),
	}, nil
}

// DeleteFile 删除文件记录
func (r *mediaRepository) DeleteFile(ctx context.Context, id string) error {
	return r.db.Where("id = ?", id).Delete(&models.MediaAttachment{}).Error
}

// ListFiles 列出文件
func (r *mediaRepository) ListFiles(ctx context.Context, userID string, page, pageSize int32) ([]*proto.MediaFile, int32, error) {
	// 注意：由于媒体附件与帖子关联而不是直接与用户关联，
	// 这里需要与内容服务进行交互来获取特定用户的所有媒体文件
	// 为简化实现，这里仅演示查询逻辑

	var mediaAttachments []models.MediaAttachment
	var total int64

	// 查询总数
	if err := r.db.Model(&models.MediaAttachment{}).Where("deleted_at IS NULL").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询列表
	offset := (page - 1) * pageSize
	if err := r.db.Where("deleted_at IS NULL").Offset(int(offset)).Limit(int(pageSize)).Find(&mediaAttachments).Error; err != nil {
		return nil, 0, err
	}

	files := make([]*proto.MediaFile, len(mediaAttachments))
	for i, media := range mediaAttachments {
		files[i] = &proto.MediaFile{
			Id:        media.ID,
			Url:       media.URL,
			Type:      media.Type,
			AltText:   toString(media.AltText),
			CreatedAt: media.CreatedAt.Format(time.RFC3339),
			UpdatedAt: media.CreatedAt.Format(time.RFC3339),
		}
	}

	return files, int32(total), nil
}

// SaveFileToStorage 保存文件到存储(简化实现)
func (r *mediaRepository) SaveFileToStorage(ctx context.Context, fileID string, fileData []byte) (string, error) {
	// 在实际实现中，这里应该将文件保存到对象存储服务(如MinIO、S3等)
	// 并返回文件的URL

	// 简化实现，返回模拟URL
	url := "https://example.com/media/" + fileID
	return url, nil
}

// DeleteFileFromStorage 从存储中删除文件(简化实现)
func (r *mediaRepository) DeleteFileFromStorage(ctx context.Context, url string) error {
	// 在实际实现中，这里应该从对象存储服务中删除文件

	// 简化实现，直接返回nil
	return nil
}

// toString 将*string转换为string
func toString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
