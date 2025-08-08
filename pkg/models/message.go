package models

import (
	"time"
)

// Conversation 会话模型
type Conversation struct {
	ID          string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	User1ID     string     `gorm:"type:uuid;not null;index" json:"user1_id"`
	User2ID     string     `gorm:"type:uuid;not null;index" json:"user2_id"`
	LastMessage *string    `gorm:"type:text" json:"last_message,omitempty"`
	CreatedAt   time.Time  `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"not null" json:"updated_at"`
	DeletedAt   *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

// Message 消息模型
type Message struct {
	ID             string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ConversationID string     `gorm:"type:uuid;not null;index" json:"conversation_id"`
	SenderID       string     `gorm:"type:uuid;not null" json:"sender_id"`
	Content        string     `gorm:"type:text;not null" json:"content"`
	MediaURL       *string    `gorm:"type:text" json:"media_url,omitempty"`
	CreatedAt      time.Time  `gorm:"not null" json:"created_at"`
	DeletedAt      *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}
