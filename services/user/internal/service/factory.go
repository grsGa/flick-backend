package service

import (
	"backend/pkg/config"
	"backend/services/user/internal/repository"
)

// NewUserServiceFactory 创建用户服务工厂
func NewUserServiceFactory(userRepo repository.UserRepository, cfg *config.Config) UserServiceFactory {
	return &userServiceFactory{
		userRepo: userRepo,
		cfg:      cfg,
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
}

// Create 创建用户服务实例
func (f *userServiceFactory) Create() UserService {
	return NewUserService(f.userRepo, f.cfg)
}
