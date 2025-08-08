package repository

// NewAuthRepositoryFactory 创建认证仓储工厂
func NewAuthRepositoryFactory() AuthRepositoryFactory {
	return &authRepositoryFactory{}
}

// AuthRepositoryFactory 认证仓储工厂接口
type AuthRepositoryFactory interface {
	Create() AuthRepository
}

// authRepositoryFactory 认证仓储工厂实现
type authRepositoryFactory struct{}

// Create 创建认证仓储实例
func (f *authRepositoryFactory) Create() AuthRepository {
	return NewAuthRepository()
}