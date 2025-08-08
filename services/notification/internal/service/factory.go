package service

import (
	"backend/services/notification/internal/repository"
)

// NewNotificationServiceFactory 创建通知服务工厂
func NewNotificationServiceFactory(notificationRepo repository.NotificationRepository) NotificationServiceFactory {
	return &notificationServiceFactory{
		notificationRepo: notificationRepo,
	}
}

// NotificationServiceFactory 通知服务工厂接口
type NotificationServiceFactory interface {
	Create() NotificationService
}

// notificationServiceFactory 通知服务工厂实现
type notificationServiceFactory struct {
	notificationRepo repository.NotificationRepository
}

// Create 创建通知服务实例
func (f *notificationServiceFactory) Create() NotificationService {
	return NewNotificationService(f.notificationRepo)
}