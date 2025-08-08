package server

import (
	"context"
	"net"
	
	"google.golang.org/grpc"
	"backend/services/content/internal/service"
	"backend/services/content/proto"
)

// grpcServer gRPC服务实现
type grpcServer struct {
	proto.UnimplementedContentServiceServer
	contentService service.ContentService
}

// NewGRPCServer 创建gRPC服务实例
func NewGRPCServer(contentService service.ContentService) *grpcServer {
	return &grpcServer{
		contentService: contentService,
	}
}

// GetContent 实现获取内容接口
func (s *grpcServer) GetContent(ctx context.Context, req *proto.GetContentRequest) (*proto.GetContentResponse, error) {
	return s.contentService.GetContent(ctx, req)
}

// CreateContent 实现创建内容接口
func (s *grpcServer) CreateContent(ctx context.Context, req *proto.CreateContentRequest) (*proto.CreateContentResponse, error) {
	return s.contentService.CreateContent(ctx, req)
}

// UpdateContent 实现更新内容接口
func (s *grpcServer) UpdateContent(ctx context.Context, req *proto.UpdateContentRequest) (*proto.UpdateContentResponse, error) {
	return s.contentService.UpdateContent(ctx, req)
}

// DeleteContent 实现删除内容接口
func (s *grpcServer) DeleteContent(ctx context.Context, req *proto.DeleteContentRequest) (*proto.DeleteContentResponse, error) {
	return s.contentService.DeleteContent(ctx, req)
}

// ListContent 实现列出内容接口
func (s *grpcServer) ListContent(ctx context.Context, req *proto.ListContentRequest) (*proto.ListContentResponse, error) {
	return s.contentService.ListContent(ctx, req)
}

// Run 启动gRPC服务
func (s *grpcServer) Run(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	
	grpcServer := grpc.NewServer()
	proto.RegisterContentServiceServer(grpcServer, s)
	
	return grpcServer.Serve(lis)
}