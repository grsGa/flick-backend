package server

import (
	"github.com/flick/backend/services/search/internal/service"
)

// NewGRPCServerFactory 创建gRPC服务工厂
func NewGRPCServerFactory(searchService service.SearchService) GRPCServerFactory {
	return &grpcServerFactory{
		searchService: searchService,
	}
}

// GRPCServerFactory gRPC服务工厂接口
type GRPCServerFactory interface {
	Create() *grpcServer
}

// grpcServerFactory gRPC服务工厂实现
type grpcServerFactory struct {
	searchService service.SearchService
}

// Create 创建gRPC服务实例
func (f *grpcServerFactory) Create() *grpcServer {
	return NewGRPCServer(f.searchService)
}
