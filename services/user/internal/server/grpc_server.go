package server

import (
	"context"
	"net"

	"github.com/flick/backend/services/user/internal/service"
	"github.com/flick/backend/services/user/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// grpcServer gRPC服务实现
type grpcServer struct {
	proto.UnimplementedUserServiceServer
	userService service.UserService
}

// NewGRPCServer 创建gRPC服务实例
func NewGRPCServer(userService service.UserService) *grpcServer {
	return &grpcServer{
		userService: userService,
	}
}

// GetUser 实现获取用户接口
func (s *grpcServer) GetUser(ctx context.Context, req *proto.GetUserRequest) (*proto.GetUserResponse, error) {
	return s.userService.GetUser(ctx, req)
}

// GetUserByUsername 实现根据用户名获取用户接口
func (s *grpcServer) GetUserByUsername(ctx context.Context, req *proto.GetUserByUsernameRequest) (*proto.GetUserResponse, error) {
	return s.userService.GetUserByUsername(ctx, req)
}

// UpdateUser 实现更新用户接口
func (s *grpcServer) UpdateUser(ctx context.Context, req *proto.UpdateUserRequest) (*proto.UpdateUserResponse, error) {
	return s.userService.UpdateUser(ctx, req)
}

// DeleteUser 实现删除用户接口
func (s *grpcServer) DeleteUser(ctx context.Context, req *proto.DeleteUserRequest) (*proto.DeleteUserResponse, error) {
	return s.userService.DeleteUser(ctx, req)
}

// Register 实现用户注册接口
func (s *grpcServer) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	return s.userService.Register(ctx, req)
}

// Login 实现用户登录接口
func (s *grpcServer) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	return s.userService.Login(ctx, req)
}

// GetFollowers 实现获取关注者接口
func (s *grpcServer) GetFollowers(ctx context.Context, req *proto.GetFollowersRequest) (*proto.GetFollowersResponse, error) {
	return s.userService.GetFollowers(ctx, req)
}

// GetFollowing 实现获取正在关注接口
func (s *grpcServer) GetFollowing(ctx context.Context, req *proto.GetFollowingRequest) (*proto.GetFollowingResponse, error) {
	return s.userService.GetFollowing(ctx, req)
}

// UpdateProfile 实现更新个人资料接口
func (s *grpcServer) UpdateProfile(ctx context.Context, req *proto.UpdateProfileRequest) (*proto.UpdateProfileResponse, error) {
	return s.userService.UpdateProfile(ctx, req)
}

// FollowUser 实现关注用户接口
func (s *grpcServer) FollowUser(ctx context.Context, req *proto.FollowUserRequest) (*proto.FollowUserResponse, error) {
	return s.userService.FollowUser(ctx, req)
}

// UnfollowUser 实现取消关注用户接口
func (s *grpcServer) UnfollowUser(ctx context.Context, req *proto.UnfollowUserRequest) (*proto.UnfollowUserResponse, error) {
	return s.userService.UnfollowUser(ctx, req)
}

// UpdateUserAvatar 实现更新用户头像接口
func (s *grpcServer) UpdateUserAvatar(ctx context.Context, req *proto.UpdateUserAvatarRequest) (*proto.UpdateUserAvatarResponse, error) {
	return s.userService.UpdateUserAvatar(ctx, req)
}

// Run 启动gRPC服务
func (s *grpcServer) Run(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	proto.RegisterUserServiceServer(grpcServer, s)

	// Register health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	// Set serving status
	healthServer.SetServingStatus("user-service", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING) // For overall server health

	return grpcServer.Serve(lis)
}
