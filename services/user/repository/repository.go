package repository

import (
	"context"

	"backend/pkg/models"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// UserRepository 定义用户仓库接口
type UserRepository interface {
	// 用户CRUD操作
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, offset, limit int) ([]*models.User, int64, error)
	SearchUsers(ctx context.Context, query string, offset, limit int) ([]*models.User, int64, error)

	// 角色相关操作
	GetUserRoles(ctx context.Context, userID string) ([]string, error)
	AssignRoleToUser(ctx context.Context, userID, role string) error
	RemoveRoleFromUser(ctx context.Context, userID, role string) error
	
	// 关系相关操作
	FollowUser(ctx context.Context, followerID, followingID string) error
	UnfollowUser(ctx context.Context, followerID, followingID string) error
	GetFollowers(ctx context.Context, userID string, offset, limit int) ([]*models.User, int64, error)
	GetFollowing(ctx context.Context, userID string, offset, limit int) ([]*models.User, int64, error)
	IsFollowing(ctx context.Context, followerID, followingID string) (bool, error)
	GetFollowStats(ctx context.Context, userID string) (followers int64, following int64, err error)
	
	// 屏蔽相关操作
	BlockUser(ctx context.Context, blockerID, blockedID string, reason string) error
	UnblockUser(ctx context.Context, blockerID, blockedID string) error
	GetBlockedUsers(ctx context.Context, userID string, offset, limit int) ([]*models.User, int64, error)
	IsBlocked(ctx context.Context, blockerID, blockedID string) (bool, error)
	
	// 认证相关操作
	CreateVerification(ctx context.Context, verification *models.UserVerification) error
	GetVerification(ctx context.Context, userID, verificationType string) (*models.UserVerification, error)
	VerifyToken(ctx context.Context, userID, verificationType, token string) (bool, error)
	MarkVerificationUsed(ctx context.Context, id string) error
	
	// 会话相关操作
	CreateSession(ctx context.Context, session *models.UserSession) error
	GetSession(ctx context.Context, id string) (*models.UserSession, error)
	DeleteSession(ctx context.Context, id string) error
	DeleteUserSessions(ctx context.Context, userID string) error
	
	// 活动日志相关操作
	LogUserActivity(ctx context.Context, activity *models.UserActivity) error
	GetUserActivities(ctx context.Context, userID string, offset, limit int) ([]*models.UserActivity, int64, error)
}

// Repository 用户数据存储层
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