package service

import (
	"context"
	"time"

	"github.com/flick/backend/services/messages/internal/repository"
	"github.com/flick/backend/services/messages/proto"
	"github.com/google/uuid"
)

// messageService 消息服务实现
type messageService struct {
	messageRepo repository.MessageRepository
}

// NewMessageService 创建消息服务实例
func NewMessageService(messageRepo repository.MessageRepository) MessageService {
	return &messageService{
		messageRepo: messageRepo,
	}
}

// CreateConversation 创建会话
func (s *messageService) CreateConversation(ctx context.Context, req *proto.CreateConversationRequest) (*proto.CreateConversationResponse, error) {
	// 创建会话对象
	conversation := &proto.Conversation{
		Id:        uuid.New().String(),
		User1Id:   req.User1Id,
		User2Id:   req.User2Id,
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	// 保存到数据库
	err := s.messageRepo.CreateConversation(ctx, conversation)
	if err != nil {
		return &proto.CreateConversationResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to create conversation: " + err.Error(),
			},
		}, err
	}

	return &proto.CreateConversationResponse{
		Conversation: conversation,
	}, nil
}

// ListConversations 获取会话列表
func (s *messageService) ListConversations(ctx context.Context, req *proto.ListConversationsRequest) (*proto.ListConversationsResponse, error) {
	conversations, total, err := s.messageRepo.ListConversations(ctx, req.UserId, req.Page, req.PageSize)
	if err != nil {
		return &proto.ListConversationsResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to list conversations: " + err.Error(),
			},
		}, err
	}

	return &proto.ListConversationsResponse{
		Conversations: conversations,
		Total:         total,
	}, nil
}

// GetConversation 获取会话详情
func (s *messageService) GetConversation(ctx context.Context, req *proto.GetConversationRequest) (*proto.GetConversationResponse, error) {
	conversation, err := s.messageRepo.GetConversationByID(ctx, req.ConversationId)
	if err != nil {
		return &proto.GetConversationResponse{
			Error: &proto.Error{
				Code:    404,
				Message: "Conversation not found: " + err.Error(),
			},
		}, err
	}

	return &proto.GetConversationResponse{
		Conversation: conversation,
	}, nil
}

// SendMessage 发送消息
func (s *messageService) SendMessage(ctx context.Context, req *proto.SendMessageRequest) (*proto.SendMessageResponse, error) {
	// 创建消息对象
	message := &proto.Message{
		Id:             uuid.New().String(),
		ConversationId: req.ConversationId,
		SenderId:       req.SenderId,
		Content:        req.Content,
		MediaUrl:       req.MediaUrl,
		IsRead:         false,
		CreatedAt:      time.Now().Format(time.RFC3339),
	}

	// 保存到数据库
	err := s.messageRepo.CreateMessage(ctx, message)
	if err != nil {
		return &proto.SendMessageResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to send message: " + err.Error(),
			},
		}, err
	}

	// 更新会话的最后消息
	s.messageRepo.UpdateConversationLastMessage(ctx, req.ConversationId, req.Content)

	return &proto.SendMessageResponse{
		Message: message,
	}, nil
}

// ListMessages 获取消息列表
func (s *messageService) ListMessages(ctx context.Context, req *proto.ListMessagesRequest) (*proto.ListMessagesResponse, error) {
	messages, total, err := s.messageRepo.ListMessages(ctx, req.ConversationId, req.Page, req.PageSize)
	if err != nil {
		return &proto.ListMessagesResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to list messages: " + err.Error(),
			},
		}, err
	}

	return &proto.ListMessagesResponse{
		Messages: messages,
		Total:    total,
	}, nil
}

// MarkAsRead 标记消息为已读
func (s *messageService) MarkAsRead(ctx context.Context, req *proto.MarkAsReadRequest) (*proto.MarkAsReadResponse, error) {
	err := s.messageRepo.MarkAsRead(ctx, req.ConversationId, req.UserId)
	if err != nil {
		return &proto.MarkAsReadResponse{
			Success: false,
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to mark messages as read: " + err.Error(),
			},
		}, err
	}

	return &proto.MarkAsReadResponse{
		Success: true,
	}, nil
}

// DeleteConversation 删除会话
func (s *messageService) DeleteConversation(ctx context.Context, req *proto.DeleteConversationRequest) (*proto.DeleteConversationResponse, error) {
	err := s.messageRepo.DeleteConversation(ctx, req.ConversationId)
	if err != nil {
		return &proto.DeleteConversationResponse{
			Success: false,
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to delete conversation: " + err.Error(),
			},
		}, err
	}

	return &proto.DeleteConversationResponse{
		Success: true,
	}, nil
}
