package server

import (
	"backend/services/recommendation/internal/service"
)

// NewGRPCServerFactory 创建gRPC服务工厂
func NewGRPCServerFactory(recommendationService service.RecommendationService) GRPCServerFactory {
	return &grpcServerFactory{
		recommendationService: recommendationService,
	}
}

// GRPCServerFactory gRPC服务工厂接口
type GRPCServerFactory interface {
	Create() *grpcServer
}

// grpcServerFactory gRPC服务工厂实现
type grpcServerFactory struct {
	recommendationService service.RecommendationService
}

// Create 创建gRPC服务实例
func (f *grpcServerFactory) Create() *grpcServer {
	return NewGRPCServer(f.recommendationService)
}