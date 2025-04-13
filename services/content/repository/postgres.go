package repository

import (
	"context"
	"errors"

	"backend/pkg/models"
	"gorm.io/gorm"
)

// PostgresRepository 是ContentRepository的PostgresSQL实现
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository 创建一个新的PostgresSQL仓库实例
func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

// 帖子CRUD操作

// CreatePost 创建新帖子
func (r *PostgresRepository) CreatePost(ctx context.Context, post *models.Post) error {
	return r.db.WithContext(ctx).Create(post).Error
}

// GetPostByID 根据ID获取帖子
func (r *PostgresRepository) GetPostByID(ctx context.Context, id string) (*models.Post, error) {
	var post models.Post
	if err := r.db.WithContext(ctx).First(&post, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("post not found")
		}
		return nil, err
	}
	return &post, nil
}

// UpdatePost 更新帖子
func (r *PostgresRepository) UpdatePost(ctx context.Context, post *models.Post) error {
	return r.db.WithContext(ctx).Save(post).Error
}

// DeletePost 删除帖子
func (r *PostgresRepository) DeletePost(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Post{}, id).Error
}

// ListPosts 获取帖子列表
func (r *PostgresRepository) ListPosts(ctx context.Context, offset, limit int) ([]*models.Post, int64, error) {
	var posts []*models.Post
	var total int64

	if err := r.db.WithContext(ctx).Model(&models.Post{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// SearchPosts 搜索帖子
func (r *PostgresRepository) SearchPosts(ctx context.Context, query string, offset, limit int) ([]*models.Post, int64, error) {
	var posts []*models.Post
	var total int64

	searchQuery := "%" + query + "%"
	queryDB := r.db.WithContext(ctx).Model(&models.Post{}).
		Where("title LIKE ? OR content LIKE ?", searchQuery, searchQuery)

	if err := queryDB.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := queryDB.Offset(offset).Limit(limit).Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// GetPostsByUser 获取用户的帖子
func (r *PostgresRepository) GetPostsByUser(ctx context.Context, userID string, offset, limit int) ([]*models.Post, int64, error) {
	var posts []*models.Post
	var total int64

	if err := r.db.WithContext(ctx).Model(&models.Post{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Offset(offset).Limit(limit).Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// GetFeaturedPosts 获取推荐帖子
func (r *PostgresRepository) GetFeaturedPosts(ctx context.Context, offset, limit int) ([]*models.Post, int64, error) {
	var posts []*models.Post
	var total int64

	if err := r.db.WithContext(ctx).Model(&models.Post{}).Where("is_featured = ?", true).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Where("is_featured = ?", true).Offset(offset).Limit(limit).Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// IncrementViewCount 增加帖子查看次数
func (r *PostgresRepository) IncrementViewCount(ctx context.Context, postID string) error {
	return r.db.WithContext(ctx).Model(&models.Post{}).Where("id = ?", postID).
		UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}

// 这里省略其他ContentRepository接口中定义的方法，实际项目中需要完整实现

// CheckConnection 提供一个方法，检查数据库连接是否正常
func (r *PostgresRepository) CheckConnection(ctx context.Context) error {
	db, err := r.db.DB()
	if err != nil {
		return err
	}
	return db.PingContext(ctx)
}
