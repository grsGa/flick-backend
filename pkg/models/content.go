package models

import (
	"time"
)

// Post 帖子模型
type Post struct {
	ID              string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID          string    `gorm:"type:uuid;not null;index" json:"user_id"`
	Content         string    `gorm:"type:text;not null" json:"content"`
	Visibility      string    `gorm:"type:varchar(20);not null;check:visibility IN ('public', 'private', 'followers');default:'public'" json:"visibility"`
	ReplyPermission string    `gorm:"type:varchar(20);not null;check:reply_permission IN ('EVERYONE', 'FOLLOWING', 'MENTIONED_ONLY');default:'EVERYONE'" json:"reply_permission"`
	ParentID        *string   `gorm:"type:uuid;index" json:"parent_id,omitempty"`
	RepostID        *string   `gorm:"type:uuid;index" json:"repost_id,omitempty"`
	HasMedia        bool      `gorm:"default:false" json:"has_media"`
	HasPoll         bool      `gorm:"default:false" json:"has_poll"`
	CreatedAt       time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt       time.Time `gorm:"not null" json:"updated_at"`
	DeletedAt       *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

// MediaAttachment 媒体附件模型
type MediaAttachment struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PostID    *string   `gorm:"type:uuid;index" json:"post_id,omitempty"` // 修改为可选，支持头像/横幅等独立文件
	URL       string    `gorm:"type:text;not null" json:"url"`
	Type      string    `gorm:"type:varchar(20);not null" json:"type"` // 扩展长度支持 avatars/banners
	AltText   *string   `gorm:"type:text" json:"alt_text,omitempty"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

// PostMention 帖子提及模型 - 用于 @username 功能
type PostMention struct {
	ID             string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PostID         string    `gorm:"type:uuid;not null;index" json:"post_id"`
	MentionedUserID string    `gorm:"type:uuid;not null;index" json:"mentioned_user_id"`
	CreatedAt      time.Time `gorm:"not null" json:"created_at"`
}

// PostTag 帖子标签模型 - 用于 #hashtag 功能
type PostTag struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PostID    string    `gorm:"type:uuid;not null;index" json:"post_id"`
	Tag       string    `gorm:"type:varchar(100);not null;index" json:"tag"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
}

// Poll 投票模型
type Poll struct {
	ID              string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PostID          string    `gorm:"type:uuid;not null;unique;index" json:"post_id"`
	Question        string    `gorm:"type:text;not null" json:"question"`
	DurationMinutes int       `gorm:"not null;default:1440" json:"duration_minutes"` // 默认24小时
	ExpiresAt       time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt       time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt       time.Time `gorm:"not null" json:"updated_at"`
}

// PollOption 投票选项模型
type PollOption struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PollID    string    `gorm:"type:uuid;not null;index" json:"poll_id"`
	Text      string    `gorm:"type:varchar(255);not null" json:"text"`
	Position  int       `gorm:"not null" json:"position"`
	VoteCount int       `gorm:"default:0" json:"vote_count"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
}
