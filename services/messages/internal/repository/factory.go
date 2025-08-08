package repository

// NewMessageRepositoryFactory 创建消息仓储工厂
func NewMessageRepositoryFactory() MessageRepositoryFactory {
	return &messageRepositoryFactory{}
}

// MessageRepositoryFactory 消息仓储工厂接口
type MessageRepositoryFactory interface {
	Create() MessageRepository
}

// messageRepositoryFactory 消息仓储工厂实现
type messageRepositoryFactory struct{}

// Create 创建消息仓储实例
func (f *messageRepositoryFactory) Create() MessageRepository {
	return NewMessageRepository()
}