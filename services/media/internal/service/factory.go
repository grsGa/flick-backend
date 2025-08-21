package service

import (
	"github.com/flick/backend/services/media/internal/repository"
	"github.com/flick/backend/services/media/internal/storage"
)

// NewMediaServiceFactory 创建媒体服务工厂
func NewMediaServiceFactory(mediaRepo repository.MediaRepository, storage storage.MediaStorage) MediaServiceFactory {
	return &mediaServiceFactory{
		mediaRepo: mediaRepo,
		storage:   storage,
	}
}

// MediaServiceFactory 媒体服务工厂接口
type MediaServiceFactory interface {
	Create() MediaService
}

// mediaServiceFactory 媒体服务工厂实现
type mediaServiceFactory struct {
	mediaRepo repository.MediaRepository
	storage   storage.MediaStorage
}

// Create 创建媒体服务实例
func (f *mediaServiceFactory) Create() MediaService {
	return NewMediaService(f.mediaRepo, f.storage, "")
}
