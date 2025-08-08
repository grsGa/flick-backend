package service

import (
	"backend/services/bookmark/internal/repository"
)

// NewBookmarkServiceFactory 创建书签服务工厂
func NewBookmarkServiceFactory(bookmarkRepo repository.BookmarkRepository) BookmarkServiceFactory {
	return &bookmarkServiceFactory{
		bookmarkRepo: bookmarkRepo,
	}
}

// BookmarkServiceFactory 书签服务工厂接口
type BookmarkServiceFactory interface {
	Create() BookmarkService
}

// bookmarkServiceFactory 书签服务工厂实现
type bookmarkServiceFactory struct {
	bookmarkRepo repository.BookmarkRepository
}

// Create 创建书签服务实例
func (f *bookmarkServiceFactory) Create() BookmarkService {
	return NewBookmarkService(f.bookmarkRepo)
}