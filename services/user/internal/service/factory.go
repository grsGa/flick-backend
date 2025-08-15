package service

import (
	"github.com/flick/backend/pkg/config"
	"github.com/flick/backend/services/user/internal/repository"

	"go.uber.org/zap"
)

// NewUserServiceFactory 创建用户服务工厂
func NewUserServiceFactory(userRepo repository.UserRepository, cfg *config.Config, logger *zap.Logger) UserServiceFactory {
	return &userServiceFactory{
		userRepo: userRepo,
		cfg:      cfg,
		logger:   logger,
	}
}

// UserServiceFactory 用户服务工厂接口
type UserServiceFactory interface {
	Create() UserService
}

// userServiceFactory 用户服务工厂实现
type userServiceFactory struct {
	userRepo repository.UserRepository
	cfg      *config.Config
	logger   *zap.Logger
}

// Create 创建用户服务实例
func (f *userServiceFactory) Create() UserService {
	return NewUserService(f.userRepo, f.cfg, f.logger)
}
