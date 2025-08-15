package server

import (
	"github.com/flick/backend/services/interaction/internal/service"
)

// NewGRPCServerFactory 创建gRPC服务工厂
func NewGRPCServerFactory(interactionService service.InteractionService) GRPCServerFactory {
	return &grpcServerFactory{
		interactionService: interactionService,
	}
}

// GRPCServerFactory gRPC服务工厂接口
type GRPCServerFactory interface {
	Create() *grpcServer
}

// grpcServerFactory gRPC服务工厂实现
type grpcServerFactory struct {
	interactionService service.InteractionService
}

// Create 创建gRPC服务实例
func (f *grpcServerFactory) Create() *grpcServer {
	return NewGRPCServer(f.interactionService)
}
