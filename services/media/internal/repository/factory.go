package repository

// NewMediaRepositoryFactory 创建媒体仓储工厂
func NewMediaRepositoryFactory() MediaRepositoryFactory {
	return &mediaRepositoryFactory{}
}

// MediaRepositoryFactory 媒体仓储工厂接口
type MediaRepositoryFactory interface {
	Create() MediaRepository
}

// mediaRepositoryFactory 媒体仓储工厂实现
type mediaRepositoryFactory struct{}

// Create 创建媒体仓储实例
func (f *mediaRepositoryFactory) Create() MediaRepository {
	return NewMediaRepository()
}