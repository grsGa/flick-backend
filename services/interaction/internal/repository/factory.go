package repository

// NewInteractionRepositoryFactory 创建互动仓储工厂
func NewInteractionRepositoryFactory() InteractionRepositoryFactory {
	return &interactionRepositoryFactory{}
}

// InteractionRepositoryFactory 互动仓储工厂接口
type InteractionRepositoryFactory interface {
	Create() InteractionRepository
}

// interactionRepositoryFactory 互动仓储工厂实现
type interactionRepositoryFactory struct{}

// Create 创建互动仓储实例
func (f *interactionRepositoryFactory) Create() InteractionRepository {
	return NewInteractionRepository()
}
