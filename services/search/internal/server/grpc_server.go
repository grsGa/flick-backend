package server

import (
	"context"
	"net"
	
	"google.golang.org/grpc"
	"backend/services/search/internal/service"
	"backend/services/search/proto"
)

// grpcServer gRPC服务实现
type grpcServer struct {
	proto.UnimplementedSearchServiceServer
	searchService service.SearchService
}

// NewGRPCServer 创建gRPC服务实例
func NewGRPCServer(searchService service.SearchService) *grpcServer {
	return &grpcServer{
		searchService: searchService,
	}
}

// SearchContent 实现搜索内容接口
func (s *grpcServer) SearchContent(ctx context.Context, req *proto.SearchContentRequest) (*proto.SearchContentResponse, error) {
	return s.searchService.SearchContent(ctx, req)
}

// SearchUsers 实现搜索用户接口
func (s *grpcServer) SearchUsers(ctx context.Context, req *proto.SearchUsersRequest) (*proto.SearchUsersResponse, error) {
	return s.searchService.SearchUsers(ctx, req)
}

// SearchHashtags 实现搜索标签接口
func (s *grpcServer) SearchHashtags(ctx context.Context, req *proto.SearchHashtagsRequest) (*proto.SearchHashtagsResponse, error) {
	return s.searchService.SearchHashtags(ctx, req)
}

// Run 启动gRPC服务
func (s *grpcServer) Run(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	
	grpcServer := grpc.NewServer()
	proto.RegisterSearchServiceServer(grpcServer, s)
	
	return grpcServer.Serve(lis)
}