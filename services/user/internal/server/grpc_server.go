package server

import (
	"context"
	"net"

	"backend/services/user/internal/service"
	"backend/services/user/proto"

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
