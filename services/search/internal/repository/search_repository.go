package repository

import (
	"context"
	"time"

	"github.com/flick/backend/pkg/database"
	"github.com/flick/backend/pkg/models"
	"github.com/flick/backend/services/search/proto"
	"gorm.io/gorm"
)

// searchRepository 搜索仓储实现
type searchRepository struct {
	db *gorm.DB
}

// NewSearchRepository 创建搜索仓储实例
func NewSearchRepository() SearchRepository {
	return &searchRepository{
		db: database.GetDB(),
	}
}

// SearchContent 搜索内容
func (r *searchRepository) SearchContent(ctx context.Context, query string, page, pageSize int32, sortBy string) ([]*proto.SearchResultItem, int32, error) {
	var posts []models.Post
	var total int64

	// 构建查询
	dbQuery := r.db.Model(&models.Post{}).Where("content ILIKE ?", "%"+query+"%").Where("deleted_at IS NULL")

	// 计算总数
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	switch sortBy {
	case "latest":
		dbQuery = dbQuery.Order("created_at DESC")
	case "popular":
		// 简化处理，实际应该根据点赞数等排序
		dbQuery = dbQuery.Order("created_at DESC")
	default:
		// 默认按相关性排序，这里简化处理
		dbQuery = dbQuery.Order("created_at DESC")
	}

	// 分页
	offset := (page - 1) * pageSize
	if err := dbQuery.Offset(int(offset)).Limit(int(pageSize)).Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*proto.SearchResultItem, len(posts))
	for i, post := range posts {
		// 获取作者信息
		var user models.User
		r.db.Where("id = ?", post.UserID).First(&user)

		// 获取媒体附件
		var mediaAttachments []models.MediaAttachment
		r.db.Where("post_id = ?", post.ID).Find(&mediaAttachments)

		// 获取点赞数
		var likeCount int64
		r.db.Model(&models.Like{}).Where("post_id = ?", post.ID).Count(&likeCount)

		// 获取回复数
		var replyCount int64
		r.db.Model(&models.Post{}).Where("parent_id = ?", post.ID).Count(&replyCount)

		items[i] = &proto.SearchResultItem{
			Id:           post.ID,
			Type:         "post",
			Content:      post.Content,
			Author:       user.Username,
			AvatarUrl:    user.AvatarURL,
			LikeCount:    int32(likeCount),
			ReplyCount:   int32(replyCount),
			CreatedAt:    post.CreatedAt.Format(time.RFC3339),
			Hashtags:     []string{}, // 简化处理，实际应从内容中提取标签
		}
	}

	return items, int32(total), nil
}

// SearchUsers 搜索用户
func (r *searchRepository) SearchUsers(ctx context.Context, query string, page, pageSize int32) ([]*proto.SearchResultItem, int32, error) {
	var users []models.User
	var total int64

	// 构建查询
	dbQuery := r.db.Model(&models.User{}).
		Where("(username ILIKE ? OR display_name ILIKE ? OR bio ILIKE ?)", "%"+query+"%", "%"+query+"%", "%"+query+"%").
		Where("deleted_at IS NULL")

	// 计算总数
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序和分页
	offset := (page - 1) * pageSize
	if err := dbQuery.Offset(int(offset)).Limit(int(pageSize)).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*proto.SearchResultItem, len(users))
	for i, user := range users {
		items[i] = &proto.SearchResultItem{
			Id:        user.ID,
			Type:      "user",
			Title:     user.Username,
			Content:   toString(user.Bio),
			Author:    user.Username,
			AvatarUrl: user.AvatarURL,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
		}
	}

	return items, int32(total), nil
}

// SearchHashtags 搜索标签
func (r *searchRepository) SearchHashtags(ctx context.Context, query string, page, pageSize int32) ([]*proto.SearchResultItem, int32, error) {
	// 简化处理，实际应从帖子内容中提取和搜索标签
	items := make([]*proto.SearchResultItem, 0)
	total := int32(0)

	return items, total, nil
}

// toString 将*string转换为string
func toString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
