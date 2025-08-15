package service

import (
	"github.com/flick/backend/pkg/config"
	"github.com/flick/backend/services/auth/internal/repository"

	"go.uber.org/zap"
)

// NewAuthServiceFactory 创建认证服务工厂
func NewAuthServiceFactory(authRepo repository.AuthRepository, cfg *config.Config, logger *zap.Logger) AuthServiceFactory {
	return &authServiceFactory{
		authRepo: authRepo,
		cfg:      cfg,
		logger:   logger,
	}
}

// AuthServiceFactory 认证服务工厂接口
type AuthServiceFactory interface {
	Create() AuthService
}

// authServiceFactory 认证服务工厂实现
type authServiceFactory struct {
	authRepo repository.AuthRepository
	cfg      *config.Config
	logger   *zap.Logger
}

// Create 创建认证服务实例
func (f *authServiceFactory) Create() AuthService {
	return NewAuthService(f.authRepo, f.cfg, f.logger)
}
