package repository

import (
	"context"
	"errors"
	"time"

	"github.com/flick/backend/pkg/database"
	"github.com/flick/backend/pkg/models"
	"github.com/flick/backend/services/messages/proto"

	"gorm.io/gorm"
)

// messageRepository 消息仓储实现
type messageRepository struct {
	db *gorm.DB
}

// NewMessageRepository 创建消息仓储实例
func NewMessageRepository() MessageRepository {
	return &messageRepository{
		db: database.GetDB(),
	}
}

// CreateConversation 创建会话
func (r *messageRepository) CreateConversation(ctx context.Context, conversation *proto.Conversation) error {
	createdAt, _ := time.Parse(time.RFC3339, conversation.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, conversation.UpdatedAt)
	c := &models.Conversation{
		ID:          conversation.Id,
		User1ID:     conversation.User1Id,
		User2ID:     conversation.User2Id,
		LastMessage: &conversation.LastMessage,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	return r.db.Create(c).Error
}

// GetConversationByID 根据ID获取会话
func (r *messageRepository) GetConversationByID(ctx context.Context, id string) (*proto.Conversation, error) {
	var conversation models.Conversation
	if err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&conversation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("conversation not found")
		}
		return nil, err
	}

	return &proto.Conversation{
		Id:          conversation.ID,
		User1Id:     conversation.User1ID,
		User2Id:     conversation.User2ID,
		LastMessage: toString(conversation.LastMessage),
		CreatedAt:   conversation.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   conversation.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// ListConversations 列出用户会话
func (r *messageRepository) ListConversations(ctx context.Context, userID string, page, pageSize int32) ([]*proto.Conversation, int32, error) {
	var conversations []models.Conversation
	var total int64

	// 查询总数
	if err := r.db.Model(&models.Conversation{}).Where("(user1_id = ? OR user2_id = ?) AND deleted_at IS NULL", userID, userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询列表
	offset := (page - 1) * pageSize
	if err := r.db.Where("(user1_id = ? OR user2_id = ?) AND deleted_at IS NULL", userID, userID).Offset(int(offset)).Limit(int(pageSize)).Find(&conversations).Error; err != nil {
		return nil, 0, err
	}

	protoConversations := make([]*proto.Conversation, len(conversations))
	for i, conversation := range conversations {
		protoConversations[i] = &proto.Conversation{
			Id:          conversation.ID,
			User1Id:     conversation.User1ID,
			User2Id:     conversation.User2ID,
			LastMessage: toString(conversation.LastMessage),
			CreatedAt:   conversation.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   conversation.UpdatedAt.Format(time.RFC3339),
		}
	}

	return protoConversations, int32(total), nil
}

// CreateMessage 创建消息
func (r *messageRepository) CreateMessage(ctx context.Context, message *proto.Message) error {
	createdAt, _ := time.Parse(time.RFC3339, message.CreatedAt)
	m := &models.Message{
		ID:             message.Id,
		ConversationID: message.ConversationId,
		SenderID:       message.SenderId,
		Content:        message.Content,
		CreatedAt:      createdAt,
	}

	if message.MediaUrl != "" {
		m.MediaURL = &message.MediaUrl
	}

	return r.db.Create(m).Error
}

// GetMessageByID 根据ID获取消息
func (r *messageRepository) GetMessageByID(ctx context.Context, id string) (*proto.Message, error) {
	var message models.Message
	if err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&message).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("message not found")
		}
		return nil, err
	}

	protoMessage := &proto.Message{
		Id:             message.ID,
		ConversationId: message.ConversationID,
		SenderId:       message.SenderID,
		Content:        message.Content,
		IsRead:         message.DeletedAt == nil, // 简化处理
		CreatedAt:      message.CreatedAt.Format(time.RFC3339),
	}

	if message.MediaURL != nil {
		protoMessage.MediaUrl = *message.MediaURL
	}

	return protoMessage, nil
}

// ListMessages 列出会话消息
func (r *messageRepository) ListMessages(ctx context.Context, conversationID string, page, pageSize int32) ([]*proto.Message, int32, error) {
	var messages []models.Message
	var total int64

	// 查询总数
	if err := r.db.Model(&models.Message{}).Where("conversation_id = ? AND deleted_at IS NULL", conversationID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询列表
	offset := (page - 1) * pageSize
	if err := r.db.Where("conversation_id = ? AND deleted_at IS NULL", conversationID).Offset(int(offset)).Limit(int(pageSize)).Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	protoMessages := make([]*proto.Message, len(messages))
	for i, message := range messages {
		protoMessages[i] = &proto.Message{
			Id:             message.ID,
			ConversationId: message.ConversationID,
			SenderId:       message.SenderID,
			Content:        message.Content,
			IsRead:         message.DeletedAt == nil, // 简化处理
			CreatedAt:      message.CreatedAt.Format(time.RFC3339),
		}

		if message.MediaURL != nil {
			protoMessages[i].MediaUrl = *message.MediaURL
		}
	}

	return protoMessages, int32(total), nil
}

// MarkAsRead 标记消息为已读
func (r *messageRepository) MarkAsRead(ctx context.Context, conversationID, userID string) error {
	// 更新会话中用户不是发送者的消息为已读
	return r.db.Model(&models.Message{}).
		Where("conversation_id = ? AND sender_id != ?", conversationID, userID).
		Update("deleted_at", nil).Error
}

// DeleteConversation 删除会话
func (r *messageRepository) DeleteConversation(ctx context.Context, id string) error {
	// 软删除会话
	return r.db.Where("id = ?", id).Delete(&models.Conversation{}).Error
}

// UpdateConversationLastMessage 更新会话最后消息
func (r *messageRepository) UpdateConversationLastMessage(ctx context.Context, conversationID, lastMessage string) error {
	return r.db.Model(&models.Conversation{}).
		Where("id = ? AND deleted_at IS NULL", conversationID).
		Update("last_message", lastMessage).
		Update("updated_at", time.Now()).Error
}

// toString 将*string转换为string
func toString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
