package server

import (
	"context"
	"net"
	
	"google.golang.org/grpc"
	"backend/services/bookmark/internal/service"
	"backend/services/bookmark/proto"
)

// grpcServer gRPC服务实现
type grpcServer struct {
	proto.UnimplementedBookmarkServiceServer
	bookmarkService service.BookmarkService
}

// NewGRPCServer 创建gRPC服务实例
func NewGRPCServer(bookmarkService service.BookmarkService) *grpcServer {
	return &grpcServer{
		bookmarkService: bookmarkService,
	}
}

// CreateBookmark 实现创建书签接口
func (s *grpcServer) CreateBookmark(ctx context.Context, req *proto.CreateBookmarkRequest) (*proto.CreateBookmarkResponse, error) {
	return s.bookmarkService.CreateBookmark(ctx, req)
}

// DeleteBookmark 实现删除书签接口
func (s *grpcServer) DeleteBookmark(ctx context.Context, req *proto.DeleteBookmarkRequest) (*proto.DeleteBookmarkResponse, error) {
	return s.bookmarkService.DeleteBookmark(ctx, req)
}

// IsBookmarked 实现检查是否已收藏接口
func (s *grpcServer) IsBookmarked(ctx context.Context, req *proto.IsBookmarkedRequest) (*proto.IsBookmarkedResponse, error) {
	return s.bookmarkService.IsBookmarked(ctx, req)
}

// ListBookmarks 实现获取用户书签列表接口
func (s *grpcServer) ListBookmarks(ctx context.Context, req *proto.ListBookmarksRequest) (*proto.ListBookmarksResponse, error) {
	return s.bookmarkService.ListBookmarks(ctx, req)
}

// Run 启动gRPC服务
func (s *grpcServer) Run(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	
	grpcServer := grpc.NewServer()
	proto.RegisterBookmarkServiceServer(grpcServer, s)
	
	return grpcServer.Serve(lis)
}