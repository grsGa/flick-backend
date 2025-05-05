package repository

import (
	"context"
	"errors"

	"backend/pkg/models"

	"gorm.io/gorm"
)

// PostgresRepository 是交互服务的PostgreSQL实现
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository 创建一个新的PostgreSQL仓库实例
func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

// CheckConnection 检查数据库连接是否正常
func (r *PostgresRepository) CheckConnection(ctx context.Context) error {
	db, err := r.db.DB()
	if err != nil {
		return err
	}
	return db.PingContext(ctx)
}

// 以下是接口方法的实现，这里只提供一个示例，实际项目中需要实现所有接口方法

// CreatePostLike 创建帖子点赞
func (r *PostgresRepository) CreatePostLike(ctx context.Context, like *models.PostLike) error {
	return r.db.WithContext(ctx).Create(like).Error
}

// DeletePostLike 删除帖子点赞
func (r *PostgresRepository) DeletePostLike(ctx context.Context, userID, postID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Delete(&models.PostLike{}).Error
}

// GetPostLike 获取帖子点赞
func (r *PostgresRepository) GetPostLike(ctx context.Context, userID, postID string) (*models.PostLike, error) {
	var like models.PostLike
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND post_id = ?", userID, postID).
		First(&like).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("post like not found")
		}
		return nil, err
	}
	return &like, nil
}

// GetPostLikes 获取帖子的所有点赞用户
func (r *PostgresRepository) GetPostLikes(ctx context.Context, postID string, offset, limit int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	query := r.db.WithContext(ctx).
		Table("post_likes").
		Select("users.*").
		Joins("JOIN users ON post_likes.user_id = users.id").
		Where("post_likes.post_id = ?", postID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetUserLikedPosts 获取用户点赞的所有帖子
func (r *PostgresRepository) GetUserLikedPosts(ctx context.Context, userID string, offset, limit int) ([]*models.Post, int64, error) {
	var posts []*models.Post
	var total int64

	query := r.db.WithContext(ctx).
		Table("post_likes").
		Select("posts.*").
		Joins("JOIN posts ON post_likes.post_id = posts.id").
		Where("post_likes.user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// CreateBookmark 创建书签
func (r *PostgresRepository) CreateBookmark(ctx context.Context, bookmark *models.Bookmark) error {
	return r.db.WithContext(ctx).Create(bookmark).Error
}

// DeleteBookmark 删除书签
func (r *PostgresRepository) DeleteBookmark(ctx context.Context, userID, postID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Delete(&models.Bookmark{}).Error
}

// GetBookmark 获取书签
func (r *PostgresRepository) GetBookmark(ctx context.Context, userID, postID string) (*models.Bookmark, error) {
	var bookmark models.Bookmark
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND post_id = ?", userID, postID).
		First(&bookmark).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("bookmark not found")
		}
		return nil, err
	}
	return &bookmark, nil
}

// GetUserBookmarks 获取用户收藏的所有帖子
func (r *PostgresRepository) GetUserBookmarks(ctx context.Context, userID string, collectionName string, offset, limit int) ([]*models.Post, int64, error) {
	var posts []*models.Post
	var total int64

	query := r.db.WithContext(ctx).
		Table("bookmarks").
		Select("posts.*").
		Joins("JOIN posts ON bookmarks.post_id = posts.id").
		Where("bookmarks.user_id = ?", userID)

	// 如果指定了收藏夹，添加条件
	if collectionName != "" {
		query = query.Where("bookmarks.collection_name = ?", collectionName)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// CreateBookmarkCollection 创建书签收藏夹
func (r *PostgresRepository) CreateBookmarkCollection(ctx context.Context, collection *models.BookmarkCollection) error {
	return r.db.WithContext(ctx).Create(collection).Error
}

// GetUserBookmarkCollections 获取用户的所有收藏夹
func (r *PostgresRepository) GetUserBookmarkCollections(ctx context.Context, userID string) ([]*models.BookmarkCollection, error) {
	var collections []*models.BookmarkCollection
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&collections).Error; err != nil {
		return nil, err
	}
	return collections, nil
}

// 其他接口方法需要在实际项目中完整实现
