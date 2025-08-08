package server

import (
	"context"
	"net"
	
	"google.golang.org/grpc"
	"backend/services/recommendation/internal/service"
	"backend/services/recommendation/proto"
)

// grpcServer gRPC服务实现
type grpcServer struct {
	proto.UnimplementedRecommendationServiceServer
	recommendationService service.RecommendationService
}

// NewGRPCServer 创建gRPC服务实例
func NewGRPCServer(recommendationService service.RecommendationService) *grpcServer {
	return &grpcServer{
		recommendationService: recommendationService,
	}
}

// GetRecommendations 实现获取推荐内容接口
func (s *grpcServer) GetRecommendations(ctx context.Context, req *proto.GetRecommendationsRequest) (*proto.GetRecommendationsResponse, error) {
	return s.recommendationService.GetRecommendations(ctx, req)
}

// RecordUserAction 实现记录用户行为接口
func (s *grpcServer) RecordUserAction(ctx context.Context, req *proto.RecordUserActionRequest) (*proto.RecordUserActionResponse, error) {
	return s.recommendationService.RecordUserAction(ctx, req)
}

// UpdateUserInterest 实现更新用户兴趣接口
func (s *grpcServer) UpdateUserInterest(ctx context.Context, req *proto.UpdateUserInterestRequest) (*proto.UpdateUserInterestResponse, error) {
	return s.recommendationService.UpdateUserInterest(ctx, req)
}

// Run 启动gRPC服务
func (s *grpcServer) Run(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	
	grpcServer := grpc.NewServer()
	proto.RegisterRecommendationServiceServer(grpcServer, s)
	
	return grpcServer.Serve(lis)
}