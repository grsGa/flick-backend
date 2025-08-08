package service

import (
	"backend/services/search/internal/repository"
)

// NewSearchServiceFactory 创建搜索服务工厂
func NewSearchServiceFactory(searchRepo repository.SearchRepository) SearchServiceFactory {
	return &searchServiceFactory{
		searchRepo: searchRepo,
	}
}

// SearchServiceFactory 搜索服务工厂接口
type SearchServiceFactory interface {
	Create() SearchService
}

// searchServiceFactory 搜索服务工厂实现
type searchServiceFactory struct {
	searchRepo repository.SearchRepository
}

// Create 创建搜索服务实例
func (f *searchServiceFactory) Create() SearchService {
	return NewSearchService(f.searchRepo)
}