package server

import (
	"context"
	"net"

	"github.com/flick/backend/services/auth/internal/service"
	"github.com/flick/backend/services/auth/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// grpcServer gRPC服务实现
type grpcServer struct {
	proto.UnimplementedAuthServiceServer
	authService service.AuthService
}

// NewGRPCServer 创建gRPC服务实例
func NewGRPCServer(authService service.AuthService) *grpcServer {
	return &grpcServer{
		authService: authService,
	}
}

// Login 实现用户登录接口
func (s *grpcServer) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	return s.authService.Login(ctx, req)
}

// Register 实现用户注册接口
func (s *grpcServer) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	return s.authService.Register(ctx, req)
}

// ValidateToken 实现验证令牌接口
func (s *grpcServer) ValidateToken(ctx context.Context, req *proto.ValidateTokenRequest) (*proto.ValidateTokenResponse, error) {
	return s.authService.ValidateToken(ctx, req)
}

// RefreshToken 实现刷新令牌接口
func (s *grpcServer) RefreshToken(ctx context.Context, req *proto.RefreshTokenRequest) (*proto.RefreshTokenResponse, error) {
	return s.authService.RefreshToken(ctx, req)
}

// Logout 实现登出接口
func (s *grpcServer) Logout(ctx context.Context, req *proto.LogoutRequest) (*proto.LogoutResponse, error) {
	return s.authService.Logout(ctx, req)
}

// GithubLogin 实现Github登录接口
func (s *grpcServer) GithubLogin(ctx context.Context, req *proto.GithubLoginRequest) (*proto.GithubLoginResponse, error) {
	return s.authService.GithubLogin(ctx, req)
}

// GithubCallback 实现Github回调接口
func (s *grpcServer) GithubCallback(ctx context.Context, req *proto.GithubCallbackRequest) (*proto.GithubCallbackResponse, error) {
	return s.authService.GithubCallback(ctx, req)
}

// GoogleLogin 实现Google登录接口
func (s *grpcServer) GoogleLogin(ctx context.Context, req *proto.GoogleLoginRequest) (*proto.GoogleLoginResponse, error) {
	return s.authService.GoogleLogin(ctx, req)
}

// GoogleCallback 实现Google回调接口
func (s *grpcServer) GoogleCallback(ctx context.Context, req *proto.GoogleCallbackRequest) (*proto.GoogleCallbackResponse, error) {
	return s.authService.GoogleCallback(ctx, req)
}

// Run 启动gRPC服务
func (s *grpcServer) Run(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	proto.RegisterAuthServiceServer(grpcServer, s)

	// Register health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	// Set serving status
	healthServer.SetServingStatus("auth-service", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING) // For overall server health

	return grpcServer.Serve(lis)
}
