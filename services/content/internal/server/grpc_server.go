package server

import (
	"context"
	"net"

	"github.com/flick/backend/pkg/discovery"
	"github.com/flick/backend/services/content/internal/service"
	"github.com/flick/backend/services/content/proto"
	"google.golang.org/grpc"
)

// grpcServer gRPC服务实现
type GRPCServer struct {
	proto.UnimplementedContentServiceServer
	postService *service.PostService
}

// NewGRPCServer 创建gRPC服务实例
func NewGRPCServer(postService *service.PostService) *GRPCServer {
	return &GRPCServer{
		postService: postService,
	}
}

// GetContent 实现获取内容接口（映射到PostService）
func (s *GRPCServer) GetContent(ctx context.Context, req *proto.GetPostRequest) (*proto.GetPostResponse, error) {
	return s.postService.GetPost(ctx, req)
}

// CreateContent 实现创建内容接口（映射到PostService）
func (s *GRPCServer) CreateContent(ctx context.Context, req *proto.CreatePostRequest) (*proto.CreatePostResponse, error) {
	return s.postService.CreatePost(ctx, req)
}

// UpdateContent 实现更新内容接口（暂不支持）
func (s *GRPCServer) UpdateContent(ctx context.Context, req *proto.CreatePostRequest) (*proto.CreatePostResponse, error) {
	return &proto.CreatePostResponse{
		Error: &proto.Error{
			Code:    501,
			Message: "Update content not implemented",
		},
	}, nil
}

// DeleteContent 实现删除内容接口（映射到PostService）
func (s *GRPCServer) DeleteContent(ctx context.Context, req *proto.DeletePostRequest) (*proto.DeletePostResponse, error) {
	return s.postService.DeletePost(ctx, req)
}

// ListContent 实现列出内容接口（映射到PostService）
func (s *GRPCServer) ListContent(ctx context.Context, req *proto.GetUserPostsRequest) (*proto.GetUserPostsResponse, error) {
	return s.postService.GetUserPosts(ctx, req)
}

// CreatePost 实现创建帖子接口
func (s *GRPCServer) CreatePost(ctx context.Context, req *proto.CreatePostRequest) (*proto.CreatePostResponse, error) {
	return s.postService.CreatePost(ctx, req)
}

// GetPost 实现获取帖子接口
func (s *GRPCServer) GetPost(ctx context.Context, req *proto.GetPostRequest) (*proto.GetPostResponse, error) {
	return s.postService.GetPost(ctx, req)
}

// GetUserPosts 实现获取用户帖子列表接口
func (s *GRPCServer) GetUserPosts(ctx context.Context, req *proto.GetUserPostsRequest) (*proto.GetUserPostsResponse, error) {
	return s.postService.GetUserPosts(ctx, req)
}

// GetTimeline 实现获取时间线接口
func (s *GRPCServer) GetTimeline(ctx context.Context, req *proto.GetTimelineRequest) (*proto.GetTimelineResponse, error) {
	return s.postService.GetTimeline(ctx, req)
}

// GetFollowingTimeline 实现获取关注用户时间线接口
func (s *GRPCServer) GetFollowingTimeline(ctx context.Context, req *proto.GetTimelineRequest) (*proto.GetTimelineResponse, error) {
	return s.postService.GetFollowingTimeline(ctx, req)
}

// DeletePost 实现删除帖子接口
func (s *GRPCServer) DeletePost(ctx context.Context, req *proto.DeletePostRequest) (*proto.DeletePostResponse, error) {
	return s.postService.DeletePost(ctx, req)
}

// CheckReplyPermission 实现检查回复权限接口
func (s *GRPCServer) CheckReplyPermission(ctx context.Context, req *proto.CheckReplyPermissionRequest) (*proto.CheckReplyPermissionResponse, error) {
	return s.postService.CheckReplyPermission(ctx, req)
}

// GetPostReplies 实现获取帖子回复接口
func (s *GRPCServer) GetPostReplies(ctx context.Context, req *proto.GetPostRepliesRequest) (*proto.GetPostRepliesResponse, error) {
	return s.postService.GetPostReplies(ctx, req)
}

// GetConversationThread 实现获取对话线程接口
func (s *GRPCServer) GetConversationThread(ctx context.Context, req *proto.GetConversationThreadRequest) (*proto.GetConversationThreadResponse, error) {
	return s.postService.GetConversationThread(ctx, req)
}

// DeleteReply 实现删除回复接口
func (s *GRPCServer) DeleteReply(ctx context.Context, req *proto.DeleteReplyRequest) (*proto.DeleteReplyResponse, error) {
	return s.postService.DeleteReply(ctx, req)
}

// GetReplyMention 实现获取回复提及接口
func (s *GRPCServer) GetReplyMention(ctx context.Context, req *proto.GetReplyMentionRequest) (*proto.GetReplyMentionResponse, error) {
	return s.postService.GetReplyMention(ctx, req)
}

// Run 启动gRPC服务
func (s *GRPCServer) Run(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	// 注册到Consul
	portInt := 50055 // content service port
	discovery.RegisterServiceToConsul(discovery.RegisterOptions{
		ServiceName:     "content-service",
		ServicePort:     portInt,
		HealthCheckType: "grpc",
	})

	grpcServer := grpc.NewServer()
	proto.RegisterContentServiceServer(grpcServer, s)

	return grpcServer.Serve(lis)
}

