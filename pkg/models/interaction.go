package models

import (
	"time"
)

// Follow 关注关系模型
type Follow struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FollowerID string    `gorm:"type:uuid;not null;index" json:"follower_id"`
	FolloweeID string    `gorm:"type:uuid;not null;index" json:"followee_id"`
	CreatedAt  time.Time `gorm:"not null" json:"created_at"`
	DeletedAt  *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

// Like 点赞模型
type Like struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    string    `gorm:"type:uuid;not null;index" json:"user_id"`
	PostID    string    `gorm:"type:uuid;not null;index" json:"post_id"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

// Repost 转发模型
type Repost struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    string    `gorm:"type:uuid;not null;index" json:"user_id"`
	PostID    string    `gorm:"type:uuid;not null;index" json:"post_id"`
	Comment   *string   `gorm:"type:text" json:"comment,omitempty"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

// Report 举报模型
type Report struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     string    `gorm:"type:uuid;not null;index" json:"user_id"`
	TargetID   string    `gorm:"type:uuid;not null" json:"target_id"`
	TargetType string    `gorm:"type:varchar(20);not null;check:target_type IN ('post', 'user')" json:"target_type"`
	Reason     string    `gorm:"type:text;not null" json:"reason"`
	CreatedAt  time.Time `gorm:"not null" json:"created_at"`
	DeletedAt  *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}