package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// PostType 表示帖子类型
type PostType string

// PostStatus 表示帖子状态
type PostStatus string

// 帖子类型常量
const (
	PostTypeText    PostType = "text"
	PostTypeImage   PostType = "image"
	PostTypeVideo   PostType = "video"
	PostTypeLink    PostType = "link"
	PostTypePoll    PostType = "poll"
	PostTypeArticle PostType = "article"
)

// 帖子状态常量
const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusPublished PostStatus = "published"
	PostStatusArchived  PostStatus = "archived"
	PostStatusDeleted   PostStatus = "deleted"
	PostStatusReported  PostStatus = "reported"
	PostStatusHidden    PostStatus = "hidden"
	PostStatusFlagged   PostStatus = "flagged" // 被标记为不当内容的帖子
)

// MediaFile 表示媒体文件
type MediaFile struct {
	URL          string `json:"url"`
	Type         string `json:"type"`          // image, video
	MimeType     string `json:"mime_type"`     // image/jpeg, video/mp4
	Size         int64  `json:"size"`          // 字节大小
	Width        int    `json:"width"`         // 宽度（像素）
	Height       int    `json:"height"`        // 高度（像素）
	Duration     int    `json:"duration"`      // 时长（秒，仅视频）
	Description  string `json:"description"`   // 描述/替代文本
	ThumbnailURL string `json:"thumbnail_url"` // 缩略图URL（仅视频）
}

// MediaFiles JSON类型，用于存储多个媒体文件
type MediaFiles []MediaFile

// Value 实现driver.Valuer接口
func (m MediaFiles) Value() (driver.Value, error) {
	if len(m) == 0 {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan 实现sql.Scanner接口
func (m *MediaFiles) Scan(value interface{}) error {
	if value == nil {
		*m = MediaFiles{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("类型断言为[]byte失败")
	}

	return json.Unmarshal(bytes, &m)
}

// PollOption 表示投票选项
type PollOption struct {
	ID    string `json:"id"`
	Text  string `json:"text"`
	Count int    `json:"count"`
}

// PollOptions JSON类型，用于存储投票选项
type PollOptions []PollOption

// Value 实现driver.Valuer接口
func (p PollOptions) Value() (driver.Value, error) {
	if len(p) == 0 {
		return nil, nil
	}
	return json.Marshal(p)
}

// Scan 实现sql.Scanner接口
func (p *PollOptions) Scan(value interface{}) error {
	if value == nil {
		*p = PollOptions{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("类型断言为[]byte失败")
	}

	return json.Unmarshal(bytes, &p)
}

// Post 帖子数据模型
type Post struct {
	ID            string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PermalinkID   string         `json:"permalink_id" gorm:"type:varchar(20);index;not null;default:gen_random_permalink()"` // 数字性永久链接ID，类似Twitter的推文ID
	UserID        string         `json:"user_id" gorm:"type:uuid;not null;index"`
	Type          PostType       `json:"type" gorm:"size:20;not null;index"`
	Title         string         `json:"title" gorm:"size:200"`
	Content       string         `json:"content" gorm:"type:text"`
	ContentHTML   string         `json:"content_html" gorm:"type:text"`
	Summary       string         `json:"summary" gorm:"size:500"`
	Status        PostStatus     `json:"status" gorm:"size:20;not null;index;default:'published'"`
	MediaFiles    MediaFiles     `json:"media_files" gorm:"type:jsonb"`
	PollOptions   PollOptions    `json:"poll_options" gorm:"type:jsonb"`
	PollEndsAt    *time.Time     `json:"poll_ends_at"`
	LinkURL       string         `json:"link_url" gorm:"size:500"`
	Tags          []Tag          `json:"tags" gorm:"many2many:post_tags;"`
	Categories    []Category     `json:"categories" gorm:"many2many:post_categories;"`
	PrivacyLevel  string         `json:"privacy_level" gorm:"size:20;default:'public'"` // public, followers, private
	AllowComments bool           `json:"allow_comments" gorm:"default:true"`
	ViewCount     int            `json:"view_count" gorm:"default:0"`
	LikeCount     int            `json:"like_count" gorm:"default:0"`
	CommentCount  int            `json:"comment_count" gorm:"default:0"`
	ShareCount    int            `json:"share_count" gorm:"default:0"`
	BookmarkCount int            `json:"bookmark_count" gorm:"default:0"` // 收藏数量
	FeaturedAt    *time.Time     `json:"featured_at"`
	PublishedAt   *time.Time     `json:"published_at"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
	User          User           `json:"user" gorm:"foreignKey:UserID"`
	IsLiked       *bool          `json:"is_liked,omitempty" gorm:"-"` // 当前用户是否已点赞，非持久化字段
	IsSaved       *bool          `json:"is_saved,omitempty" gorm:"-"` // 当前用户是否已收藏，非持久化字段
}

// Tag 标签数据模型
type Tag struct {
	ID          string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"uniqueIndex;size:50;not null"`
	Slug        string    `json:"slug" gorm:"uniqueIndex;size:50;not null"`
	Description string    `json:"description" gorm:"size:500"`
	PostCount   int       `json:"post_count" gorm:"default:0"`
	Featured    bool      `json:"featured" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Posts       []Post    `json:"-" gorm:"many2many:post_tags;"`
}

// Category 分类数据模型
type Category struct {
	ID           string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name         string         `json:"name" gorm:"uniqueIndex;size:50;not null"`
	Slug         string         `json:"slug" gorm:"uniqueIndex;size:50;not null"`
	Description  string         `json:"description" gorm:"size:500"`
	ParentID     *string        `json:"parent_id" gorm:"type:uuid;index"`
	IconURL      string         `json:"icon_url"`
	PostCount    int            `json:"post_count" gorm:"default:0"`
	DisplayOrder int            `json:"display_order" gorm:"default:0"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
	Posts        []Post         `json:"-" gorm:"many2many:post_categories;"`
	Parent       *Category      `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children     []Category     `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}

// PostAuditLog 帖子审计日志
type PostAuditLog struct {
	ID         string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PostID     string    `json:"post_id" gorm:"type:uuid;not null;index"`
	UserID     string    `json:"user_id" gorm:"type:uuid;index"`      // 操作者ID
	ActionType string    `json:"action_type" gorm:"size:50;not null"` // create, update, delete, publish, etc
	OldValue   string    `json:"old_value" gorm:"type:text"`
	NewValue   string    `json:"new_value" gorm:"type:text"`
	IPAddress  string    `json:"-"`
	UserAgent  string    `json:"-"`
	CreatedAt  time.Time `json:"created_at"`
	Post       Post      `json:"-" gorm:"foreignKey:PostID"`
	User       User      `json:"-" gorm:"foreignKey:UserID"`
}

// PostReport 帖子举报
type PostReport struct {
	ID          string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PostID      string     `json:"post_id" gorm:"type:uuid;not null;index"`
	ReporterID  string     `json:"reporter_id" gorm:"type:uuid;not null;index"`
	ReasonCode  string     `json:"reason_code" gorm:"size:50;not null"` // spam, violence, hate, etc
	Description string     `json:"description" gorm:"type:text"`
	Status      string     `json:"status" gorm:"size:20;default:'pending'"` // pending, approved, rejected
	ReviewerID  string     `json:"reviewer_id" gorm:"type:uuid;index"`
	ReviewNote  string     `json:"review_note" gorm:"type:text"`
	ReviewedAt  *time.Time `json:"reviewed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Post        Post       `json:"-" gorm:"foreignKey:PostID"`
	Reporter    User       `json:"-" gorm:"foreignKey:ReporterID"`
	Reviewer    *User      `json:"-" gorm:"foreignKey:ReviewerID"`
}

// SavedPost 用户保存的帖子
type SavedPost struct {
	ID         string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     string    `json:"user_id" gorm:"type:uuid;not null;index"`
	PostID     string    `json:"post_id" gorm:"type:uuid;not null;index"`
	Collection string    `json:"collection" gorm:"size:50;default:'default'"` // 允许用户创建多个收藏夹
	CreatedAt  time.Time `json:"created_at"`
	User       User      `json:"-" gorm:"foreignKey:UserID"`
	Post       Post      `json:"-" gorm:"foreignKey:PostID"`
}
