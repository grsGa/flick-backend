package repository

import (
	"context"

	"github.com/flick/backend/services/user/proto"
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

	// GetFollowers 获取关注者
	GetFollowers(ctx context.Context, userID string, first int, after string) ([]*proto.User, *proto.PageInfo, error)

	// GetFollowing 获取正在关注
	GetFollowing(ctx context.Context, userID string, first int, after string) ([]*proto.User, *proto.PageInfo, error)

	// FollowUser 关注用户
	FollowUser(ctx context.Context, followerID, followingID string) error

	// UnfollowUser 取消关注用户
	UnfollowUser(ctx context.Context, followerID, followingID string) error

	// UpdateUserAvatar 更新用户头像版本和URL
	UpdateUserAvatar(ctx context.Context, userID, avatarURL string, avatarVersion int32) error
}
