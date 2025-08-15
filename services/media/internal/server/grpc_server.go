package server

import (
	"context"
	"net"
	
	"google.golang.org/grpc"
	"github.com/flick/backend/services/media/internal/service"
	"github.com/flick/backend/services/media/proto"
)

// grpcServer gRPC服务实现
type grpcServer struct {
	proto.UnimplementedMediaServiceServer
	mediaService service.MediaService
}

// NewGRPCServer 创建gRPC服务实例
func NewGRPCServer(mediaService service.MediaService) *grpcServer {
	return &grpcServer{
		mediaService: mediaService,
	}
}

// UploadFile 实现上传文件接口
func (s *grpcServer) UploadFile(ctx context.Context, req *proto.UploadFileRequest) (*proto.UploadFileResponse, error) {
	return s.mediaService.UploadFile(ctx, req)
}

// GetFile 实现获取文件信息接口
func (s *grpcServer) GetFile(ctx context.Context, req *proto.GetFileRequest) (*proto.GetFileResponse, error) {
	return s.mediaService.GetFile(ctx, req)
}

// DeleteFile 实现删除文件接口
func (s *grpcServer) DeleteFile(ctx context.Context, req *proto.DeleteFileRequest) (*proto.DeleteFileResponse, error) {
	return s.mediaService.DeleteFile(ctx, req)
}

// ListFiles 实现获取文件列表接口
func (s *grpcServer) ListFiles(ctx context.Context, req *proto.ListFilesRequest) (*proto.ListFilesResponse, error) {
	return s.mediaService.ListFiles(ctx, req)
}

// Run 启动gRPC服务
func (s *grpcServer) Run(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	
	grpcServer := grpc.NewServer()
	proto.RegisterMediaServiceServer(grpcServer, s)
	
	return grpcServer.Serve(lis)
}
