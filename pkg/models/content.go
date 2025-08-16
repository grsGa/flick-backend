package models

import (
	"time"
)

// Post 帖子模型
type Post struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     string    `gorm:"type:uuid;not null;index" json:"user_id"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	Visibility string    `gorm:"type:varchar(20);not null;check:visibility IN ('public', 'private', 'followers');default:'public'" json:"visibility"`
	ParentID   *string   `gorm:"type:uuid;index" json:"parent_id,omitempty"`
	RepostID   *string   `gorm:"type:uuid;index" json:"repost_id,omitempty"`
	CreatedAt  time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt  time.Time `gorm:"not null" json:"updated_at"`
	DeletedAt  *time.Time `gorm:"index" json:"deleted_at,omitempty"`
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
