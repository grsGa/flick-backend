package service

import (
	"backend/services/interaction/internal/repository"
)

// NewInteractionServiceFactory 创建互动服务工厂
func NewInteractionServiceFactory(interactionRepo repository.InteractionRepository) InteractionServiceFactory {
	return &interactionServiceFactory{
		interactionRepo: interactionRepo,
	}
}

// InteractionServiceFactory 互动服务工厂接口
type InteractionServiceFactory interface {
	Create() InteractionService
}

// interactionServiceFactory 互动服务工厂实现
type interactionServiceFactory struct {
	interactionRepo repository.InteractionRepository
}

// Create 创建互动服务实例
func (f *interactionServiceFactory) Create() InteractionService {
	return NewInteractionService(f.interactionRepo)
}