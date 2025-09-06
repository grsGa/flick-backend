package models

import (
	"time"
)

// User 用户模型
type User struct {
	ID              string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Username        string     `gorm:"size:30;uniqueIndex;not null" json:"username"`
	DisplayName     string     `gorm:"size:50" json:"display_name"`
	Email           string     `gorm:"uniqueIndex;not null" json:"email"`
	Phone           *string    `gorm:"size:20;uniqueIndex" json:"phone,omitempty"`
	PasswordHash    string     `gorm:"type:text;not null" json:"password_hash"`
	AvatarURL       string     `gorm:"type:text" json:"avatar_url"`
	AvatarVersion   int        `gorm:"default:0" json:"avatar_version"`
	BannerURL       *string    `gorm:"type:text" json:"banner_url,omitempty"`
	Bio             *string    `gorm:"type:text" json:"bio,omitempty"`
	Location        *string    `gorm:"size:50" json:"location,omitempty"`
	WebsiteURL      *string    `gorm:"type:text" json:"website_url,omitempty"`
	FollowersCount  int        `gorm:"default:0" json:"followers_count"`
	FollowingCount  int        `gorm:"default:0" json:"following_count"`
	IsFollowing     bool       `gorm:"-" json:"is_following"` // Ignored by GORM, populated at runtime
	IsVerified      bool       `gorm:"default:false" json:"is_verified"`
	IsEmailVerified bool       `gorm:"default:false" json:"is_email_verified"`
	IsPhoneVerified bool       `gorm:"default:false" json:"is_phone_verified"`
	LoginMethod     string     `gorm:"type:varchar(20);not null;check:login_method IN ('password', 'google', 'apple', 'github')" json:"login_method"`
	LastLoginAt     *time.Time `gorm:"index" json:"last_login_at,omitempty"`
	Status          string     `gorm:"type:varchar(20);not null;default:'active';check:status IN ('active', 'suspended', 'banned')" json:"status"`
	CreatedAt       time.Time  `gorm:"not null" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"not null" json:"updated_at"`
	DeletedAt       *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}
