package models

import (
	"time"

	"gorm.io/gorm"
)

// UserMedia 用户媒体文件模型
type UserMedia struct {
	ID          string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      string         `json:"user_id" gorm:"type:uuid;not null;index"`
	MediaType   string         `json:"media_type" gorm:"size:20;not null"` // avatar, cover_image, attachment
	URL         string         `json:"url" gorm:"not null"`
	FileName    string         `json:"file_name"`
	FileSize    int64          `json:"file_size"`
	ContentType string         `json:"content_type"`
	ObjectName  string         `json:"object_name"` // 存储桶中的对象名称
	Active      bool           `json:"active" gorm:"default:true"` // 是否当前使用中
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	User        User           `json:"-" gorm:"foreignKey:UserID"`
} 