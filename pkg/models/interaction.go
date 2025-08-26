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
	Reply     *string   `gorm:"type:text" json:"reply,omitempty"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

// Reply 回复模型
type Reply struct {
	ID            string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PostID        string    `gorm:"type:uuid;not null;index" json:"post_id"`
	UserID        string    `gorm:"type:uuid;not null;index" json:"user_id"`
	Content       string    `gorm:"type:text;not null" json:"content"`
	ParentReplyID *string   `gorm:"type:uuid;index" json:"parent_reply_id,omitempty"` // 支持楼中楼
	CreatedAt     time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt     time.Time `gorm:"not null" json:"updated_at"`
	DeletedAt     *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

// PostStats 帖子统计模型 - 用于快速查询互动数据
type PostStats struct {
	PostID       string    `gorm:"type:uuid;primaryKey" json:"post_id"`
	LikeCount    int       `gorm:"default:0" json:"like_count"`
	ReplyCount   int       `gorm:"default:0" json:"reply_count"`
	RepostCount  int       `gorm:"default:0" json:"repost_count"`
	ViewCount    int       `gorm:"default:0" json:"view_count"`
	UpdatedAt    time.Time `gorm:"not null" json:"updated_at"`
}

// PollVote 投票记录模型
type PollVote struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PollID       string    `gorm:"type:uuid;not null;index" json:"poll_id"`
	PollOptionID string    `gorm:"type:uuid;not null;index" json:"poll_option_id"`
	UserID       string    `gorm:"type:uuid;not null;index" json:"user_id"`
	CreatedAt    time.Time `gorm:"not null" json:"created_at"`
}

// Report 举报模型
type Report struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     string    `gorm:"type:uuid;not null;index" json:"user_id"`
	TargetID   string    `gorm:"type:uuid;not null" json:"target_id"`
	TargetType string    `gorm:"type:varchar(20);not null;check:target_type IN ('post', 'user', 'reply')" json:"target_type"`
	Reason     string    `gorm:"type:text;not null" json:"reason"`
	CreatedAt  time.Time `gorm:"not null" json:"created_at"`
	DeletedAt  *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}
