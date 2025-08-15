package service

import (
	"github.com/flick/backend/services/messages/internal/repository"
)

// NewMessageServiceFactory 创建消息服务工厂
func NewMessageServiceFactory(messageRepo repository.MessageRepository) MessageServiceFactory {
	return &messageServiceFactory{
		messageRepo: messageRepo,
	}
}

// MessageServiceFactory 消息服务工厂接口
type MessageServiceFactory interface {
	Create() MessageService
}

// messageServiceFactory 消息服务工厂实现
type messageServiceFactory struct {
	messageRepo repository.MessageRepository
}

// Create 创建消息服务实例
func (f *messageServiceFactory) Create() MessageService {
	return NewMessageService(f.messageRepo)
}
