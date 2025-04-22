package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID              string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Username        string         `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Email           string         `json:"email" gorm:"uniqueIndex;not null"`
	Password        string         `json:"-" gorm:"not null"`                       // 不返回密码
	PhoneNumber     string         `json:"phone_number" gorm:"uniqueIndex;size:20"` // 对手机号码添加唯一索引
	DisplayName     string         `json:"display_name" gorm:"size:100"`
	Bio             string         `json:"bio" gorm:"size:500"`
	Location        string         `json:"location" gorm:"size:100"`
	Website         string         `json:"website" gorm:"size:200"`
	AvatarURL       string         `json:"avatar_url"`
	CoverImageURL   string         `json:"cover_image_url"`
	ProfileComplete bool           `json:"profile_complete" gorm:"default:false"`
	VerifiedEmail   bool           `json:"verified_email" gorm:"default:false"`
	VerifiedPhone   bool           `json:"verified_phone" gorm:"default:false"`
	AccountStatus   string         `json:"account_status" gorm:"default:'active';size:20"` // active, suspended, deleted
	LastLogin       *time.Time     `json:"last_login"`
	LastIPAddress   string         `json:"-"` // 不返回IP地址
	UserAgent       string         `json:"-"` // 不返回用户代理
	Roles           []Role         `json:"roles" gorm:"many2many:flick_user_roles;"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

// UserCredential 存储额外验证凭据
type UserCredential struct {
	ID          string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      string         `json:"user_id" gorm:"type:uuid;not null;index"`
	Type        string         `json:"type" gorm:"size:20;not null"` // 类型：2fa, oauth, api_key
	Provider    string         `json:"provider" gorm:"size:20"`      // 提供者：google, github, etc
	Key         string         `json:"-"`                            // 密钥/令牌
	Description string         `json:"description"`
	LastUsed    *time.Time     `json:"last_used"`
	ExpiresAt   *time.Time     `json:"expires_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	User        User           `json:"-" gorm:"foreignKey:UserID"`
}

// UserSession 存储用户会话信息
type UserSession struct {
	ID           string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       string         `json:"user_id" gorm:"type:uuid;not null;index"`
	RefreshToken string         `json:"-"`
	IPAddress    string         `json:"-"`
	UserAgent    string         `json:"-"`
	Device       string         `json:"device"`
	LastActivity time.Time      `json:"last_activity"`
	ExpiresAt    time.Time      `json:"expires_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
	User         User           `json:"-" gorm:"foreignKey:UserID"`
}

// Role 角色模型
type Role struct {
	ID          string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string         `json:"name" gorm:"uniqueIndex;size:50;not null"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	Users       []User         `json:"-" gorm:"many2many:flick_user_roles;"`
}

// UserVerification 用户验证（邮箱、手机验证码等）
type UserVerification struct {
	ID        string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string         `json:"user_id" gorm:"type:uuid;not null;index"`
	Type      string         `json:"type" gorm:"size:20;not null"` // email, phone, password_reset
	Token     string         `json:"-"`                            // 验证码/令牌
	Used      bool           `json:"used" gorm:"default:false"`
	ExpiresAt time.Time      `json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	User      User           `json:"-" gorm:"foreignKey:UserID"`
}

// UserActivity 用户活动日志
type UserActivity struct {
	ID          string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      string    `json:"user_id" gorm:"type:uuid;not null;index"`
	ActionType  string    `json:"action_type" gorm:"size:50;not null"` // login, profile_update, etc
	IPAddress   string    `json:"-"`
	UserAgent   string    `json:"-"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	User        User      `json:"-" gorm:"foreignKey:UserID"`
}

// SetPassword 设置用户密码（hash后存储）
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword 验证用户密码
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// UserFollow 用户关注关系
type UserFollow struct {
	ID          string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FollowerID  string         `json:"follower_id" gorm:"type:uuid;not null;index"`
	FollowingID string         `json:"following_id" gorm:"type:uuid;not null;index"`
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	Follower    User           `json:"-" gorm:"foreignKey:FollowerID"`
	Following   User           `json:"-" gorm:"foreignKey:FollowingID"`
}

// UserBlock 用户屏蔽关系
type UserBlock struct {
	ID        string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	BlockerID string         `json:"blocker_id" gorm:"type:uuid;not null;index"`
	BlockedID string         `json:"blocked_id" gorm:"type:uuid;not null;index"`
	Reason    string         `json:"reason" gorm:"size:200"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	Blocker   User           `json:"-" gorm:"foreignKey:BlockerID"`
	Blocked   User           `json:"-" gorm:"foreignKey:BlockedID"`
}

// UserStats 用户统计信息
type UserStats struct {
	ID             string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID         string         `json:"user_id" gorm:"type:uuid;not null;uniqueIndex"`
	PostCount      int            `json:"post_count" gorm:"default:0"`
	FollowerCount  int            `json:"follower_count" gorm:"default:0"`
	FollowingCount int            `json:"following_count" gorm:"default:0"`
	LikeCount      int            `json:"like_count" gorm:"default:0"`
	CommentCount   int            `json:"comment_count" gorm:"default:0"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
	User           User           `json:"-" gorm:"foreignKey:UserID"`
}
