package repository

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"backend/pkg/models"

	"github.com/google/uuid"
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
	// 生成永久链接ID (数字形式，类似Twitter的ID)
	if post.PermalinkID == "" {
		// 使用一个更安全、更唯一的算法生成permalinkID
		// 类似Twitter的ID生成算法：时间戳+机器ID+序列号
		// 这里我们使用时间戳+随机数+用户ID的部分哈希组合

		// 1. 获取当前时间毫秒时间戳（类Twitter算法使用毫秒级时间戳）
		now := time.Now().UnixNano() / int64(time.Millisecond)

		// 2. 初始化随机数生成器，确保随机数种子唯一
		randSource := rand.NewSource(now + int64(uuid.New().ID()))
		randGen := rand.New(randSource)

		// 3. 生成随机数片段（5位数），增加ID的唯一性和不可预测性
		randPart := randGen.Intn(90000) + 10000 // 10000到99999之间的随机数

		// 4. 使用用户ID的部分哈希（最后6位）增加唯一性
		userHash := 0
		for _, c := range post.UserID {
			userHash = userHash*31 + int(c)
		}
		userHash = userHash & 0xFFFFFF // 保留低24位

		// 5. 组合生成最终的permalinkID：时间戳最后10位+随机数5位+用户哈希后4位
		// 这样可以生成类似Twitter的19位数值ID
		permalinkID := fmt.Sprintf("%010d%05d%04d", now%10000000000, randPart, userHash%10000)

		// 6. 确保永久链接ID唯一
		var count int64
		for {
			// 验证生成的permalinkID是否已存在
			if err := r.db.WithContext(ctx).Model(&models.Post{}).Where("permalink_id = ?", permalinkID).Count(&count).Error; err != nil {
				return err
			}

			// 如果未找到相同的ID，则退出循环
			if count == 0 {
				break
			}

			// 否则重新生成（修改随机部分）
			randPart = randGen.Intn(90000) + 10000
			permalinkID = fmt.Sprintf("%010d%05d%04d", now%10000000000, randPart, userHash%10000)
		}

		post.PermalinkID = permalinkID
	}

	// 检查数据库是否有permalink_id列
	if err := r.db.Exec("SELECT permalink_id FROM flick_posts LIMIT 1").Error; err != nil {
		// 如果没有此列，添加该列并创建索引
		if strings.Contains(err.Error(), "column \"permalink_id\" does not exist") {
			r.db.Exec("ALTER TABLE flick_posts ADD COLUMN permalink_id VARCHAR(20)")
			// 添加唯一索引以确保永久链接ID的唯一性
			r.db.Exec("CREATE UNIQUE INDEX idx_flick_posts_permalink_id ON flick_posts (permalink_id)")
		}
	}

	// 创建帖子，但不包含permalink_id字段
	err := r.db.WithContext(ctx).Omit("permalink_id").Create(post).Error
	if err != nil {
		return err
	}

	// 如果创建成功，更新permalink_id
	if post.ID != "" {
		return r.db.WithContext(ctx).Model(post).Update("permalink_id", post.PermalinkID).Error
	}

	return nil
}

// GetPostByID 根据ID获取帖子
func (r *PostgresRepository) GetPostByID(ctx context.Context, id string) (*models.Post, error) {
	var post models.Post
	if err := r.db.WithContext(ctx).Preload("User").First(&post, "id = ?", id).Error; err != nil {
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

	if err := r.db.WithContext(ctx).Preload("User").Order("created_at DESC").Offset(offset).Limit(limit).Find(&posts).Error; err != nil {
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

	// 明确选择包含permalink_id的字段
	if err := queryDB.
		Select("id, permalink_id, user_id, type, title, content, content_html, status, media_files, poll_options, poll_ends_at, link_url, privacy_level, allow_comments, view_count, like_count, comment_count, share_count, created_at, updated_at").
		Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&posts).Error; err != nil {
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

	// 明确选择包含permalink_id的字段
	if err := r.db.WithContext(ctx).
		Select("id, permalink_id, user_id, type, title, content, content_html, status, media_files, poll_options, poll_ends_at, link_url, privacy_level, allow_comments, view_count, like_count, comment_count, share_count, created_at, updated_at").
		Where("user_id = ?", userID).
		Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&posts).Error; err != nil {
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

	// 明确选择包含permalink_id的字段
	if err := r.db.WithContext(ctx).
		Select("id, permalink_id, user_id, type, title, content, content_html, status, media_files, poll_options, poll_ends_at, link_url, privacy_level, allow_comments, view_count, like_count, comment_count, share_count, created_at, updated_at").
		Where("is_featured = ?", true).
		Preload("User").
		Offset(offset).
		Limit(limit).
		Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// IncrementViewCount 增加帖子查看次数
func (r *PostgresRepository) IncrementViewCount(ctx context.Context, postID string) error {
	return r.db.WithContext(ctx).Model(&models.Post{}).Where("id = ?", postID).
		UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}

// 帖子审计日志操作

// LogPostActivity 记录帖子活动
func (r *PostgresRepository) LogPostActivity(ctx context.Context, log *models.PostAuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetPostAuditLogs 获取帖子审计日志
func (r *PostgresRepository) GetPostAuditLogs(ctx context.Context, postID string, offset, limit int) ([]*models.PostAuditLog, int64, error) {
	var logs []*models.PostAuditLog
	var total int64

	if err := r.db.WithContext(ctx).Model(&models.PostAuditLog{}).Where("post_id = ?", postID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Where("post_id = ?", postID).Offset(offset).Limit(limit).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// 标签相关操作

// CreateTag 创建标签
func (r *PostgresRepository) CreateTag(ctx context.Context, tag *models.Tag) error {
	return r.db.WithContext(ctx).Create(tag).Error
}

// GetTagByID 根据ID获取标签
func (r *PostgresRepository) GetTagByID(ctx context.Context, id string) (*models.Tag, error) {
	var tag models.Tag
	if err := r.db.WithContext(ctx).First(&tag, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("tag not found")
		}
		return nil, err
	}
	return &tag, nil
}

// GetTagBySlug 根据Slug获取标签
func (r *PostgresRepository) GetTagBySlug(ctx context.Context, slug string) (*models.Tag, error) {
	var tag models.Tag
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&tag).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("tag not found")
		}
		return nil, err
	}
	return &tag, nil
}

// UpdateTag 更新标签
func (r *PostgresRepository) UpdateTag(ctx context.Context, tag *models.Tag) error {
	return r.db.WithContext(ctx).Save(tag).Error
}

// DeleteTag 删除标签
func (r *PostgresRepository) DeleteTag(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Tag{}, id).Error
}

// ListTags 列出标签
func (r *PostgresRepository) ListTags(ctx context.Context, offset, limit int) ([]*models.Tag, int64, error) {
	var tags []*models.Tag
	var total int64

	if err := r.db.WithContext(ctx).Model(&models.Tag{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Order("name ASC").Find(&tags).Error; err != nil {
		return nil, 0, err
	}

	return tags, total, nil
}

// GetPostTags 获取帖子的标签
func (r *PostgresRepository) GetPostTags(ctx context.Context, postID string) ([]*models.Tag, error) {
	var post models.Post
	if err := r.db.WithContext(ctx).Preload("Tags").First(&post, "id = ?", postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("post not found")
		}
		return nil, err
	}

	// 将 []models.Tag 转换为 []*models.Tag
	result := make([]*models.Tag, len(post.Tags))
	for i := range post.Tags {
		result[i] = &post.Tags[i]
	}

	return result, nil
}

// AddTagToPost 添加标签到帖子
func (r *PostgresRepository) AddTagToPost(ctx context.Context, postID, tagID string) error {
	// 检查帖子是否存在
	var post models.Post
	if err := r.db.WithContext(ctx).First(&post, "id = ?", postID).Error; err != nil {
		return err
	}

	// 检查标签是否存在
	var tag models.Tag
	if err := r.db.WithContext(ctx).First(&tag, "id = ?", tagID).Error; err != nil {
		return err
	}

	// 添加关联
	return r.db.WithContext(ctx).Exec("INSERT INTO post_tags (post_id, tag_id) VALUES (?, ?) ON CONFLICT DO NOTHING", postID, tagID).Error
}

// RemoveTagFromPost 从帖子中移除标签
func (r *PostgresRepository) RemoveTagFromPost(ctx context.Context, postID, tagID string) error {
	return r.db.WithContext(ctx).Exec("DELETE FROM post_tags WHERE post_id = ? AND tag_id = ?", postID, tagID).Error
}

// 分类相关操作

// CreateCategory 创建分类
func (r *PostgresRepository) CreateCategory(ctx context.Context, category *models.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

// GetCategoryByID 根据ID获取分类
func (r *PostgresRepository) GetCategoryByID(ctx context.Context, id string) (*models.Category, error) {
	var category models.Category
	if err := r.db.WithContext(ctx).First(&category, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}
	return &category, nil
}

// GetCategoryBySlug 根据Slug获取分类
func (r *PostgresRepository) GetCategoryBySlug(ctx context.Context, slug string) (*models.Category, error) {
	var category models.Category
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}
	return &category, nil
}

// UpdateCategory 更新分类
func (r *PostgresRepository) UpdateCategory(ctx context.Context, category *models.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

// DeleteCategory 删除分类
func (r *PostgresRepository) DeleteCategory(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Category{}, id).Error
}

// ListCategories 列出分类
func (r *PostgresRepository) ListCategories(ctx context.Context, offset, limit int) ([]*models.Category, int64, error) {
	var categories []*models.Category
	var total int64

	if err := r.db.WithContext(ctx).Model(&models.Category{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Order("display_order ASC, name ASC").Find(&categories).Error; err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

// GetCategoryHierarchy 获取分类层次结构
func (r *PostgresRepository) GetCategoryHierarchy(ctx context.Context) ([]*models.Category, error) {
	var rootCategories []*models.Category

	// 先获取所有顶级分类
	if err := r.db.WithContext(ctx).Where("parent_id IS NULL").Order("display_order ASC, name ASC").Find(&rootCategories).Error; err != nil {
		return nil, err
	}

	// 递归加载子分类
	for _, category := range rootCategories {
		if err := r.loadChildCategories(ctx, category); err != nil {
			return nil, err
		}
	}

	return rootCategories, nil
}

// loadChildCategories 递归加载子分类
func (r *PostgresRepository) loadChildCategories(ctx context.Context, parent *models.Category) error {
	var children []*models.Category
	if err := r.db.WithContext(ctx).Where("parent_id = ?", parent.ID).Order("display_order ASC, name ASC").Find(&children).Error; err != nil {
		return err
	}

	// 将 []*models.Category 转换为 []models.Category
	childrenSlice := make([]models.Category, len(children))
	for i, child := range children {
		childrenSlice[i] = *child
	}

	parent.Children = childrenSlice

	// 递归加载子分类的子分类
	for _, child := range children {
		if err := r.loadChildCategories(ctx, child); err != nil {
			return err
		}
	}

	return nil
}

// GetPostCategories 获取帖子的分类
func (r *PostgresRepository) GetPostCategories(ctx context.Context, postID string) ([]*models.Category, error) {
	var post models.Post
	if err := r.db.WithContext(ctx).Preload("Categories").First(&post, "id = ?", postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("post not found")
		}
		return nil, err
	}

	// 将 []models.Category 转换为 []*models.Category
	result := make([]*models.Category, len(post.Categories))
	for i := range post.Categories {
		result[i] = &post.Categories[i]
	}

	return result, nil
}

// AddCategoryToPost 添加分类到帖子
func (r *PostgresRepository) AddCategoryToPost(ctx context.Context, postID, categoryID string) error {
	// 检查帖子是否存在
	var post models.Post
	if err := r.db.WithContext(ctx).First(&post, "id = ?", postID).Error; err != nil {
		return err
	}

	// 检查分类是否存在
	var category models.Category
	if err := r.db.WithContext(ctx).First(&category, "id = ?", categoryID).Error; err != nil {
		return err
	}

	// 添加关联
	return r.db.WithContext(ctx).Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?) ON CONFLICT DO NOTHING", postID, categoryID).Error
}

// RemoveCategoryFromPost 从帖子中移除分类
func (r *PostgresRepository) RemoveCategoryFromPost(ctx context.Context, postID, categoryID string) error {
	return r.db.WithContext(ctx).Exec("DELETE FROM post_categories WHERE post_id = ? AND category_id = ?", postID, categoryID).Error
}

// GetPostsByCategory 获取分类下的帖子
func (r *PostgresRepository) GetPostsByCategory(ctx context.Context, categoryID string, offset, limit int) ([]*models.Post, int64, error) {
	var posts []*models.Post
	var total int64

	// 创建子查询获取指定分类下的帖子ID
	postIDs := r.db.Table("post_categories").Select("post_id").Where("category_id = ?", categoryID)

	// 统计总数
	if err := r.db.WithContext(ctx).Model(&models.Post{}).Where("id IN (?)", postIDs).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取帖子
	if err := r.db.WithContext(ctx).
		Select("id, permalink_id, user_id, type, title, content, content_html, status, media_files, poll_options, poll_ends_at, link_url, privacy_level, allow_comments, view_count, like_count, comment_count, share_count, created_at, updated_at").
		Where("id IN (?)", postIDs).
		Preload("User").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// GetPostsByTag 获取标签下的帖子
func (r *PostgresRepository) GetPostsByTag(ctx context.Context, tagID string, offset, limit int) ([]*models.Post, int64, error) {
	var posts []*models.Post
	var total int64

	// 创建子查询获取指定标签下的帖子ID
	postIDs := r.db.Table("post_tags").Select("post_id").Where("tag_id = ?", tagID)

	// 统计总数
	if err := r.db.WithContext(ctx).Model(&models.Post{}).Where("id IN (?)", postIDs).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取帖子
	if err := r.db.WithContext(ctx).
		Select("id, permalink_id, user_id, type, title, content, content_html, status, media_files, poll_options, poll_ends_at, link_url, privacy_level, allow_comments, view_count, like_count, comment_count, share_count, created_at, updated_at").
		Where("id IN (?)", postIDs).
		Preload("User").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// 帖子保存相关操作

// SavePost 保存帖子
func (r *PostgresRepository) SavePost(ctx context.Context, savedPost *models.SavedPost) error {
	return r.db.WithContext(ctx).Create(savedPost).Error
}

// UnsavePost 取消保存帖子
func (r *PostgresRepository) UnsavePost(ctx context.Context, userID, postID string) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND post_id = ?", userID, postID).Delete(&models.SavedPost{}).Error
}

// GetSavedPosts 获取用户保存的帖子
func (r *PostgresRepository) GetSavedPosts(ctx context.Context, userID string, collection string, offset, limit int) ([]*models.Post, int64, error) {
	var posts []*models.Post
	var total int64

	// 创建查询构建器
	query := r.db.WithContext(ctx).Table("flick_posts").
		Select("flick_posts.id, flick_posts.permalink_id, flick_posts.user_id, flick_posts.type, flick_posts.title, flick_posts.content, flick_posts.content_html, flick_posts.status, flick_posts.media_files, flick_posts.poll_options, flick_posts.poll_ends_at, flick_posts.link_url, flick_posts.privacy_level, flick_posts.allow_comments, flick_posts.view_count, flick_posts.like_count, flick_posts.comment_count, flick_posts.share_count, flick_posts.created_at, flick_posts.updated_at").
		Joins("JOIN flick_saved_posts ON flick_posts.id = flick_saved_posts.post_id").
		Where("flick_saved_posts.user_id = ?", userID)

	// 如果指定了收藏夹，添加条件
	if collection != "" {
		query = query.Where("flick_saved_posts.collection = ?", collection)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取帖子列表
	if err := query.Preload("User").Offset(offset).Limit(limit).Order("flick_saved_posts.created_at DESC").Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// IsSaved 检查帖子是否被用户保存
func (r *PostgresRepository) IsSaved(ctx context.Context, userID, postID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.SavedPost{}).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Count(&count).Error
	return count > 0, err
}

// 帖子媒体相关操作

// UpdatePostMedia 更新帖子的媒体文件
func (r *PostgresRepository) UpdatePostMedia(ctx context.Context, postID string, media models.MediaFiles) error {
	return r.db.WithContext(ctx).Model(&models.Post{}).Where("id = ?", postID).UpdateColumn("media_files", media).Error
}

// 投票相关操作

// UpdatePostPoll 更新帖子的投票选项
func (r *PostgresRepository) UpdatePostPoll(ctx context.Context, postID string, options models.PollOptions, endsAt time.Time) error {
	return r.db.WithContext(ctx).Model(&models.Post{}).Where("id = ?", postID).Updates(map[string]interface{}{
		"poll_options": options,
		"poll_ends_at": endsAt,
	}).Error
}

// VoteInPoll 在投票中投票
func (r *PostgresRepository) VoteInPoll(ctx context.Context, vote *models.PollVote) error {
	// 开启事务
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 判断是否已经投过票
	var existingVote models.PollVote
	if err := tx.Where("user_id = ? AND post_id = ?", vote.UserID, vote.PostID).First(&existingVote).Error; err == nil {
		// 如果用户已经投票，返回错误
		return errors.New("user has already voted in this poll")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// 如果出现其他错误，返回错误
		return err
	}

	// 创建投票记录
	if err := tx.Create(vote).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 增加选项计数
	var post models.Post
	if err := tx.First(&post, "id = ?", vote.PostID).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 更新选项的计数
	updated := false
	for i, option := range post.PollOptions {
		if option.ID == vote.OptionID {
			post.PollOptions[i].Count++
			updated = true
			break
		}
	}

	if !updated {
		tx.Rollback()
		return errors.New("option not found")
	}

	// 保存更新后的投票选项
	if err := tx.Model(&post).UpdateColumn("poll_options", post.PollOptions).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	return tx.Commit().Error
}

// GetPollVotes 获取投票结果
func (r *PostgresRepository) GetPollVotes(ctx context.Context, postID string) (*models.PollOptions, int64, error) {
	var post models.Post
	if err := r.db.WithContext(ctx).Select("poll_options").First(&post, "id = ?", postID).Error; err != nil {
		return nil, 0, err
	}

	// 计算总票数
	total := int64(0)
	for _, option := range post.PollOptions {
		total += int64(option.Count)
	}

	return &post.PollOptions, total, nil
}

// HasVoted 检查用户是否已经投票
func (r *PostgresRepository) HasVoted(ctx context.Context, userID, postID string) (bool, string, error) {
	var vote models.PollVote
	err := r.db.WithContext(ctx).Where("user_id = ? AND post_id = ?", userID, postID).First(&vote).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, "", nil
		}
		return false, "", err
	}
	return true, vote.OptionID, nil
}

// 统计相关操作

// GetPostStats 获取帖子统计数据
func (r *PostgresRepository) GetPostStats(ctx context.Context, postID string) (map[string]int64, error) {
	var post models.Post
	if err := r.db.WithContext(ctx).Select("view_count, like_count, comment_count, share_count").First(&post, "id = ?", postID).Error; err != nil {
		return nil, err
	}

	// 获取投票总数
	var voteCount int64
	if err := r.db.WithContext(ctx).Model(&models.PollVote{}).Where("post_id = ?", postID).Count(&voteCount).Error; err != nil {
		return nil, err
	}

	// 获取收藏总数
	var saveCount int64
	if err := r.db.WithContext(ctx).Model(&models.SavedPost{}).Where("post_id = ?", postID).Count(&saveCount).Error; err != nil {
		return nil, err
	}

	// 返回统计数据
	return map[string]int64{
		"view_count":    int64(post.ViewCount),
		"like_count":    int64(post.LikeCount),
		"comment_count": int64(post.CommentCount),
		"share_count":   int64(post.ShareCount),
		"vote_count":    voteCount,
		"save_count":    saveCount,
	}, nil
}

// GetTopPosts 获取热门帖子
func (r *PostgresRepository) GetTopPosts(ctx context.Context, period string, limit int) ([]*models.Post, error) {
	var posts []*models.Post
	query := r.db.WithContext(ctx).Model(&models.Post{})

	// 根据时间段过滤
	switch period {
	case "day":
		query = query.Where("created_at >= ?", time.Now().AddDate(0, 0, -1))
	case "week":
		query = query.Where("created_at >= ?", time.Now().AddDate(0, 0, -7))
	case "month":
		query = query.Where("created_at >= ?", time.Now().AddDate(0, -1, 0))
	case "year":
		query = query.Where("created_at >= ?", time.Now().AddDate(-1, 0, 0))
	}

	// 按照评分（可以自定义评分规则）排序
	// 这里使用简单的权重计算: 浏览量*0.1 + 点赞数*0.3 + 评论数*0.4 + 分享数*0.2
	query = query.Order("(view_count * 0.1) + (like_count * 0.3) + (comment_count * 0.4) + (share_count * 0.2) DESC")

	// 获取限定数量的热门帖子
	if err := query.
		Select("id, permalink_id, user_id, type, title, content, content_html, status, media_files, poll_options, poll_ends_at, link_url, privacy_level, allow_comments, view_count, like_count, comment_count, share_count, created_at, updated_at").
		Preload("User").
		Limit(limit).
		Find(&posts).Error; err != nil {
		return nil, err
	}

	return posts, nil
}

// GetUserContentStats 获取用户内容统计数据
func (r *PostgresRepository) GetUserContentStats(ctx context.Context, userID string) (map[string]int64, error) {
	stats := make(map[string]int64)

	// 获取用户的帖子数量
	var postCount int64
	if err := r.db.WithContext(ctx).Model(&models.Post{}).Where("user_id = ?", userID).Count(&postCount).Error; err != nil {
		return nil, err
	}
	stats["post_count"] = postCount

	// 获取用户的评论数量
	var commentCount int64
	if err := r.db.WithContext(ctx).Model(&models.Comment{}).Where("user_id = ?", userID).Count(&commentCount).Error; err != nil {
		return nil, err
	}
	stats["comment_count"] = commentCount

	// 获取用户收到的点赞数量
	var likeCount int64
	if err := r.db.WithContext(ctx).Model(&models.PostLike{}).
		Joins("JOIN flick_posts ON flick_post_likes.post_id = flick_posts.id").
		Where("flick_posts.user_id = ?", userID).
		Count(&likeCount).Error; err != nil {
		return nil, err
	}
	stats["received_like_count"] = likeCount

	// 获取用户的收藏数量
	var saveCount int64
	if err := r.db.WithContext(ctx).Model(&models.SavedPost{}).Where("user_id = ?", userID).Count(&saveCount).Error; err != nil {
		return nil, err
	}
	stats["saved_post_count"] = saveCount

	return stats, nil
}

// 帖子举报相关操作

// CreatePostReport 创建帖子举报
func (r *PostgresRepository) CreatePostReport(ctx context.Context, report *models.PostReport) error {
	return r.db.WithContext(ctx).Create(report).Error
}

// GetPostReportByID 根据ID获取帖子举报
func (r *PostgresRepository) GetPostReportByID(ctx context.Context, id string) (*models.PostReport, error) {
	var report models.PostReport
	if err := r.db.WithContext(ctx).First(&report, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("report not found")
		}
		return nil, err
	}
	return &report, nil
}

// UpdatePostReport 更新帖子举报
func (r *PostgresRepository) UpdatePostReport(ctx context.Context, report *models.PostReport) error {
	return r.db.WithContext(ctx).Save(report).Error
}

// GetPostReports 获取帖子举报列表
func (r *PostgresRepository) GetPostReports(ctx context.Context, status string, offset, limit int) ([]*models.PostReport, int64, error) {
	var reports []*models.PostReport
	var total int64
	query := r.db.WithContext(ctx).Model(&models.PostReport{})

	// 如果指定了状态，添加过滤条件
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取举报列表
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

// GetUserPostReports 获取用户提交的帖子举报
func (r *PostgresRepository) GetUserPostReports(ctx context.Context, userID string, offset, limit int) ([]*models.PostReport, int64, error) {
	var reports []*models.PostReport
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&models.PostReport{}).Where("reporter_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取举报列表
	if err := r.db.WithContext(ctx).Where("reporter_id = ?", userID).Offset(offset).Limit(limit).Order("created_at DESC").Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

// CheckConnection 提供一个方法，检查数据库连接是否正常
func (r *PostgresRepository) CheckConnection(ctx context.Context) error {
	db, err := r.db.DB()
	if err != nil {
		return err
	}
	return db.PingContext(ctx)
}

// IncrementUserPostCount 增加用户的帖子计数
func (r *PostgresRepository) IncrementUserPostCount(ctx context.Context, userID string) error {
	// 这里我们可能需要调用用户服务或者直接更新用户表
	// 假设我们有一个user_stats表来存储用户统计信息

	// 检查用户统计记录是否存在
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.UserStats{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return err
	}

	// 如果记录不存在，创建一个新记录
	if count == 0 {
		userStats := &models.UserStats{
			UserID:    userID,
			PostCount: 1,
		}
		return r.db.WithContext(ctx).Create(userStats).Error
	}

	// 更新帖子计数
	return r.db.WithContext(ctx).Model(&models.UserStats{}).
		Where("user_id = ?", userID).
		UpdateColumn("post_count", gorm.Expr("post_count + ?", 1)).
		UpdateColumn("updated_at", time.Now()).
		Error
}

// SaveContentMedia 保存媒体文件记录到数据库
func (r *PostgresRepository) SaveContentMedia(ctx context.Context, media *models.ContentMedia) error {
	// 确保ID不为空
	if media.ID == "" {
		media.ID = uuid.New().String()
	}

	// 确保创建时间不为空
	if media.CreatedAt.IsZero() {
		media.CreatedAt = time.Now()
	}

	// 更新时间设置为当前时间
	media.UpdatedAt = time.Now()

	// 使用GORM的方式保存，并通过map直接设置NULL值
	return r.db.WithContext(ctx).Model(&models.ContentMedia{}).Create(map[string]interface{}{
		"id":           media.ID,
		"user_id":      media.UserID,
		"media_type":   media.MediaType,
		"url":          media.URL,
		"file_name":    media.FileName,
		"file_size":    media.FileSize,
		"content_type": media.ContentType,
		"object_name":  media.ObjectName,
		"width":        media.Width,
		"height":       media.Height,
		"duration":     media.Duration,
		"description":  media.Description,
		"post_id":      nil, // 明确设置为nil而不是空字符串
		"created_at":   media.CreatedAt,
		"updated_at":   media.UpdatedAt,
	}).Error
}

// GetContentMediaByID 根据ID获取媒体文件记录
func (r *PostgresRepository) GetContentMediaByID(ctx context.Context, id string) (*models.ContentMedia, error) {
	var media models.ContentMedia
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&media).Error
	if err != nil {
		return nil, err
	}
	return &media, nil
}

// ListUserContentMedia 获取用户的媒体文件列表
func (r *PostgresRepository) ListUserContentMedia(ctx context.Context, userID string, offset, limit int) ([]*models.ContentMedia, int64, error) {
	var media []*models.ContentMedia
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&models.ContentMedia{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取数据
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Offset(offset).Limit(limit).Find(&media).Error; err != nil {
		return nil, 0, err
	}

	return media, total, nil
}

// DeleteContentMedia 删除媒体文件记录
func (r *PostgresRepository) DeleteContentMedia(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.ContentMedia{}, "id = ?", id).Error
}

// FindDuplicateMedia 查找可能的重复媒体文件
func (r *PostgresRepository) FindDuplicateMedia(ctx context.Context, userID, mediaType, contentType string, fileSize int64, objectName string) (*models.ContentMedia, error) {
	var media models.ContentMedia

	// 基于用户ID、文件大小和对象名称来检查是否存在重复文件
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND media_type = ? AND content_type = ? AND file_size = ?",
			userID, mediaType, contentType, fileSize).
		First(&media).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &media, nil
}

// GetPostByPermalink 根据永久链接ID获取帖子
func (r *PostgresRepository) GetPostByPermalink(ctx context.Context, permalinkID string) (*models.Post, error) {
	var post models.Post

	// 参数验证 - 确保permalinkID不为空且长度合理
	if permalinkID == "" {
		return nil, errors.New("permalink id cannot be empty")
	}

	if len(permalinkID) > 25 {
		return nil, fmt.Errorf("permalink id too long: %s", permalinkID)
	}

	// 打印调试日志
	log.Info().
		Str("permalink_id", permalinkID).
		Msg("开始查询permalink_id")

	// 使用正确的表名称flick_users进行关联查询
	err := r.db.WithContext(ctx).
		Table("flick_posts").
		Select("flick_posts.*, flick_users.*").
		Joins("JOIN flick_users ON flick_posts.user_id = flick_users.id").
		Where("flick_posts.permalink_id = ?", permalinkID).
		Scan(&post).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 记录详细的错误信息
			log.Error().
				Str("permalink_id", permalinkID).
				Msg("使用permalink_id未找到帖子")

			// 尝试直接查询帖子，不加载用户信息
			err = r.db.WithContext(ctx).
				Where("permalink_id = ?", permalinkID).
				First(&post).Error

			if err != nil {
				log.Error().
					Err(err).
					Str("permalink_id", permalinkID).
					Msg("直接查询帖子也失败")
				return nil, fmt.Errorf("post with permalink_id %s not found", permalinkID)
			}
		} else {
			log.Error().
				Err(err).
				Str("permalink_id", permalinkID).
				Msg("查询数据库时出错")

			return nil, fmt.Errorf("database error: %w", err)
		}
	}

	// 记录查询成功的日志
	log.Info().
		Str("post_id", post.ID).
		Str("permalink_id", post.PermalinkID).
		Str("user_id", post.UserID).
		Time("created_at", post.CreatedAt).
		Msg("成功查询到帖子")

	// 单独获取用户信息
	if post.User.ID == "" && post.UserID != "" {
		log.Info().
			Str("post_id", post.ID).
			Str("user_id", post.UserID).
			Msg("需要单独加载用户信息")

		var user models.User
		// 修改查询表名为正确的flick_users表
		userErr := r.db.WithContext(ctx).Table("flick_users").Where("id = ?", post.UserID).First(&user).Error
		if userErr == nil {
			post.User = user
			log.Info().
				Str("user_id", user.ID).
				Str("username", user.Username).
				Msg("已成功加载用户信息")
		} else {
			log.Error().
				Err(userErr).
				Str("user_id", post.UserID).
				Msg("加载用户信息失败")
		}
	}

	return &post, nil
}

// GetUserByID 根据ID获取用户信息
func (r *PostgresRepository) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Table("flick_users").Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// IsPostLikedByUser 检查帖子是否被特定用户点赞
func (r *PostgresRepository) IsPostLikedByUser(ctx context.Context, postID string, userID string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.PostLike{}).
		Where("post_id = ? AND user_id = ?", postID, userID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
