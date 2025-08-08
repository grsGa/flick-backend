package service

import (
	"context"
	"backend/services/messages/proto"
)

// MessageService 定义消息服务接口
type MessageService interface {
	// CreateConversation 创建会话
	CreateConversation(ctx context.Context, req *proto.CreateConversationRequest) (*proto.CreateConversationResponse, error)
	
	// ListConversations 获取会话列表
	ListConversations(ctx context.Context, req *proto.ListConversationsRequest) (*proto.ListConversationsResponse, error)
	
	// GetConversation 获取会话详情
	GetConversation(ctx context.Context, req *proto.GetConversationRequest) (*proto.GetConversationResponse, error)
	
	// SendMessage 发送消息
	SendMessage(ctx context.Context, req *proto.SendMessageRequest) (*proto.SendMessageResponse, error)
	
	// ListMessages 获取消息列表
	ListMessages(ctx context.Context, req *proto.ListMessagesRequest) (*proto.ListMessagesResponse, error)
	
	// MarkAsRead 标记消息为已读
	MarkAsRead(ctx context.Context, req *proto.MarkAsReadRequest) (*proto.MarkAsReadResponse, error)
	
	// DeleteConversation 删除会话
	DeleteConversation(ctx context.Context, req *proto.DeleteConversationRequest) (*proto.DeleteConversationResponse, error)
}