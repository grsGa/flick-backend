package service

import (
	"backend/services/recommendation/internal/repository"
)

// NewRecommendationServiceFactory 创建推荐服务工厂
func NewRecommendationServiceFactory(recommendationRepo repository.RecommendationRepository) RecommendationServiceFactory {
	return &recommendationServiceFactory{
		recommendationRepo: recommendationRepo,
	}
}

// RecommendationServiceFactory 推荐服务工厂接口
type RecommendationServiceFactory interface {
	Create() RecommendationService
}

// recommendationServiceFactory 推荐服务工厂实现
type recommendationServiceFactory struct {
	recommendationRepo repository.RecommendationRepository
}

// Create 创建推荐服务实例
func (f *recommendationServiceFactory) Create() RecommendationService {
	return NewRecommendationService(f.recommendationRepo)
}