package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
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
	RootID          *string   `gorm:"type:uuid;index" json:"root_id,omitempty"` // 根帖子ID，用于两层回复结构
	RepostID        *string   `gorm:"type:uuid;index" json:"repost_id,omitempty"`
	IsReply         bool      `gorm:"default:false;index" json:"is_reply"` // 是否为回复
	ReplyLevel      int       `gorm:"default:0;index" json:"reply_level"` // 回复层级：0=原帖，1=顶级回复，2=次级回复
	HasMedia        bool      `gorm:"default:false" json:"has_media"`
	HasPoll         bool      `gorm:"default:false" json:"has_poll"`
	CreatedAt       time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt       time.Time `gorm:"not null" json:"updated_at"`
	DeletedAt       *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

// MediaVariant 媒体版本信息
type MediaVariant struct {
	URL      string  `json:"url"`      // 版本URL
	Width    int32   `json:"width"`    // 宽度
	Height   int32   `json:"height"`   // 高度
	Size     int64   `json:"size"`     // 文件大小
	Duration float64 `json:"duration,omitempty"` // 视频时长(秒)，仅视频文件使用
}

// MediaVariants 多版本媒体信息
type MediaVariants struct {
	Thumbnail *MediaVariant `json:"thumbnail,omitempty"` // 缩略图 (150px)
	Small     *MediaVariant `json:"small,omitempty"`     // 小图 (300px)
	Medium    *MediaVariant `json:"medium,omitempty"`    // 中图 (600px)
	Large     *MediaVariant `json:"large,omitempty"`     // 大图 (1200px)
	Original  *MediaVariant `json:"original,omitempty"`  // 原图
	// 视频特有
	Preview   *MediaVariant `json:"preview,omitempty"`   // 视频预览图
	LowRes    *MediaVariant `json:"low_res,omitempty"`   // 低分辨率视频 (240p)
	MidRes    *MediaVariant `json:"mid_res,omitempty"`   // 中分辨率视频 (480p)
	HighRes   *MediaVariant `json:"high_res,omitempty"`  // 高分辨率视频 (720p)
}

// Value 实现 driver.Valuer 接口，用于将 MediaVariants 转换为数据库值
func (mv MediaVariants) Value() (driver.Value, error) {
	if mv == (MediaVariants{}) {
		return nil, nil
	}
	return json.Marshal(mv)
}

// Scan 实现 sql.Scanner 接口，用于从数据库值扫描到 MediaVariants
func (mv *MediaVariants) Scan(value interface{}) error {
	if value == nil {
		*mv = MediaVariants{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot scan value into MediaVariants")
	}

	return json.Unmarshal(bytes, mv)
}

// MediaAttachment 媒体附件模型
type MediaAttachment struct {
	ID           string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PostID       *string        `gorm:"type:uuid;index" json:"post_id,omitempty"` // 修改为可选，支持头像/横幅等独立文件
	UserID       string         `gorm:"type:uuid;not null;index" json:"user_id"` // 添加用户ID字段
	Filename     string         `gorm:"type:varchar(255);not null" json:"filename"`
	URL          string         `gorm:"type:text;not null" json:"url"` // 保留原始URL字段向后兼容
	Type         string         `gorm:"type:varchar(20);not null" json:"type"` // image, video, gif
	MimeType     string         `gorm:"type:varchar(100);not null" json:"mime_type"` // image/jpeg, video/mp4等
	Size         int64          `gorm:"not null;default:0" json:"size"`
	Status       string         `gorm:"type:varchar(20);not null;default:'pending'" json:"status"` // pending, processing, ready, failed
	Width        int32          `gorm:"default:0" json:"width"`
	Height       int32          `gorm:"default:0" json:"height"`
	Duration     int32          `gorm:"default:0" json:"duration"` // 视频时长(秒)
	ThumbnailURL *string        `gorm:"type:text" json:"thumbnail_url,omitempty"` // 保留向后兼容
	Variants     MediaVariants  `gorm:"type:jsonb" json:"variants"` // 多版本URL存储
	AltText      *string        `gorm:"type:text" json:"alt_text,omitempty"`
	ProcessedAt  *time.Time     `gorm:"index" json:"processed_at,omitempty"` // 处理完成时间
	CreatedAt    time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt    *time.Time     `gorm:"index" json:"deleted_at,omitempty"`
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

// ReplyMention 回复提及模型 - 用于次级回复中的"回复@某人"功能
type ReplyMention struct {
	ID             string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ReplyID        string    `gorm:"type:uuid;not null;index" json:"reply_id"` // 回复帖子ID
	MentionedUserID string    `gorm:"type:uuid;not null;index" json:"mentioned_user_id"` // 被回复的用户ID
	CreatedAt      time.Time `gorm:"not null" json:"created_at"`
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
