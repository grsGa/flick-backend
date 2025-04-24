package models

import (
	"time"

	"gorm.io/gorm"
)

// ContentMedia 内容媒体文件模型（用于内容服务的图片和视频存储）
type ContentMedia struct {
	ID          string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      string         `json:"user_id" gorm:"type:uuid;not null;index"`
	MediaType   string         `json:"media_type" gorm:"size:20;not null"` // image, video
	URL         string         `json:"url" gorm:"not null"`
	FileName    string         `json:"file_name"`
	FileSize    int64          `json:"file_size"`
	ContentType string         `json:"content_type"`
	ObjectName  string         `json:"object_name"` // 存储桶中的对象名称
	Width       int            `json:"width" gorm:"default:0"`
	Height      int            `json:"height" gorm:"default:0"`
	Duration    int            `json:"duration" gorm:"default:0"` // 仅视频，单位秒
	Description string         `json:"description"` // 媒体描述
	PostID      string         `json:"post_id" gorm:"type:uuid;index"` // 关联的帖子ID，可为空表示尚未关联到帖子
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
} 