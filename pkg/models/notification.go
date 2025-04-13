package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// NotificationType 表示通知类型
type NotificationType string

// 通知类型常量
const (
	NotificationLike      NotificationType = "like"
	NotificationComment   NotificationType = "comment"
	NotificationMention   NotificationType = "mention"
	NotificationFollow    NotificationType = "follow"
	NotificationSystem    NotificationType = "system"
	NotificationMessage   NotificationType = "message"
	NotificationShare     NotificationType = "share"
)

// NotificationStatus 表示通知状态
type NotificationStatus string

// 通知状态常量
const (
	NotificationStatusUnread  NotificationStatus = "unread"
	NotificationStatusRead    NotificationStatus = "read"
	NotificationStatusDeleted NotificationStatus = "deleted"
)

// NotificationData 通知数据，存储根据类型不同的附加信息
type NotificationData map[string]interface{}

// Value 实现driver.Valuer接口
func (n NotificationData) Value() (driver.Value, error) {
	if len(n) == 0 {
		return nil, nil
	}
	return json.Marshal(n)
}

// Scan 实现sql.Scanner接口
func (n *NotificationData) Scan(value interface{}) error {
	if value == nil {
		*n = NotificationData{}
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("类型断言为[]byte失败")
	}
	
	return json.Unmarshal(bytes, &n)
}

// Notification 通知模型
type Notification struct {
	ID          string             `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      string             `json:"user_id" gorm:"type:uuid;not null;index"`
	Type        NotificationType   `json:"type" gorm:"size:20;not null;index"`
	ActorID     string             `json:"actor_id" gorm:"type:uuid;index"` // 触发通知的用户ID
	TargetType  string             `json:"target_type" gorm:"size:20;index"` // post, comment, user, etc
	TargetID    string             `json:"target_id" gorm:"type:uuid;index"` // 目标对象ID
	Data        NotificationData   `json:"data" gorm:"type:jsonb"` // 附加数据
	Status      NotificationStatus `json:"status" gorm:"size:20;not null;index;default:'unread'"`
	CreatedAt   time.Time          `json:"created_at" gorm:"index"`
	UpdatedAt   time.Time          `json:"updated_at"`
	ReadAt      *time.Time         `json:"read_at"`
	DeletedAt   gorm.DeletedAt     `json:"-" gorm:"index"`
	User        User               `json:"-" gorm:"foreignKey:UserID"`
	Actor       User               `json:"actor,omitempty" gorm:"foreignKey:ActorID"`
}

// NotificationPreference 用户通知偏好设置
type NotificationPreference struct {
	ID                string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID            string         `json:"user_id" gorm:"type:uuid;not null;uniqueIndex"`
	LikeEnabled       bool           `json:"like_enabled" gorm:"default:true"`
	CommentEnabled    bool           `json:"comment_enabled" gorm:"default:true"`
	MentionEnabled    bool           `json:"mention_enabled" gorm:"default:true"`
	FollowEnabled     bool           `json:"follow_enabled" gorm:"default:true"`
	SystemEnabled     bool           `json:"system_enabled" gorm:"default:true"`
	MessageEnabled    bool           `json:"message_enabled" gorm:"default:true"`
	ShareEnabled      bool           `json:"share_enabled" gorm:"default:true"`
	EmailEnabled      bool           `json:"email_enabled" gorm:"default:true"`
	PushEnabled       bool           `json:"push_enabled" gorm:"default:true"`
	QuietHoursEnabled bool           `json:"quiet_hours_enabled" gorm:"default:false"`
	QuietHoursStart   string         `json:"quiet_hours_start" gorm:"size:5;default:'22:00'"` // 24小时制，格式: "HH:MM"
	QuietHoursEnd     string         `json:"quiet_hours_end" gorm:"size:5;default:'08:00'"`   // 24小时制，格式: "HH:MM"
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
	User              User           `json:"-" gorm:"foreignKey:UserID"`
}

// DeviceToken 用户设备令牌，用于推送通知
type DeviceToken struct {
	ID          string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      string         `json:"user_id" gorm:"type:uuid;not null;index"`
	DeviceType  string         `json:"device_type" gorm:"size:20;not null"` // ios, android, web
	Token       string         `json:"token" gorm:"size:500;not null;uniqueIndex"`
	AppVersion  string         `json:"app_version" gorm:"size:20"`
	LastUsed    time.Time      `json:"last_used"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	User        User           `json:"-" gorm:"foreignKey:UserID"`
}

// NotificationCount 用户未读通知计数
type NotificationCount struct {
	ID          string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      string    `json:"user_id" gorm:"type:uuid;not null;uniqueIndex"`
	UnreadCount int       `json:"unread_count" gorm:"default:0"`
	UpdatedAt   time.Time `json:"updated_at"`
	User        User      `json:"-" gorm:"foreignKey:UserID"`
}

// NotificationBatch 批量通知（用于发送给多个用户的相同通知）
type NotificationBatch struct {
	ID          string           `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Type        NotificationType `json:"type" gorm:"size:20;not null"`
	Title       string           `json:"title" gorm:"size:200"`
	Content     string           `json:"content" gorm:"type:text"`
	Data        NotificationData `json:"data" gorm:"type:jsonb"` // 附加数据
	TargetGroup string           `json:"target_group" gorm:"size:50"` // all, active_users, etc
	SentCount   int              `json:"sent_count" gorm:"default:0"`
	ReadCount   int              `json:"read_count" gorm:"default:0"`
	CreatedAt   time.Time        `json:"created_at"`
	SentAt      *time.Time       `json:"sent_at"`
}

// NotificationTemplate 通知模板
type NotificationTemplate struct {
	ID          string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Type        NotificationType `json:"type" gorm:"size:20;not null;uniqueIndex"`
	Title       string         `json:"title" gorm:"size:200"`
	Content     string         `json:"content" gorm:"type:text"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// UserDevice 用户设备信息
type UserDevice struct {
	ID                   string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID               string         `json:"user_id" gorm:"type:uuid;not null;index"`
	DeviceToken          string         `json:"device_token" gorm:"size:500;not null;uniqueIndex"`
	DeviceType           string         `json:"device_type" gorm:"size:20;not null"` // ios, android, web
	NotificationsEnabled bool           `json:"notifications_enabled" gorm:"default:true"`
	AppVersion           string         `json:"app_version" gorm:"size:50"`
	DeviceModel          string         `json:"device_model" gorm:"size:100"`
	OSVersion            string         `json:"os_version" gorm:"size:50"`
	LastActiveAt         time.Time      `json:"last_active_at"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `json:"-" gorm:"index"`
	User                 User           `json:"-" gorm:"foreignKey:UserID"`
} 