package repository

import (
	"context"
	"errors"

	"backend/pkg/models"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// InteractionRepository 定义交互服务的数据存储接口
type InteractionRepository interface {
	// 帖子点赞相关
	CreatePostLike(ctx context.Context, like *models.PostLike) error
	DeletePostLike(ctx context.Context, userID, postID string) error
	GetPostLike(ctx context.Context, userID, postID string) (*models.PostLike, error)
	GetPostLikes(ctx context.Context, postID string, offset, limit int) ([]*models.User, int64, error)
	GetUserLikedPosts(ctx context.Context, userID string, offset, limit int) ([]*models.Post, int64, error)

	// 收藏相关
	CreateBookmark(ctx context.Context, bookmark *models.Bookmark) error
	DeleteBookmark(ctx context.Context, userID, postID string) error
	GetBookmark(ctx context.Context, userID, postID string) (*models.Bookmark, error)
	GetUserBookmarks(ctx context.Context, userID string, collectionName string, offset, limit int) ([]*models.Post, int64, error)
	CreateBookmarkCollection(ctx context.Context, collection *models.BookmarkCollection) error
	GetUserBookmarkCollections(ctx context.Context, userID string) ([]*models.BookmarkCollection, error)

	// 评论相关
	CreateComment(ctx context.Context, comment *models.Comment) error
	GetCommentByID(ctx context.Context, commentID string) (*models.Comment, error)
	UpdateComment(ctx context.Context, comment *models.Comment) error
	DeleteComment(ctx context.Context, commentID string) error
	GetPostComments(ctx context.Context, postID string, offset, limit int) ([]*models.Comment, int64, error)
	GetUserComments(ctx context.Context, userID string, offset, limit int) ([]*models.Comment, int64, error)

	// 评论回复
	CreateCommentReply(ctx context.Context, reply *models.CommentReply) error
	GetCommentReplies(ctx context.Context, commentID string, offset, limit int) ([]*models.CommentReply, int64, error)
	GetReplyByID(ctx context.Context, replyID string) (*models.CommentReply, error)
	UpdateReply(ctx context.Context, reply *models.CommentReply) error
	DeleteReply(ctx context.Context, replyID string) error

	// 评论点赞
	CreateCommentLike(ctx context.Context, like *models.CommentLike) error
	DeleteCommentLike(ctx context.Context, userID, commentID string) error
	GetCommentLike(ctx context.Context, userID, commentID string) (*models.CommentLike, error)
	GetCommentLikes(ctx context.Context, commentID string, offset, limit int) ([]*models.User, int64, error)

	// 分享
	CreatePostShare(ctx context.Context, share *models.PostShare) error
	GetPostShares(ctx context.Context, postID string, offset, limit int) ([]*models.PostShare, int64, error)
	GetUserShares(ctx context.Context, userID string, offset, limit int) ([]*models.PostShare, int64, error)

	// 举报
	CreateCommentReport(ctx context.Context, report *models.CommentReport) error
	GetCommentReportByID(ctx context.Context, reportID string) (*models.CommentReport, error)
	UpdateCommentReport(ctx context.Context, report *models.CommentReport) error
	GetCommentReports(ctx context.Context, status string, offset, limit int) ([]*models.CommentReport, int64, error)

	// 用户互动历史
	LogInteraction(ctx context.Context, history *models.InteractionHistory) error
	GetUserInteractionHistory(ctx context.Context, userID string, actionType string, offset, limit int) ([]*models.InteractionHistory, int64, error)

	// 统计
	GetPostInteractionCounts(ctx context.Context, postID string) (map[string]int64, error)
	GetUserInteractionCounts(ctx context.Context, userID string) (map[string]int64, error)

	// 获取数据库连接
	GetDB() *gorm.DB
}

// Repository 是交互服务的数据存储层，实现了InteractionRepository接口
type Repository struct {
	db     *gorm.DB
	logger zerolog.Logger
}

// NewRepository 创建一个新的Repository实例
func NewRepository(db *gorm.DB, logger zerolog.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

// GetDB 返回GORM数据库连接实例
func (r *Repository) GetDB() *gorm.DB {
	return r.db
}

// CreatePostLike 创建帖子点赞
func (r *Repository) CreatePostLike(ctx context.Context, like *models.PostLike) error {
	return r.db.WithContext(ctx).Create(like).Error
}

// DeletePostLike 删除帖子点赞
func (r *Repository) DeletePostLike(ctx context.Context, userID, postID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Delete(&models.PostLike{}).Error
}

// GetPostLike 获取帖子点赞
func (r *Repository) GetPostLike(ctx context.Context, userID, postID string) (*models.PostLike, error) {
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
func (r *Repository) GetPostLikes(ctx context.Context, postID string, offset, limit int) ([]*models.User, int64, error) {
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
func (r *Repository) GetUserLikedPosts(ctx context.Context, userID string, offset, limit int) ([]*models.Post, int64, error) {
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

// LogInteraction 记录用户互动历史
func (r *Repository) LogInteraction(ctx context.Context, history *models.InteractionHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

// CreateBookmark 创建书签
func (r *Repository) CreateBookmark(ctx context.Context, bookmark *models.Bookmark) error {
	return r.db.WithContext(ctx).Create(bookmark).Error
}

// DeleteBookmark 删除书签
func (r *Repository) DeleteBookmark(ctx context.Context, userID, postID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Delete(&models.Bookmark{}).Error
}

// GetBookmark 获取书签
func (r *Repository) GetBookmark(ctx context.Context, userID, postID string) (*models.Bookmark, error) {
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
func (r *Repository) GetUserBookmarks(ctx context.Context, userID string, collectionName string, offset, limit int) ([]*models.Post, int64, error) {
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
func (r *Repository) CreateBookmarkCollection(ctx context.Context, collection *models.BookmarkCollection) error {
	return r.db.WithContext(ctx).Create(collection).Error
}

// GetUserBookmarkCollections 获取用户的所有收藏夹
func (r *Repository) GetUserBookmarkCollections(ctx context.Context, userID string) ([]*models.BookmarkCollection, error) {
	var collections []*models.BookmarkCollection
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&collections).Error; err != nil {
		return nil, err
	}
	return collections, nil
}

// CreateComment 创建评论
func (r *Repository) CreateComment(ctx context.Context, comment *models.Comment) error {
	return nil
}

// GetCommentByID 获取评论
func (r *Repository) GetCommentByID(ctx context.Context, commentID string) (*models.Comment, error) {
	return nil, nil
}

// UpdateComment 更新评论
func (r *Repository) UpdateComment(ctx context.Context, comment *models.Comment) error {
	return nil
}

// DeleteComment 删除评论
func (r *Repository) DeleteComment(ctx context.Context, commentID string) error {
	return nil
}

// GetPostComments 获取帖子评论
func (r *Repository) GetPostComments(ctx context.Context, postID string, offset, limit int) ([]*models.Comment, int64, error) {
	return nil, 0, nil
}

// GetUserComments 获取用户评论
func (r *Repository) GetUserComments(ctx context.Context, userID string, offset, limit int) ([]*models.Comment, int64, error) {
	return nil, 0, nil
}

// CreateCommentReply 创建评论回复
func (r *Repository) CreateCommentReply(ctx context.Context, reply *models.CommentReply) error {
	return nil
}

// GetCommentReplies 获取评论回复
func (r *Repository) GetCommentReplies(ctx context.Context, commentID string, offset, limit int) ([]*models.CommentReply, int64, error) {
	return nil, 0, nil
}

// GetReplyByID 获取回复
func (r *Repository) GetReplyByID(ctx context.Context, replyID string) (*models.CommentReply, error) {
	return nil, nil
}

// UpdateReply 更新回复
func (r *Repository) UpdateReply(ctx context.Context, reply *models.CommentReply) error {
	return nil
}

// DeleteReply 删除回复
func (r *Repository) DeleteReply(ctx context.Context, replyID string) error {
	return nil
}

// CreateCommentLike 点赞评论
func (r *Repository) CreateCommentLike(ctx context.Context, like *models.CommentLike) error {
	return nil
}

// DeleteCommentLike 取消点赞评论
func (r *Repository) DeleteCommentLike(ctx context.Context, userID, commentID string) error {
	return nil
}

// GetCommentLike 获取评论点赞
func (r *Repository) GetCommentLike(ctx context.Context, userID, commentID string) (*models.CommentLike, error) {
	return nil, nil
}

// GetCommentLikes 获取评论点赞列表
func (r *Repository) GetCommentLikes(ctx context.Context, commentID string, offset, limit int) ([]*models.User, int64, error) {
	return nil, 0, nil
}

// CreatePostShare 创建分享
func (r *Repository) CreatePostShare(ctx context.Context, share *models.PostShare) error {
	return nil
}

// GetPostShares 获取帖子分享
func (r *Repository) GetPostShares(ctx context.Context, postID string, offset, limit int) ([]*models.PostShare, int64, error) {
	return nil, 0, nil
}

// GetUserShares 获取用户分享
func (r *Repository) GetUserShares(ctx context.Context, userID string, offset, limit int) ([]*models.PostShare, int64, error) {
	return nil, 0, nil
}

// CreateCommentReport 创建评论举报
func (r *Repository) CreateCommentReport(ctx context.Context, report *models.CommentReport) error {
	return nil
}

// GetCommentReportByID 获取评论举报
func (r *Repository) GetCommentReportByID(ctx context.Context, reportID string) (*models.CommentReport, error) {
	return nil, nil
}

// UpdateCommentReport 更新评论举报
func (r *Repository) UpdateCommentReport(ctx context.Context, report *models.CommentReport) error {
	return nil
}

// GetCommentReports 获取评论举报列表
func (r *Repository) GetCommentReports(ctx context.Context, status string, offset, limit int) ([]*models.CommentReport, int64, error) {
	return nil, 0, nil
}

// GetUserInteractionHistory 获取用户互动历史
func (r *Repository) GetUserInteractionHistory(ctx context.Context, userID string, actionType string, offset, limit int) ([]*models.InteractionHistory, int64, error) {
	return nil, 0, nil
}

// GetPostInteractionCounts 获取帖子互动计数
func (r *Repository) GetPostInteractionCounts(ctx context.Context, postID string) (map[string]int64, error) {
	return nil, nil
}

// GetUserInteractionCounts 获取用户互动计数
func (r *Repository) GetUserInteractionCounts(ctx context.Context, userID string) (map[string]int64, error) {
	return nil, nil
}
