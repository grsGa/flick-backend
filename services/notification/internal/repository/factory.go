package repository

// NewNotificationRepositoryFactory 创建通知仓储工厂
func NewNotificationRepositoryFactory() NotificationRepositoryFactory {
	return &notificationRepositoryFactory{}
}

// NotificationRepositoryFactory 通知仓储工厂接口
type NotificationRepositoryFactory interface {
	Create() NotificationRepository
}

// notificationRepositoryFactory 通知仓储工厂实现
type notificationRepositoryFactory struct{}

// Create 创建通知仓储实例
func (f *notificationRepositoryFactory) Create() NotificationRepository {
	return NewNotificationRepository()
}
