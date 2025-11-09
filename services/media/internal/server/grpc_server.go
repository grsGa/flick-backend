package server

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/flick/backend/pkg/config"
	"github.com/flick/backend/pkg/discovery"
	"github.com/flick/backend/services/media/internal/middleware"
	"github.com/flick/backend/services/media/internal/service"
	"github.com/flick/backend/services/media/proto"
	user_proto "github.com/flick/backend/services/user/proto"
	"google.golang.org/grpc"
)

// grpcServer gRPC服务实现
type GRPCServer struct {
	proto.UnimplementedMediaServiceServer
	mediaService *service.MediaService
	config       *config.Config
	userClient   user_proto.UserServiceClient
}

// NewGRPCServer 创建gRPC服务实例
func NewGRPCServer(mediaService *service.MediaService, cfg *config.Config) *GRPCServer {
	// 创建user service客户端连接 - 重试机制
	var userConn *grpc.ClientConn
	var err error
	maxRetries := 5
	retryDelay := time.Second * 2

	for i := 0; i < maxRetries; i++ {
		userConn, err = discovery.GetServiceConnection("user-service")
		if err == nil {
			log.Printf("[MEDIA SERVER] Successfully connected to user-service on attempt %d", i+1)
			break
		}
		
		log.Printf("[MEDIA SERVER] Failed to connect to user-service (attempt %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			log.Printf("[MEDIA SERVER] Retrying in %v...", retryDelay)
			time.Sleep(retryDelay)
		}
	}

	if userConn == nil {
		log.Fatalf("[MEDIA SERVER] CRITICAL: Failed to connect to user-service after %d attempts. Media service cannot start without user validation capability.", maxRetries)
	}

	userClient := user_proto.NewUserServiceClient(userConn)

	return &GRPCServer{
		mediaService: mediaService,
		config:       cfg,
		userClient:   userClient,
	}
}

// UploadFile 实现上传文件接口
func (s *GRPCServer) UploadFile(ctx context.Context, req *proto.UploadFileRequest) (*proto.UploadFileResponse, error) {
	return s.mediaService.UploadFile(ctx, req)
}

// GetFile 实现获取文件信息接口
func (s *GRPCServer) GetFile(ctx context.Context, req *proto.GetFileRequest) (*proto.GetFileResponse, error) {
	return s.mediaService.GetFile(ctx, req)
}

// DeleteFile 实现删除文件接口
func (s *GRPCServer) DeleteFile(ctx context.Context, req *proto.DeleteFileRequest) (*proto.DeleteFileResponse, error) {
	return s.mediaService.DeleteFile(ctx, req)
}

// ListFiles 实现获取文件列表接口
func (s *GRPCServer) ListFiles(ctx context.Context, req *proto.ListFilesRequest) (*proto.ListFilesResponse, error) {
	return s.mediaService.ListFiles(ctx, req)
}

// Run 启动gRPC服务
func (s *GRPCServer) Run(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	// Set max message size to 100MB for large file uploads
	maxMsgSize := 100 * 1024 * 1024 // 100MB
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize),
		grpc.UnaryInterceptor(middleware.AuthInterceptor(s.config, s.userClient)), // Add JWT authentication with user validation
	)
	proto.RegisterMediaServiceServer(grpcServer, s)

	return grpcServer.Serve(lis)
}

