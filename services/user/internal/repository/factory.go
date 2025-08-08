package repository

// NewUserRepositoryFactory 创建用户仓储工厂
func NewUserRepositoryFactory() UserRepositoryFactory {
	return &userRepositoryFactory{}
}

// UserRepositoryFactory 用户仓储工厂接口
type UserRepositoryFactory interface {
	Create() UserRepository
}

// userRepositoryFactory 用户仓储工厂实现
type userRepositoryFactory struct{}

// Create 创建用户仓储实例
func (f *userRepositoryFactory) Create() UserRepository {
	return NewUserRepository()
}