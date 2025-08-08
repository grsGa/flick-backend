package repository

import (
	"context"
	"time"

	"backend/pkg/database"
	"backend/pkg/models"
	"backend/services/bookmark/proto"

	"gorm.io/gorm"
)

// bookmarkRepository 书签仓储实现
type bookmarkRepository struct {
	db *gorm.DB
}

// NewBookmarkRepository 创建书签仓储实例
func NewBookmarkRepository() BookmarkRepository {
	return &bookmarkRepository{
		db: database.GetDB(),
	}
}

// CreateBookmark 创建书签
func (r *bookmarkRepository) CreateBookmark(ctx context.Context, bookmark *proto.Bookmark) error {
	createdAt, _ := time.Parse(time.RFC3339, bookmark.CreatedAt)
	b := &models.Bookmark{
		ID:        bookmark.Id,
		UserID:    bookmark.UserId,
		PostID:    bookmark.PostId,
		CreatedAt: createdAt,
	}

	return r.db.Create(b).Error
}

// DeleteBookmark 删除书签
func (r *bookmarkRepository) DeleteBookmark(ctx context.Context, userID, postID string) error {
	return r.db.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&models.Bookmark{}).Error
}

// IsBookmarked 检查是否已收藏
func (r *bookmarkRepository) IsBookmarked(ctx context.Context, userID, postID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Bookmark{}).
		Where("user_id = ? AND post_id = ? AND deleted_at IS NULL", userID, postID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ListBookmarks 获取用户书签列表
func (r *bookmarkRepository) ListBookmarks(ctx context.Context, userID string, page, pageSize int32) ([]*proto.Bookmark, int32, error) {
	var bookmarks []models.Bookmark
	var total int64

	// 查询总数
	if err := r.db.Model(&models.Bookmark{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询列表
	offset := (page - 1) * pageSize
	if err := r.db.Where("user_id = ? AND deleted_at IS NULL", userID).
		Offset(int(offset)).
		Limit(int(pageSize)).
		Find(&bookmarks).Error; err != nil {
		return nil, 0, err
	}

	protoBookmarks := make([]*proto.Bookmark, len(bookmarks))
	for i, bookmark := range bookmarks {
		protoBookmarks[i] = &proto.Bookmark{
			Id:        bookmark.ID,
			UserId:    bookmark.UserID,
			PostId:    bookmark.PostID,
			CreatedAt: bookmark.CreatedAt.Format(time.RFC3339),
		}
	}

	return protoBookmarks, int32(total), nil
}
