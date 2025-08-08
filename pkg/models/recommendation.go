package models

import (
	"time"
)

// FeedItem 推荐内容模型
type FeedItem struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    string    `gorm:"type:uuid;not null;index" json:"user_id"`
	PostID    string    `gorm:"type:uuid;not null;index" json:"post_id"`
	RankScore float64   `gorm:"type:decimal(10,6);not null;default:0.0" json:"rank_score"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}