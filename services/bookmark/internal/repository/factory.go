package repository

// NewBookmarkRepositoryFactory 创建书签仓储工厂
func NewBookmarkRepositoryFactory() BookmarkRepositoryFactory {
	return &bookmarkRepositoryFactory{}
}

// BookmarkRepositoryFactory 书签仓储工厂接口
type BookmarkRepositoryFactory interface {
	Create() BookmarkRepository
}

// bookmarkRepositoryFactory 书签仓储工厂实现
type bookmarkRepositoryFactory struct{}

// Create 创建书签仓储实例
func (f *bookmarkRepositoryFactory) Create() BookmarkRepository {
	return NewBookmarkRepository()
}