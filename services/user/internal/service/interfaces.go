package service

import (
	"backend/services/user/proto"
	"context"
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
}
