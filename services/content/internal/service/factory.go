package service

import (
	"github.com/flick/backend/services/content/internal/repository"
	"github.com/flick/backend/services/content/internal/events"
	"github.com/flick/backend/services/content/internal/media"
)

// NewContentServiceFactory 创建内容服务工厂
func NewContentServiceFactory(contentRepo repository.ContentRepository, postRepo repository.PostRepository) ContentServiceFactory {
	return &contentServiceFactory{
		contentRepo: contentRepo,
		postRepo:    postRepo,
	}
}

// ContentServiceFactory 内容服务工厂接口
type ContentServiceFactory interface {
	Create() ContentService
	CreatePostService() PostService
}

// contentServiceFactory 内容服务工厂实现
type contentServiceFactory struct {
	contentRepo repository.ContentRepository
	postRepo    repository.PostRepository
}

// Create 创建内容服务实例
func (f *contentServiceFactory) Create() ContentService {
	return NewContentService(f.contentRepo)
}

// CreatePostService 创建帖子服务实例（带事件发布和媒体处理）
func (f *contentServiceFactory) CreatePostService() PostService {
	// 创建事件发布器
	eventPublisher := events.NewEventPublisher()
	
	// 创建媒体处理器
	mediaProcessor := media.NewMediaProcessor(eventPublisher)
	
	// 创建帖子服务
	return NewPostService(f.postRepo, eventPublisher, mediaProcessor)
}
