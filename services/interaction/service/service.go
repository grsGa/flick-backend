package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog"

	"backend/pkg/models"
	"backend/services/interaction/repository"
)

// InteractionServiceInterface 定义互动服务接口
type InteractionServiceInterface interface {
	// 点赞相关操作
	LikePost(ctx context.Context, userID, postID string) error
	UnlikePost(ctx context.Context, userID, postID string) error
	IsPostLiked(ctx context.Context, userID, postID string) (bool, error)
	GetPostLikes(ctx context.Context, postID string, page, pageSize int) ([]*models.User, int64, error)
	GetUserLikedPosts(ctx context.Context, userID string, page, pageSize int) ([]*models.Post, int64, error)

	// 评论相关操作
	CreateComment(ctx context.Context, comment *models.Comment) error
	GetCommentByID(ctx context.Context, commentID string) (*models.Comment, error)
	UpdateComment(ctx context.Context, comment *models.Comment) error
	DeleteComment(ctx context.Context, commentID string) error
	GetPostComments(ctx context.Context, postID string, page, pageSize int) ([]*models.Comment, int64, error)
	GetUserComments(ctx context.Context, userID string, page, pageSize int) ([]*models.Comment, int64, error)

	// 评论回复相关操作
	ReplyToComment(ctx context.Context, reply *models.CommentReply) error
	GetCommentReplies(ctx context.Context, commentID string, page, pageSize int) ([]*models.CommentReply, int64, error)
	UpdateReply(ctx context.Context, reply *models.CommentReply) error
	DeleteReply(ctx context.Context, replyID string) error

	// 评论点赞相关操作
	LikeComment(ctx context.Context, userID, commentID string) error
	UnlikeComment(ctx context.Context, userID, commentID string) error
	IsCommentLiked(ctx context.Context, userID, commentID string) (bool, error)
	GetCommentLikes(ctx context.Context, commentID string, page, pageSize int) ([]*models.User, int64, error)

	// 分享相关操作
	SharePost(ctx context.Context, share *models.PostShare) error
	GetPostShares(ctx context.Context, postID string, page, pageSize int) ([]*models.PostShare, int64, error)
	GetUserShares(ctx context.Context, userID string, page, pageSize int) ([]*models.PostShare, int64, error)

	// 收藏/书签相关操作
	BookmarkPost(ctx context.Context, userID, postID, collectionName string) error
	UnbookmarkPost(ctx context.Context, userID, postID string) error
	IsPostBookmarked(ctx context.Context, userID, postID string) (bool, error)
	GetUserBookmarks(ctx context.Context, userID string, collectionName string, page, pageSize int) ([]*models.Post, int64, error)
	CreateBookmarkCollection(ctx context.Context, collection *models.BookmarkCollection) error
	GetUserBookmarkCollections(ctx context.Context, userID string) ([]*models.BookmarkCollection, error)

	// 评论举报相关操作
	ReportComment(ctx context.Context, report *models.CommentReport) error
	ReviewCommentReport(ctx context.Context, reportID string, reviewerID string, approved bool, note string) error
	GetCommentReports(ctx context.Context, status string, page, pageSize int) ([]*models.CommentReport, int64, error)

	// 统计相关操作
	GetInteractionStats(ctx context.Context, postID string) (*models.InteractionStats, error)
	GetUserInteractionStats(ctx context.Context, userID string) (*models.UserInteractionStats, error)
	GetUserInteractionHistory(ctx context.Context, userID string, actionType string, page, pageSize int) ([]*models.InteractionHistory, int64, error)
}

// CommentCreateRequest 创建评论请求
type CommentCreateRequest struct {
	PostID   string `json:"post_id" binding:"required"`
	UserID   string `json:"user_id" binding:"required"`
	Content  string `json:"content" binding:"required"`
	ParentID string `json:"parent_id"`
}

// CommentUpdateRequest 更新评论请求
type CommentUpdateRequest struct {
	CommentID string `json:"comment_id" binding:"required"`
	UserID    string `json:"user_id" binding:"required"`
	Content   string `json:"content" binding:"required"`
}

// CommentReplyRequest 评论回复请求
type CommentReplyRequest struct {
	CommentID     string `json:"comment_id" binding:"required"`
	UserID        string `json:"user_id" binding:"required"`
	Content       string `json:"content" binding:"required"`
	MentionUserID string `json:"mention_user_id"`
}

// CommentReportRequest 评论举报请求
type CommentReportRequest struct {
	CommentID   string `json:"comment_id" binding:"required"`
	UserID      string `json:"user_id" binding:"required"`
	ReasonCode  string `json:"reason_code" binding:"required"`
	Description string `json:"description"`
}

// SharePostRequest 分享帖子请求
type SharePostRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	PostID   string `json:"post_id" binding:"required"`
	Platform string `json:"platform" binding:"required"`
	Content  string `json:"content"`
}

// InteractionService 实现交互服务逻辑
type InteractionService struct {
	repo        *repository.Repository
	redisClient *redis.Client
	logger      zerolog.Logger
}

// NewInteractionService 创建一个新的InteractionService实例
func NewInteractionService(repo *repository.Repository, redisClient *redis.Client, logger zerolog.Logger) *InteractionService {
	return &InteractionService{
		repo:        repo,
		redisClient: redisClient,
		logger:      logger,
	}
}

// CreateInteractionHandler 处理创建交互请求
func (s *InteractionService) CreateInteractionHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not implemented"))
}

// ListInteractionsHandler 处理列出交互请求
func (s *InteractionService) ListInteractionsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not implemented"))
}

// GetInteractionHandler 处理获取交互请求
func (s *InteractionService) GetInteractionHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not implemented"))
}

// UpdateInteractionHandler 处理更新交互请求
func (s *InteractionService) UpdateInteractionHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not implemented"))
}

// DeleteInteractionHandler 处理删除交互请求
func (s *InteractionService) DeleteInteractionHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not implemented"))
}

// 检查用户是否存在
func (s *InteractionService) checkUserExists(ctx context.Context, userID string) (bool, error) {
	db := s.repo.GetDB()
	if db == nil {
		return false, errors.New("无法获取数据库连接")
	}

	var count int64
	if err := db.WithContext(ctx).Table("flick_users").
		Where("id = ?", userID).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// 检查帖子是否存在
func (s *InteractionService) checkPostExists(ctx context.Context, postID string) (bool, error) {
	db := s.repo.GetDB()
	if db == nil {
		return false, errors.New("无法获取数据库连接")
	}

	var count int64
	if err := db.WithContext(ctx).Table("flick_posts").
		Where("id = ?", postID).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}
