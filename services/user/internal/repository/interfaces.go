package repository

import (
	"backend/services/user/proto"
	"context"
)

// UserRepository 定义用户仓储接口
type UserRepository interface {
	// CreateUser 创建用户
	CreateUser(ctx context.Context, user *proto.User) error

	// GetUserByID 根据ID获取用户
	GetUserByID(ctx context.Context, id string) (*proto.User, error)

	// UpdateUser 更新用户
	UpdateUser(ctx context.Context, user *proto.User) error

	// DeleteUser 删除用户
	DeleteUser(ctx context.Context, id string) error

	// GetUserByUsername 根据用户名获取用户
	GetUserByUsername(ctx context.Context, username string) (*proto.User, error)

	// GetUserByEmail 根据邮箱获取用户
	GetUserByEmail(ctx context.Context, email string) (*proto.User, error)

	// Authenticate 用户认证
	Authenticate(ctx context.Context, identifier, password string) (*proto.User, error)
}
