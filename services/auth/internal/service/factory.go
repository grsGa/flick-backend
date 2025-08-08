package service

import (
	"backend/pkg/config"
	"backend/services/auth/internal/repository"
)

// NewAuthServiceFactory 创建认证服务工厂
func NewAuthServiceFactory(authRepo repository.AuthRepository, cfg *config.Config) AuthServiceFactory {
	return &authServiceFactory{
		authRepo: authRepo,
		cfg:      cfg,
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
}

// Create 创建认证服务实例
func (f *authServiceFactory) Create() AuthService {
	return NewAuthService(f.authRepo, f.cfg)
}
