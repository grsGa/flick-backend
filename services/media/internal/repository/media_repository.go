package repository

import (
	"bytes"
	"context"
	"errors"
	"time"

	"github.com/flick/backend/pkg/database"
	"github.com/flick/backend/pkg/models"
	"github.com/flick/backend/services/media/proto"

	"gorm.io/gorm"
)

// mediaRepository 媒体仓储实现
type mediaRepository struct {
	db          *gorm.DB
	storageRepo *MinIOStorageRepository
}

// NewMediaRepository 创建媒体仓储实例
func NewMediaRepository(storageRepo *MinIOStorageRepository) MediaRepository {
	return &mediaRepository{
		db:          database.GetDB(),
		storageRepo: storageRepo,
	}
}

// CreateFile 创建文件记录
func (r *mediaRepository) CreateFile(ctx context.Context, file *proto.MediaFile) error {
	createdAt, _ := time.Parse(time.RFC3339, file.CreatedAt)
	media := &models.MediaAttachment{
		ID:        file.Id,
		PostID:    nil,           // 对于头像/横幅等独立文件，PostID为nil
		UserID:    file.UserId,   // 添加用户ID字段
		Filename:  file.Filename, // 设置文件名
		URL:       file.Url,
		Type:      file.Type,
		MimeType:  file.MimeType, // 设置MIME类型
		AltText:   &file.AltText,
		CreatedAt: createdAt,
	}

	// 如果有PostID，则设置它
	if file.PostId != "" {
		media.PostID = &file.PostId
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

	// Convert MediaVariants to proto format
	var protoVariants *proto.MediaVariants
	if media.Variants != (models.MediaVariants{}) {
		protoVariants = &proto.MediaVariants{}
		
		if media.Variants.Thumbnail != nil {
			protoVariants.Thumbnail = &proto.MediaVariant{
				Url:    media.Variants.Thumbnail.URL,
				Width:  media.Variants.Thumbnail.Width,
				Height: media.Variants.Thumbnail.Height,
				Size:   media.Variants.Thumbnail.Size,
			}
		}
		
		if media.Variants.Small != nil {
			protoVariants.Small = &proto.MediaVariant{
				Url:    media.Variants.Small.URL,
				Width:  media.Variants.Small.Width,
				Height: media.Variants.Small.Height,
				Size:   media.Variants.Small.Size,
			}
		}
		
		if media.Variants.Medium != nil {
			protoVariants.Medium = &proto.MediaVariant{
				Url:    media.Variants.Medium.URL,
				Width:  media.Variants.Medium.Width,
				Height: media.Variants.Medium.Height,
				Size:   media.Variants.Medium.Size,
			}
		}
		
		if media.Variants.Large != nil {
			protoVariants.Large = &proto.MediaVariant{
				Url:    media.Variants.Large.URL,
				Width:  media.Variants.Large.Width,
				Height: media.Variants.Large.Height,
				Size:   media.Variants.Large.Size,
			}
		}
		
		if media.Variants.Original != nil {
			protoVariants.Original = &proto.MediaVariant{
				Url:    media.Variants.Original.URL,
				Width:  media.Variants.Original.Width,
				Height: media.Variants.Original.Height,
				Size:   media.Variants.Original.Size,
			}
		}
		
		// Video variants
		if media.Variants.Preview != nil {
			protoVariants.Preview = &proto.MediaVariant{
				Url:    media.Variants.Preview.URL,
				Width:  media.Variants.Preview.Width,
				Height: media.Variants.Preview.Height,
				Size:   media.Variants.Preview.Size,
			}
		}
		
		if media.Variants.LowRes != nil {
			protoVariants.LowRes = &proto.MediaVariant{
				Url:    media.Variants.LowRes.URL,
				Width:  media.Variants.LowRes.Width,
				Height: media.Variants.LowRes.Height,
				Size:   media.Variants.LowRes.Size,
			}
		}
		
		if media.Variants.MidRes != nil {
			protoVariants.MidRes = &proto.MediaVariant{
				Url:    media.Variants.MidRes.URL,
				Width:  media.Variants.MidRes.Width,
				Height: media.Variants.MidRes.Height,
				Size:   media.Variants.MidRes.Size,
			}
		}
		
		if media.Variants.HighRes != nil {
			protoVariants.HighRes = &proto.MediaVariant{
				Url:    media.Variants.HighRes.URL,
				Width:  media.Variants.HighRes.Width,
				Height: media.Variants.HighRes.Height,
				Size:   media.Variants.HighRes.Size,
			}
		}
	}

	// Handle optional fields
	var altText string
	if media.AltText != nil {
		altText = *media.AltText
	}

	var postID string
	if media.PostID != nil {
		postID = *media.PostID
	}

	var thumbnailURL string
	if media.ThumbnailURL != nil {
		thumbnailURL = *media.ThumbnailURL
	}

	var processedAt string
	if media.ProcessedAt != nil {
		processedAt = media.ProcessedAt.Format(time.RFC3339)
	}

	return &proto.MediaFile{
		Id:          media.ID,
		UserId:      media.UserID,
		Filename:    media.Filename,
		Url:         media.URL,
		Type:        media.Type,
		MimeType:    media.MimeType,
		Size:        media.Size,
		AltText:     altText,
		CreatedAt:   media.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   media.UpdatedAt.Format(time.RFC3339),
		PostId:      postID,
		Status:      media.Status,
		Width:       media.Width,
		Height:      media.Height,
		Duration:    media.Duration,
		ThumbnailUrl: thumbnailURL,
		Variants:    protoVariants,
		ProcessedAt: processedAt,
	}, nil
}

// GetFileByURL 根据URL获取文件
func (r *mediaRepository) GetFileByURL(ctx context.Context, url string) (*proto.MediaFile, error) {
	var media models.MediaAttachment
	if err := r.db.Where("url = ? AND deleted_at IS NULL", url).First(&media).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("file not found")
		}
		return nil, err
	}

	// Convert MediaVariants to proto format
	var protoVariants *proto.MediaVariants
	if media.Variants != (models.MediaVariants{}) {
		protoVariants = &proto.MediaVariants{}
		
		if media.Variants.Thumbnail != nil {
			protoVariants.Thumbnail = &proto.MediaVariant{
				Url:    media.Variants.Thumbnail.URL,
				Width:  media.Variants.Thumbnail.Width,
				Height: media.Variants.Thumbnail.Height,
				Size:   media.Variants.Thumbnail.Size,
			}
		}
		
		if media.Variants.Small != nil {
			protoVariants.Small = &proto.MediaVariant{
				Url:    media.Variants.Small.URL,
				Width:  media.Variants.Small.Width,
				Height: media.Variants.Small.Height,
				Size:   media.Variants.Small.Size,
			}
		}
		
		if media.Variants.Medium != nil {
			protoVariants.Medium = &proto.MediaVariant{
				Url:    media.Variants.Medium.URL,
				Width:  media.Variants.Medium.Width,
				Height: media.Variants.Medium.Height,
				Size:   media.Variants.Medium.Size,
			}
		}
		
		if media.Variants.Large != nil {
			protoVariants.Large = &proto.MediaVariant{
				Url:    media.Variants.Large.URL,
				Width:  media.Variants.Large.Width,
				Height: media.Variants.Large.Height,
				Size:   media.Variants.Large.Size,
			}
		}
		
		if media.Variants.Original != nil {
			protoVariants.Original = &proto.MediaVariant{
				Url:    media.Variants.Original.URL,
				Width:  media.Variants.Original.Width,
				Height: media.Variants.Original.Height,
				Size:   media.Variants.Original.Size,
			}
		}
		
		if media.Variants.Preview != nil {
			protoVariants.Preview = &proto.MediaVariant{
				Url:    media.Variants.Preview.URL,
				Width:  media.Variants.Preview.Width,
				Height: media.Variants.Preview.Height,
				Size:   media.Variants.Preview.Size,
			}
		}
		
		if media.Variants.LowRes != nil {
			protoVariants.LowRes = &proto.MediaVariant{
				Url:    media.Variants.LowRes.URL,
				Width:  media.Variants.LowRes.Width,
				Height: media.Variants.LowRes.Height,
				Size:   media.Variants.LowRes.Size,
			}
		}
		
		if media.Variants.MidRes != nil {
			protoVariants.MidRes = &proto.MediaVariant{
				Url:    media.Variants.MidRes.URL,
				Width:  media.Variants.MidRes.Width,
				Height: media.Variants.MidRes.Height,
				Size:   media.Variants.MidRes.Size,
			}
		}
		
		if media.Variants.HighRes != nil {
			protoVariants.HighRes = &proto.MediaVariant{
				Url:    media.Variants.HighRes.URL,
				Width:  media.Variants.HighRes.Width,
				Height: media.Variants.HighRes.Height,
				Size:   media.Variants.HighRes.Size,
			}
		}
	}

	// Handle optional fields
	var altText string
	if media.AltText != nil {
		altText = *media.AltText
	}

	var postID string
	if media.PostID != nil {
		postID = *media.PostID
	}

	var thumbnailURL string
	if media.ThumbnailURL != nil {
		thumbnailURL = *media.ThumbnailURL
	}

	var processedAt string
	if media.ProcessedAt != nil {
		processedAt = media.ProcessedAt.Format(time.RFC3339)
	}

	return &proto.MediaFile{
		Id:          media.ID,
		UserId:      media.UserID,
		Filename:    media.Filename,
		Url:         media.URL,
		Type:        media.Type,
		MimeType:    media.MimeType,
		Size:        media.Size,
		AltText:     altText,
		CreatedAt:   media.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   media.UpdatedAt.Format(time.RFC3339),
		PostId:      postID,
		Status:      media.Status,
		Width:       media.Width,
		Height:      media.Height,
		Duration:    media.Duration,
		ThumbnailUrl: thumbnailURL,
		Variants:    protoVariants,
		ProcessedAt: processedAt,
	}, nil
}

// UpdateFile 更新文件记录
func (r *mediaRepository) UpdateFile(ctx context.Context, file *proto.MediaFile) error {
	updatedAt, _ := time.Parse(time.RFC3339, file.UpdatedAt)

	// 将proto.MediaVariants转换为models.MediaVariants（JSON格式）
	var variants models.MediaVariants
	if file.Variants != nil {
		if file.Variants.Thumbnail != nil {
			variants.Thumbnail = &models.MediaVariant{
				URL:    file.Variants.Thumbnail.Url,
				Width:  file.Variants.Thumbnail.Width,
				Height: file.Variants.Thumbnail.Height,
				Size:   file.Variants.Thumbnail.Size,
			}
		}
		if file.Variants.Small != nil {
			variants.Small = &models.MediaVariant{
				URL:    file.Variants.Small.Url,
				Width:  file.Variants.Small.Width,
				Height: file.Variants.Small.Height,
				Size:   file.Variants.Small.Size,
			}
		}
		if file.Variants.Medium != nil {
			variants.Medium = &models.MediaVariant{
				URL:    file.Variants.Medium.Url,
				Width:  file.Variants.Medium.Width,
				Height: file.Variants.Medium.Height,
				Size:   file.Variants.Medium.Size,
			}
		}
		if file.Variants.Large != nil {
			variants.Large = &models.MediaVariant{
				URL:    file.Variants.Large.Url,
				Width:  file.Variants.Large.Width,
				Height: file.Variants.Large.Height,
				Size:   file.Variants.Large.Size,
			}
		}
		if file.Variants.Original != nil {
			variants.Original = &models.MediaVariant{
				URL:    file.Variants.Original.Url,
				Width:  file.Variants.Original.Width,
				Height: file.Variants.Original.Height,
				Size:   file.Variants.Original.Size,
			}
		}
		// 视频特有的variants
		if file.Variants.Preview != nil {
			variants.Preview = &models.MediaVariant{
				URL:    file.Variants.Preview.Url,
				Width:  file.Variants.Preview.Width,
				Height: file.Variants.Preview.Height,
				Size:   file.Variants.Preview.Size,
			}
		}
		if file.Variants.LowRes != nil {
			variants.LowRes = &models.MediaVariant{
				URL:    file.Variants.LowRes.Url,
				Width:  file.Variants.LowRes.Width,
				Height: file.Variants.LowRes.Height,
				Size:   file.Variants.LowRes.Size,
			}
		}
		if file.Variants.MidRes != nil {
			variants.MidRes = &models.MediaVariant{
				URL:    file.Variants.MidRes.Url,
				Width:  file.Variants.MidRes.Width,
				Height: file.Variants.MidRes.Height,
				Size:   file.Variants.MidRes.Size,
			}
		}
		if file.Variants.HighRes != nil {
			variants.HighRes = &models.MediaVariant{
				URL:    file.Variants.HighRes.Url,
				Width:  file.Variants.HighRes.Width,
				Height: file.Variants.HighRes.Height,
				Size:   file.Variants.HighRes.Size,
			}
		}
	}

	updates := map[string]interface{}{
		"url":          file.Url,
		"type":         file.Type,
		"alt_text":     &file.AltText,
		"mime_type":    file.MimeType,
		"status":       file.Status,
		"variants":     variants, // 使用转换后的JSON格式variants
		"processed_at": file.ProcessedAt,
		"updated_at":   updatedAt,
	}

	return r.db.Model(&models.MediaAttachment{}).Where("id = ?", file.Id).Updates(updates).Error
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

// SaveFileToStorage 保存文件到MinIO存储
func (r *mediaRepository) SaveFileToStorage(ctx context.Context, fileID string, fileData []byte, contentType, category, userID string) (string, error) {
	// 使用MinIO存储文件
	reader := bytes.NewReader(fileData)
	fileSize := int64(len(fileData))
	return r.storageRepo.UploadFile(ctx, reader, fileSize, contentType, category, userID)
}

// DeleteFileFromStorage 从存储中删除文件
func (r *mediaRepository) DeleteFileFromStorage(ctx context.Context, url string) error {
	return r.storageRepo.DeleteFile(ctx, url)
}

// toString 将*string转换为string
func toString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// formatTimePtr 将*time.Time转换为RFC3339格式字符串
func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
