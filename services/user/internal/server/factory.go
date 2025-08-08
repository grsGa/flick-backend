package server

import (
	"backend/services/user/internal/service"
)

// NewGRPCServerFactory 创建gRPC服务工厂
func NewGRPCServerFactory(userService service.UserService) GRPCServerFactory {
	return &grpcServerFactory{
		userService: userService,
	}
}

// GRPCServerFactory gRPC服务工厂接口
type GRPCServerFactory interface {
	Create() *grpcServer
}

// grpcServerFactory gRPC服务工厂实现
type grpcServerFactory struct {
	userService service.UserService
}

// Create 创建gRPC服务实例
func (f *grpcServerFactory) Create() *grpcServer {
	return NewGRPCServer(f.userService)
}