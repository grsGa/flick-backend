package service

import (
	"context"

	"github.com/flick/backend/services/user/proto"
)

// UserService 定义用户服务接口
type UserService interface {
	// Register 用户注册
	Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error)

	// Login 用户登录
	Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error)

	// GetUser 获取用户信息
	GetUser(ctx context.Context, req *proto.GetUserRequest) (*proto.GetUserResponse, error)

	// GetUserByUsername 根据用户名获取用户信息
	GetUserByUsername(ctx context.Context, req *proto.GetUserByUsernameRequest) (*proto.GetUserResponse, error)

	// UpdateUser 更新用户
	UpdateUser(ctx context.Context, req *proto.UpdateUserRequest) (*proto.UpdateUserResponse, error)

	// DeleteUser 删除用户
	DeleteUser(ctx context.Context, req *proto.DeleteUserRequest) (*proto.DeleteUserResponse, error)

	// GetFollowers 获取关注者
	GetFollowers(ctx context.Context, req *proto.GetFollowersRequest) (*proto.GetFollowersResponse, error)

	// GetFollowing 获取正在关注
	GetFollowing(ctx context.Context, req *proto.GetFollowingRequest) (*proto.GetFollowingResponse, error)

	// UpdateProfile 更新个人资料
	UpdateProfile(ctx context.Context, req *proto.UpdateProfileRequest) (*proto.UpdateProfileResponse, error)

	// FollowUser 关注用户
	FollowUser(ctx context.Context, req *proto.FollowUserRequest) (*proto.FollowUserResponse, error)

	// UnfollowUser 取消关注用户
	UnfollowUser(ctx context.Context, req *proto.UnfollowUserRequest) (*proto.UnfollowUserResponse, error)

	// UpdateUserAvatar 更新用户头像版本和URL
	UpdateUserAvatar(ctx context.Context, req *proto.UpdateUserAvatarRequest) (*proto.UpdateUserAvatarResponse, error)

	// UpdateUserBanner 更新用户横幅版本和URL
	UpdateUserBanner(ctx context.Context, req *proto.UpdateUserBannerRequest) (*proto.UpdateUserBannerResponse, error)
}
