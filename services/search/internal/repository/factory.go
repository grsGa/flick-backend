package repository

// NewSearchRepositoryFactory 创建搜索仓储工厂
func NewSearchRepositoryFactory() SearchRepositoryFactory {
	return &searchRepositoryFactory{}
}

// SearchRepositoryFactory 搜索仓储工厂接口
type SearchRepositoryFactory interface {
	Create() SearchRepository
}

// searchRepositoryFactory 搜索仓储工厂实现
type searchRepositoryFactory struct{}

// Create 创建搜索仓储实例
func (f *searchRepositoryFactory) Create() SearchRepository {
	return NewSearchRepository()
}