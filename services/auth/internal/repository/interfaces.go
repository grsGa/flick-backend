package repository

import (
	"backend/services/auth/proto"
	"context"
)

// AuthRepository 定义认证仓储接口
type AuthRepository interface {
	// GetUserByIdentifier 根据标识符（用户名、邮箱或手机号）获取用户
	GetUserByIdentifier(ctx context.Context, identifier string) (*proto.User, error)

	// GetUserByID 根据ID获取用户
	GetUserByID(ctx context.Context, id string) (*proto.User, error)

	// VerifyPassword 验证密码
	VerifyPassword(ctx context.Context, userID string, password string) error

	// CreateUser 创建用户
	CreateUser(ctx context.Context, user *proto.User, password string) (*proto.User, error)

	// UpdateUserLoginInfo 更新用户登录信息
	UpdateUserLoginInfo(ctx context.Context, userID string, lastLoginAt string) error

	// CreateSession 创建会话
	CreateSession(ctx context.Context, session *Session) error

	// GetSession 获取会话
	GetSession(ctx context.Context, refreshToken string) (*Session, error)

	// DeleteSession 删除会话
	DeleteSession(ctx context.Context, refreshToken string) error

	// DeleteUserSessions 删除用户所有会话
	DeleteUserSessions(ctx context.Context, userID string) error
}

// Session 会话模型
type Session struct {
	ID           string
	UserID       string
	RefreshToken string
	ExpiresAt    string
	CreatedAt    string
}
