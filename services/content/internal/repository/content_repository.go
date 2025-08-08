package repository

import (
	"context"
	"errors"
	"time"

	"backend/pkg/database"
	"backend/pkg/models"
	"backend/services/content/proto"

	"gorm.io/gorm"
)

// contentRepository 内容仓储实现
type contentRepository struct {
	db *gorm.DB
}

// NewContentRepository 创建内容仓储实例
func NewContentRepository() ContentRepository {
	return &contentRepository{
		db: database.GetDB(),
	}
}

// GetContentByID 根据ID获取内容
func (r *contentRepository) GetContentByID(ctx context.Context, id string) (*proto.Content, error) {
	var post models.Post
	if err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&post).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("content not found")
		}
		return nil, err
	}

	// 获取媒体附件
	var mediaAttachments []models.MediaAttachment
	r.db.Where("post_id = ? AND deleted_at IS NULL", post.ID).Find(&mediaAttachments)

	mediaFiles := make([]*proto.MediaFile, len(mediaAttachments))
	for i, media := range mediaAttachments {
		mediaFiles[i] = &proto.MediaFile{
			Id:   media.ID,
			Url:  media.URL,
			Type: media.Type,
		}
	}

	return &proto.Content{
		Id:           post.ID,
		UserId:       post.UserID,
		Title:        "", // 标题字段在模型中不存在，可能需要调整
		Body:         post.Content,
		MediaFiles:   mediaFiles,
		LikeCount:    0, // 需要从likes表统计
		CommentCount: 0, // 需要从posts表统计回复数量
		CreatedAt:    post.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    post.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// CreateContent 创建内容
func (r *contentRepository) CreateContent(ctx context.Context, content *proto.Content) error {
	// 创建帖子
	createdAt, _ := time.Parse(time.RFC3339, content.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, content.UpdatedAt)
	post := &models.Post{
		ID:        content.Id,
		UserID:    content.UserId,
		Content:   content.Body,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	if err := r.db.Create(post).Error; err != nil {
		return err
	}

	// 创建媒体附件
	for _, mediaFile := range content.MediaFiles {
		media := &models.MediaAttachment{
			ID:     mediaFile.Id,
			PostID: content.Id,
			URL:    mediaFile.Url,
			Type:   mediaFile.Type,
		}

		if err := r.db.Create(media).Error; err != nil {
			return err
		}
	}

	return nil
}

// UpdateContent 更新内容
func (r *contentRepository) UpdateContent(ctx context.Context, content *proto.Content) error {
	// 更新帖子
	updatedAt, _ := time.Parse(time.RFC3339, content.UpdatedAt)
	post := &models.Post{
		ID:        content.Id,
		Content:   content.Body,
		UpdatedAt: updatedAt,
	}

	if err := r.db.Where("id = ? AND deleted_at IS NULL", content.Id).Updates(post).Error; err != nil {
		return err
	}

	// 更新媒体附件（简化处理，实际应支持增删改）
	for _, mediaFile := range content.MediaFiles {
		media := &models.MediaAttachment{
			ID:   mediaFile.Id,
			URL:  mediaFile.Url,
			Type: mediaFile.Type,
		}

		if err := r.db.Where("id = ?", mediaFile.Id).Updates(media).Error; err != nil {
			return err
		}
	}

	return nil
}

// DeleteContent 删除内容
func (r *contentRepository) DeleteContent(ctx context.Context, id string) error {
	// 软删除帖子
	if err := r.db.Where("id = ?", id).Delete(&models.Post{}).Error; err != nil {
		return err
	}

	// 同时软删除相关的媒体附件
	return r.db.Where("post_id = ?", id).Delete(&models.MediaAttachment{}).Error
}

// ListContent 列出内容
func (r *contentRepository) ListContent(ctx context.Context, userID string, page, pageSize int32) ([]*proto.Content, int32, error) {
	var posts []models.Post
	var total int64

	// 查询总数
	if err := r.db.Model(&models.Post{}).Where("user_id = ? AND deleted_at IS NULL", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询帖子列表
	offset := (page - 1) * pageSize
	if err := r.db.Where("user_id = ? AND deleted_at IS NULL", userID).Offset(int(offset)).Limit(int(pageSize)).Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	contents := make([]*proto.Content, len(posts))
	for i, post := range posts {
		// 获取媒体附件
		var mediaAttachments []models.MediaAttachment
		r.db.Where("post_id = ? AND deleted_at IS NULL", post.ID).Find(&mediaAttachments)

		mediaFiles := make([]*proto.MediaFile, len(mediaAttachments))
		for j, media := range mediaAttachments {
			mediaFiles[j] = &proto.MediaFile{
				Id:   media.ID,
				Url:  media.URL,
				Type: media.Type,
			}
		}

		contents[i] = &proto.Content{
			Id:         post.ID,
			UserId:     post.UserID,
			Body:       post.Content,
			MediaFiles: mediaFiles,
			CreatedAt:  post.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:  post.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return contents, int32(total), nil
}
