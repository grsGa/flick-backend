package server

import (
	"github.com/flick/backend/pkg/config"
	"github.com/flick/backend/services/media/internal/service"
)

// NewGRPCServerFactory 创建gRPC服务工厂
func NewGRPCServerFactory(mediaService service.MediaService, cfg *config.Config) GRPCServerFactory {
	return &grpcServerFactory{
		mediaService: mediaService,
		config:       cfg,
	}
}

// GRPCServerFactory gRPC服务工厂接口
type GRPCServerFactory interface {
	Create() *grpcServer
}

// grpcServerFactory gRPC服务工厂实现
type grpcServerFactory struct {
	mediaService service.MediaService
	config       *config.Config
}

// Create 创建gRPC服务实例
func (f *grpcServerFactory) Create() *grpcServer {
	return NewGRPCServer(f.mediaService, f.config)
}
