package models

import (
	"time"
)

// Notification 通知模型
type Notification struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ReceiverID string    `gorm:"type:uuid;not null;index" json:"receiver_id"`
	ActorID    string    `gorm:"type:uuid;not null" json:"actor_id"`
	ActionType string    `gorm:"type:varchar(20);not null;check:action_type IN ('follow', 'like', 'reply', 'repost')" json:"action_type"`
	TargetType string    `gorm:"type:varchar(20);not null;check:target_type IN ('post', 'user')" json:"target_type"`
	TargetID   string    `gorm:"type:uuid;not null" json:"target_id"`
	IsRead     bool      `gorm:"not null;default:false" json:"is_read"`
	CreatedAt  time.Time `gorm:"not null" json:"created_at"`
	DeletedAt  *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}
