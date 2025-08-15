package server

import (
	"context"
	"net"

	"github.com/flick/backend/services/interaction/internal/service"
	"github.com/flick/backend/services/interaction/proto"
	"google.golang.org/grpc"
)

// grpcServer gRPC服务实现
type grpcServer struct {
	proto.UnimplementedInteractionServiceServer
	interactionService service.InteractionService
}

// NewGRPCServer 创建gRPC服务实例
func NewGRPCServer(interactionService service.InteractionService) *grpcServer {
	return &grpcServer{
		interactionService: interactionService,
	}
}

// CreateFollow 实现创建关注接口
func (s *grpcServer) CreateFollow(ctx context.Context, req *proto.CreateFollowRequest) (*proto.CreateFollowResponse, error) {
	return s.interactionService.CreateFollow(ctx, req)
}

// DeleteFollow 实现删除关注接口
func (s *grpcServer) DeleteFollow(ctx context.Context, req *proto.DeleteFollowRequest) (*proto.DeleteFollowResponse, error) {
	return s.interactionService.DeleteFollow(ctx, req)
}

// IsFollowing 实现检查是否关注接口
func (s *grpcServer) IsFollowing(ctx context.Context, req *proto.IsFollowingRequest) (*proto.IsFollowingResponse, error) {
	return s.interactionService.IsFollowing(ctx, req)
}

// GetFollowers 实现获取粉丝列表接口
func (s *grpcServer) GetFollowers(ctx context.Context, req *proto.GetFollowersRequest) (*proto.GetFollowersResponse, error) {
	return s.interactionService.GetFollowers(ctx, req)
}

// GetFollowing 实现获取关注列表接口
func (s *grpcServer) GetFollowing(ctx context.Context, req *proto.GetFollowingRequest) (*proto.GetFollowingResponse, error) {
	return s.interactionService.GetFollowing(ctx, req)
}

// CreateLike 实现创建点赞接口
func (s *grpcServer) CreateLike(ctx context.Context, req *proto.CreateLikeRequest) (*proto.CreateLikeResponse, error) {
	return s.interactionService.CreateLike(ctx, req)
}

// DeleteLike 实现删除点赞接口
func (s *grpcServer) DeleteLike(ctx context.Context, req *proto.DeleteLikeRequest) (*proto.DeleteLikeResponse, error) {
	return s.interactionService.DeleteLike(ctx, req)
}

// IsLiked 实现检查是否点赞接口
func (s *grpcServer) IsLiked(ctx context.Context, req *proto.IsLikedRequest) (*proto.IsLikedResponse, error) {
	return s.interactionService.IsLiked(ctx, req)
}

// GetLikes 实现获取点赞列表接口
func (s *grpcServer) GetLikes(ctx context.Context, req *proto.GetLikesRequest) (*proto.GetLikesResponse, error) {
	return s.interactionService.GetLikes(ctx, req)
}

// CreateRepost 实现创建转发接口
func (s *grpcServer) CreateRepost(ctx context.Context, req *proto.CreateRepostRequest) (*proto.CreateRepostResponse, error) {
	return s.interactionService.CreateRepost(ctx, req)
}

// DeleteRepost 实现删除转发接口
func (s *grpcServer) DeleteRepost(ctx context.Context, req *proto.DeleteRepostRequest) (*proto.DeleteRepostResponse, error) {
	return s.interactionService.DeleteRepost(ctx, req)
}

// CreateReport 实现创建举报接口
func (s *grpcServer) CreateReport(ctx context.Context, req *proto.CreateReportRequest) (*proto.CreateReportResponse, error) {
	return s.interactionService.CreateReport(ctx, req)
}

// Run 启动gRPC服务
func (s *grpcServer) Run(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	proto.RegisterInteractionServiceServer(grpcServer, s)

	return grpcServer.Serve(lis)
}
