package models

import (
	//"database/sql/driver"
	//"encoding/json"
	//"errors"
	"time"

	"gorm.io/gorm"
)

// ReactionType 表示互动反应类型
type ReactionType string

// 反应类型常量
const (
	ReactionLike  ReactionType = "like"
	ReactionLove  ReactionType = "love"
	ReactionHaha  ReactionType = "haha"
	ReactionWow   ReactionType = "wow"
	ReactionSad   ReactionType = "sad"
	ReactionAngry ReactionType = "angry"
)

// Reaction 互动反应模型
type Reaction struct {
	ID         string       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     string       `json:"user_id" gorm:"type:uuid;not null;index"`
	TargetType string       `json:"target_type" gorm:"size:20;not null;index"` // post, comment
	TargetID   string       `json:"target_id" gorm:"type:uuid;not null;index"`
	Type       ReactionType `json:"type" gorm:"size:20;not null"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
	User       User         `json:"user" gorm:"foreignKey:UserID"`
}

// Comment 评论模型
type Comment struct {
	ID          string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      string         `json:"user_id" gorm:"type:uuid;not null;index"`
	PostID      string         `json:"post_id" gorm:"type:uuid;not null;index"`
	ParentID    *string        `json:"parent_id" gorm:"type:uuid;index"`
	Content     string         `json:"content" gorm:"type:text;not null"`
	ContentHTML string         `json:"content_html" gorm:"type:text"`
	Status      string         `json:"status" gorm:"size:20;not null;default:'active'"` // active, hidden, deleted
	LikeCount   int            `json:"like_count" gorm:"default:0"`
	ReplyCount  int            `json:"reply_count" gorm:"default:0"`
	IPAddress   string         `json:"-"`
	MediaFiles  MediaFiles     `json:"media_files" gorm:"type:jsonb"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	User        User           `json:"user" gorm:"foreignKey:UserID"`
	Post        Post           `json:"-" gorm:"foreignKey:PostID"`
	Parent      *Comment       `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Replies     []Comment      `json:"replies,omitempty" gorm:"foreignKey:ParentID"`
}

// CommentReport 评论举报
type CommentReport struct {
	ID          string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CommentID   string     `json:"comment_id" gorm:"type:uuid;not null;index"`
	ReporterID  string     `json:"reporter_id" gorm:"type:uuid;not null;index"`
	ReasonCode  string     `json:"reason_code" gorm:"size:50;not null"` // spam, violence, hate, etc
	Description string     `json:"description" gorm:"type:text"`
	Status      string     `json:"status" gorm:"size:20;default:'pending'"` // pending, approved, rejected
	ReviewerID  string     `json:"reviewer_id" gorm:"type:uuid;index"`
	ReviewNote  string     `json:"review_note" gorm:"type:text"`
	ReviewedAt  *time.Time `json:"reviewed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Comment     Comment    `json:"-" gorm:"foreignKey:CommentID"`
	Reporter    User       `json:"-" gorm:"foreignKey:ReporterID"`
	Reviewer    *User      `json:"-" gorm:"foreignKey:ReviewerID"`
}

// MentionType 表示提及的类型
type MentionType string

// 提及类型常量
const (
	MentionUser    MentionType = "user"
	MentionHashtag MentionType = "hashtag"
)

// Mention 提及模型
type Mention struct {
	ID         string      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Type       MentionType `json:"type" gorm:"size:20;not null"`
	TargetID   string      `json:"target_id" gorm:"type:uuid;index"`           // 被提及的对象ID（用户ID或标签ID）
	SourceType string      `json:"source_type" gorm:"size:20;not null;index"`  // post, comment
	SourceID   string      `json:"source_id" gorm:"type:uuid;not null;index"`  // 提及来源ID
	CreatorID  string      `json:"creator_id" gorm:"type:uuid;not null;index"` // 创建提及的用户
	CreatedAt  time.Time   `json:"created_at"`
	Creator    User        `json:"-" gorm:"foreignKey:CreatorID"`
}

// PollVote 投票记录
type PollVote struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string    `json:"user_id" gorm:"type:uuid;not null;index"`
	PostID    string    `json:"post_id" gorm:"type:uuid;not null;index"`
	OptionID  string    `json:"option_id" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	User      User      `json:"-" gorm:"foreignKey:UserID"`
	Post      Post      `json:"-" gorm:"foreignKey:PostID"`
}

// Share 分享记录
type Share struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string    `json:"user_id" gorm:"type:uuid;not null;index"`
	PostID    string    `json:"post_id" gorm:"type:uuid;not null;index"`
	Platform  string    `json:"platform" gorm:"size:50"` // internal, twitter, facebook, etc
	Note      string    `json:"note" gorm:"type:text"`   // 分享时添加的备注
	CreatedAt time.Time `json:"created_at"`
	User      User      `json:"-" gorm:"foreignKey:UserID"`
	Post      Post      `json:"-" gorm:"foreignKey:PostID"`
}

// UserInteraction 用户互动统计
type UserInteraction struct {
	ID                   string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID               string    `json:"user_id" gorm:"type:uuid;not null;uniqueIndex"`
	PostCount            int       `json:"post_count" gorm:"default:0"`
	CommentCount         int       `json:"comment_count" gorm:"default:0"`
	ReactionCount        int       `json:"reaction_count" gorm:"default:0"`
	ReceivedLikeCount    int       `json:"received_like_count" gorm:"default:0"`
	ReceivedCommentCount int       `json:"received_comment_count" gorm:"default:0"`
	ShareCount           int       `json:"share_count" gorm:"default:0"`
	UpdatedAt            time.Time `json:"updated_at"`
	User                 User      `json:"-" gorm:"foreignKey:UserID"`
}

// PostLike 帖子点赞模型
type PostLike struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string    `json:"user_id" gorm:"type:uuid;not null;index"`
	PostID    string    `json:"post_id" gorm:"type:uuid;not null;index"`
	CreatedAt time.Time `json:"created_at"`
	User      User      `json:"-" gorm:"foreignKey:UserID"`
	Post      Post      `json:"-" gorm:"foreignKey:PostID"`
}

// CommentLike 评论点赞模型
type CommentLike struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string    `json:"user_id" gorm:"type:uuid;not null;index"`
	CommentID string    `json:"comment_id" gorm:"type:uuid;not null;index"`
	CreatedAt time.Time `json:"created_at"`
	User      User      `json:"-" gorm:"foreignKey:UserID"`
	Comment   Comment   `json:"-" gorm:"foreignKey:CommentID"`
}

// CommentReply 评论回复模型
type CommentReply struct {
	ID        string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string         `json:"user_id" gorm:"type:uuid;not null;index"`
	CommentID string         `json:"comment_id" gorm:"type:uuid;not null;index"`
	Content   string         `json:"content" gorm:"type:text;not null"`
	Status    string         `json:"status" gorm:"size:20;not null;default:'active'"` // active, hidden, deleted
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	User      User           `json:"user" gorm:"foreignKey:UserID"`
	Comment   Comment        `json:"-" gorm:"foreignKey:CommentID"`
}

// PostShare 帖子分享模型
type PostShare struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string    `json:"user_id" gorm:"type:uuid;not null;index"`
	PostID    string    `json:"post_id" gorm:"type:uuid;not null;index"`
	Platform  string    `json:"platform" gorm:"size:50"` // internal, twitter, facebook, etc
	Note      string    `json:"note" gorm:"type:text"`   // 分享时添加的备注
	CreatedAt time.Time `json:"created_at"`
	User      User      `json:"-" gorm:"foreignKey:UserID"`
	Post      Post      `json:"-" gorm:"foreignKey:PostID"`
}

// Bookmark 书签/收藏模型
type Bookmark struct {
	ID             string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID         string    `json:"user_id" gorm:"type:uuid;not null;index"`
	PostID         string    `json:"post_id" gorm:"type:uuid;not null;index"`
	CollectionName string    `json:"collection_name" gorm:"size:100;default:'default'"`
	CreatedAt      time.Time `json:"created_at"`
	User           User      `json:"-" gorm:"foreignKey:UserID"`
	Post           Post      `json:"-" gorm:"foreignKey:PostID"`
}

// BookmarkCollection 书签集合模型
type BookmarkCollection struct {
	ID          string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      string    `json:"user_id" gorm:"type:uuid;not null;index"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Description string    `json:"description" gorm:"type:text"`
	IsPrivate   bool      `json:"is_private" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	User        User      `json:"-" gorm:"foreignKey:UserID"`
}

// InteractionHistory 互动历史记录
type InteractionHistory struct {
	ID         string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     string    `json:"user_id" gorm:"type:uuid;not null;index"`
	ObjectID   string    `json:"object_id" gorm:"type:uuid;not null;index"` // 被互动的对象ID（帖子、评论等）
	ObjectType string    `json:"object_type" gorm:"size:20;not null;index"` // post, comment, user
	ActionType string    `json:"action_type" gorm:"size:30;not null;index"` // like, comment, share, etc
	DetailID   *string   `json:"detail_id" gorm:"type:uuid;index"`          // 详情ID（如评论ID）
	CreatedAt  time.Time `json:"created_at"`
	User       User      `json:"-" gorm:"foreignKey:UserID"`
}

// InteractionStats 帖子互动统计
type InteractionStats struct {
	PostID        string `json:"post_id"`
	LikeCount     int64  `json:"like_count"`
	CommentCount  int64  `json:"comment_count"`
	ShareCount    int64  `json:"share_count"`
	BookmarkCount int64  `json:"bookmark_count"`
	ViewCount     int64  `json:"view_count"`
}

// UserInteractionStats 用户互动统计
type UserInteractionStats struct {
	UserID               string `json:"user_id"`
	PostsLikedCount      int64  `json:"posts_liked_count"`
	CommentsCount        int64  `json:"comments_count"`
	CommentsLikedCount   int64  `json:"comments_liked_count"`
	PostsSharedCount     int64  `json:"posts_shared_count"`
	PostsBookmarkedCount int64  `json:"posts_bookmarked_count"`
}
