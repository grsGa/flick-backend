package repository

import (
	"context"
	"time"

	"backend/pkg/models"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// Repository 是内容服务的数据存储层
type Repository struct {
	*PostgresRepository
	logger zerolog.Logger
}

// NewRepository 创建一个新的Repository实例
func NewRepository(db *gorm.DB, logger zerolog.Logger) *Repository {
	return &Repository{
		PostgresRepository: NewPostgresRepository(db),
		logger: logger,
	}
}

// ContentRepository 定义内容仓库接口
type ContentRepository interface {
	// 帖子CRUD操作
	CreatePost(ctx context.Context, post *models.Post) error
	GetPostByID(ctx context.Context, id string) (*models.Post, error)
	UpdatePost(ctx context.Context, post *models.Post) error
	DeletePost(ctx context.Context, id string) error
	ListPosts(ctx context.Context, offset, limit int) ([]*models.Post, int64, error)
	SearchPosts(ctx context.Context, query string, offset, limit int) ([]*models.Post, int64, error)
	GetPostsByUser(ctx context.Context, userID string, offset, limit int) ([]*models.Post, int64, error)
	GetFeaturedPosts(ctx context.Context, offset, limit int) ([]*models.Post, int64, error)
	GetPostsByCategory(ctx context.Context, categoryID string, offset, limit int) ([]*models.Post, int64, error)
	GetPostsByTag(ctx context.Context, tagID string, offset, limit int) ([]*models.Post, int64, error)
	IncrementViewCount(ctx context.Context, postID string) error
	
	// 帖子审计日志
	LogPostActivity(ctx context.Context, log *models.PostAuditLog) error
	GetPostAuditLogs(ctx context.Context, postID string, offset, limit int) ([]*models.PostAuditLog, int64, error)
	
	// 标签相关操作
	CreateTag(ctx context.Context, tag *models.Tag) error
	GetTagByID(ctx context.Context, id string) (*models.Tag, error)
	GetTagBySlug(ctx context.Context, slug string) (*models.Tag, error)
	UpdateTag(ctx context.Context, tag *models.Tag) error
	DeleteTag(ctx context.Context, id string) error
	ListTags(ctx context.Context, offset, limit int) ([]*models.Tag, int64, error)
	GetPostTags(ctx context.Context, postID string) ([]*models.Tag, error)
	AddTagToPost(ctx context.Context, postID, tagID string) error
	RemoveTagFromPost(ctx context.Context, postID, tagID string) error
	
	// 分类相关操作
	CreateCategory(ctx context.Context, category *models.Category) error
	GetCategoryByID(ctx context.Context, id string) (*models.Category, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*models.Category, error)
	UpdateCategory(ctx context.Context, category *models.Category) error
	DeleteCategory(ctx context.Context, id string) error
	ListCategories(ctx context.Context, offset, limit int) ([]*models.Category, int64, error)
	GetCategoryHierarchy(ctx context.Context) ([]*models.Category, error)
	GetPostCategories(ctx context.Context, postID string) ([]*models.Category, error)
	AddCategoryToPost(ctx context.Context, postID, categoryID string) error
	RemoveCategoryFromPost(ctx context.Context, postID, categoryID string) error
	
	// 帖子保存相关操作
	SavePost(ctx context.Context, savedPost *models.SavedPost) error
	UnsavePost(ctx context.Context, userID, postID string) error
	GetSavedPosts(ctx context.Context, userID string, collection string, offset, limit int) ([]*models.Post, int64, error)
	IsSaved(ctx context.Context, userID, postID string) (bool, error)
	
	// 帖子媒体相关操作
	UpdatePostMedia(ctx context.Context, postID string, media models.MediaFiles) error
	
	// 帖子举报相关操作
	CreatePostReport(ctx context.Context, report *models.PostReport) error
	GetPostReportByID(ctx context.Context, id string) (*models.PostReport, error)
	UpdatePostReport(ctx context.Context, report *models.PostReport) error
	GetPostReports(ctx context.Context, status string, offset, limit int) ([]*models.PostReport, int64, error)
	GetUserPostReports(ctx context.Context, userID string, offset, limit int) ([]*models.PostReport, int64, error)
	
	// 投票相关操作
	UpdatePostPoll(ctx context.Context, postID string, options models.PollOptions, endsAt time.Time) error
	VoteInPoll(ctx context.Context, vote *models.PollVote) error
	GetPollVotes(ctx context.Context, postID string) (*models.PollOptions, int64, error)
	HasVoted(ctx context.Context, userID, postID string) (bool, string, error)
	
	// 统计相关操作
	GetPostStats(ctx context.Context, postID string) (map[string]int64, error)
	GetTopPosts(ctx context.Context, period string, limit int) ([]*models.Post, error)
	GetUserContentStats(ctx context.Context, userID string) (map[string]int64, error)
} 