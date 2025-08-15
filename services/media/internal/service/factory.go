package service

import (
	"github.com/flick/backend/services/media/internal/repository"
)

// NewMediaServiceFactory 创建媒体服务工厂
func NewMediaServiceFactory(mediaRepo repository.MediaRepository) MediaServiceFactory {
	return &mediaServiceFactory{
		mediaRepo: mediaRepo,
	}
}

// MediaServiceFactory 媒体服务工厂接口
type MediaServiceFactory interface {
	Create() MediaService
}

// mediaServiceFactory 媒体服务工厂实现
type mediaServiceFactory struct {
	mediaRepo repository.MediaRepository
}

// Create 创建媒体服务实例
func (f *mediaServiceFactory) Create() MediaService {
	return NewMediaService(f.mediaRepo)
}
