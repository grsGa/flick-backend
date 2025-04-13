package repository

import (
	"context"

	"backend/pkg/models"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// Repository 是交互服务的数据存储层
type Repository struct {
	*PostgresRepository
	logger zerolog.Logger
}

// NewRepository 创建一个新的Repository实例
func NewRepository(db *gorm.DB, logger zerolog.Logger) *Repository {
	// 创建PostgreSQL仓库实现
	postgresRepo := NewPostgresRepository(db)
	
	return &Repository{
		PostgresRepository: postgresRepo,
		logger: logger,
	}
}

// InteractionRepository 定义互动仓库接口
type InteractionRepository interface {
	// 点赞相关操作
	CreatePostLike(ctx context.Context, like *models.PostLike) error
	DeletePostLike(ctx context.Context, userID, postID string) error
	GetPostLike(ctx context.Context, userID, postID string) (*models.PostLike, error)
	GetPostLikes(ctx context.Context, postID string, offset, limit int) ([]*models.User, int64, error)
	GetUserLikedPosts(ctx context.Context, userID string, offset, limit int) ([]*models.Post, int64, error)
	
	// 评论相关操作
	CreateComment(ctx context.Context, comment *models.Comment) error
	GetCommentByID(ctx context.Context, commentID string) (*models.Comment, error)
	UpdateComment(ctx context.Context, comment *models.Comment) error
	DeleteComment(ctx context.Context, commentID string) error
	GetPostComments(ctx context.Context, postID string, offset, limit int) ([]*models.Comment, int64, error)
	GetUserComments(ctx context.Context, userID string, offset, limit int) ([]*models.Comment, int64, error)
	
	// 评论回复相关操作
	CreateCommentReply(ctx context.Context, reply *models.CommentReply) error
	GetCommentReplies(ctx context.Context, commentID string, offset, limit int) ([]*models.CommentReply, int64, error)
	GetReplyByID(ctx context.Context, replyID string) (*models.CommentReply, error)
	UpdateReply(ctx context.Context, reply *models.CommentReply) error
	DeleteReply(ctx context.Context, replyID string) error
	
	// 评论点赞相关操作
	CreateCommentLike(ctx context.Context, like *models.CommentLike) error
	DeleteCommentLike(ctx context.Context, userID, commentID string) error
	GetCommentLike(ctx context.Context, userID, commentID string) (*models.CommentLike, error)
	GetCommentLikes(ctx context.Context, commentID string, offset, limit int) ([]*models.User, int64, error)
	
	// 分享相关操作
	CreatePostShare(ctx context.Context, share *models.PostShare) error
	GetPostShares(ctx context.Context, postID string, offset, limit int) ([]*models.PostShare, int64, error)
	GetUserShares(ctx context.Context, userID string, offset, limit int) ([]*models.PostShare, int64, error)
	
	// 收藏/书签相关操作
	CreateBookmark(ctx context.Context, bookmark *models.Bookmark) error
	DeleteBookmark(ctx context.Context, userID, postID string) error
	GetBookmark(ctx context.Context, userID, postID string) (*models.Bookmark, error)
	GetUserBookmarks(ctx context.Context, userID string, collectionName string, offset, limit int) ([]*models.Post, int64, error)
	CreateBookmarkCollection(ctx context.Context, collection *models.BookmarkCollection) error
	GetUserBookmarkCollections(ctx context.Context, userID string) ([]*models.BookmarkCollection, error)
	
	// 评论举报相关操作
	CreateCommentReport(ctx context.Context, report *models.CommentReport) error
	GetCommentReportByID(ctx context.Context, reportID string) (*models.CommentReport, error)
	UpdateCommentReport(ctx context.Context, report *models.CommentReport) error
	GetCommentReports(ctx context.Context, status string, offset, limit int) ([]*models.CommentReport, int64, error)
	
	// 交互历史记录
	LogInteraction(ctx context.Context, history *models.InteractionHistory) error
	GetUserInteractionHistory(ctx context.Context, userID string, actionType string, offset, limit int) ([]*models.InteractionHistory, int64, error)
	
	// 统计查询
	GetPostInteractionCounts(ctx context.Context, postID string) (map[string]int64, error)
	GetUserInteractionCounts(ctx context.Context, userID string) (map[string]int64, error)
} 