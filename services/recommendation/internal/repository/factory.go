package repository

// NewRecommendationRepositoryFactory 创建推荐仓储工厂
func NewRecommendationRepositoryFactory() RecommendationRepositoryFactory {
	return &recommendationRepositoryFactory{}
}

// RecommendationRepositoryFactory 推荐仓储工厂接口
type RecommendationRepositoryFactory interface {
	Create() RecommendationRepository
}

// recommendationRepositoryFactory 推荐仓储工厂实现
type recommendationRepositoryFactory struct{}

// Create 创建推荐仓储实例
func (f *recommendationRepositoryFactory) Create() RecommendationRepository {
	return NewRecommendationRepository()
}