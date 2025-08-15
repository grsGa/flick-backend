package models

import (
	"time"
)

// Bookmark 书签模型
type Bookmark struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    string    `gorm:"type:uuid;not null;index" json:"user_id"`
	PostID    string    `gorm:"type:uuid;not null;index" json:"post_id"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}
