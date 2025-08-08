package repository

// NewContentRepositoryFactory 创建内容仓储工厂
func NewContentRepositoryFactory() ContentRepositoryFactory {
	return &contentRepositoryFactory{}
}

// ContentRepositoryFactory 内容仓储工厂接口
type ContentRepositoryFactory interface {
	Create() ContentRepository
}

// contentRepositoryFactory 内容仓储工厂实现
type contentRepositoryFactory struct{}

// Create 创建内容仓储实例
func (f *contentRepositoryFactory) Create() ContentRepository {
	return NewContentRepository()
}