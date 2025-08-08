package repository

import (
	"context"
	"backend/services/messages/proto"
)

// MessageRepository 定义消息仓储接口
type MessageRepository interface {
	// CreateConversation 创建会话
	CreateConversation(ctx context.Context, conversation *proto.Conversation) error
	
	// GetConversationByID 根据ID获取会话
	GetConversationByID(ctx context.Context, id string) (*proto.Conversation, error)
	
	// ListConversations 列出用户会话
	ListConversations(ctx context.Context, userID string, page, pageSize int32) ([]*proto.Conversation, int32, error)
	
	// CreateMessage 创建消息
	CreateMessage(ctx context.Context, message *proto.Message) error
	
	// GetMessageByID 根据ID获取消息
	GetMessageByID(ctx context.Context, id string) (*proto.Message, error)
	
	// ListMessages 列出会话消息
	ListMessages(ctx context.Context, conversationID string, page, pageSize int32) ([]*proto.Message, int32, error)
	
	// MarkAsRead 标记消息为已读
	MarkAsRead(ctx context.Context, conversationID, userID string) error
	
	// DeleteConversation 删除会话
	DeleteConversation(ctx context.Context, id string) error
	
	// UpdateConversationLastMessage 更新会话最后消息
	UpdateConversationLastMessage(ctx context.Context, conversationID, lastMessage string) error
}