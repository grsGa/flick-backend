package service

import (
	"github.com/flick/backend/services/content/internal/repository"
)

// NewContentServiceFactory 创建内容服务工厂
func NewContentServiceFactory(contentRepo repository.ContentRepository) ContentServiceFactory {
	return &contentServiceFactory{
		contentRepo: contentRepo,
	}
}

// ContentServiceFactory 内容服务工厂接口
type ContentServiceFactory interface {
	Create() ContentService
}

// contentServiceFactory 内容服务工厂实现
type contentServiceFactory struct {
	contentRepo repository.ContentRepository
}

// Create 创建内容服务实例
func (f *contentServiceFactory) Create() ContentService {
	return NewContentService(f.contentRepo)
}
