package repository

import (
	"context"
	content_proto "github.com/flick/backend/services/content/proto"
)

// PostRepository 定义帖子仓储接口
type PostRepository interface {
	// CreatePost 创建帖子
	CreatePost(ctx context.Context, req *content_proto.CreatePostRequest) (*content_proto.Post, error)
	
	// GetPost 根据ID获取帖子
	GetPost(ctx context.Context, postID, requestingUserID string) (*content_proto.Post, error)
	
	// GetUserPosts 获取用户帖子列表
	GetUserPosts(ctx context.Context, userID, requestingUserID string, limit int32, cursor string) ([]*content_proto.Post, string, bool, error)
	
	// GetTimeline 获取时间线
	GetTimeline(ctx context.Context, userID string, limit int32, cursor string) ([]*content_proto.Post, string, bool, error)
	
	// DeletePost 删除帖子
	DeletePost(ctx context.Context, postID, userID string) error
	
	// CheckReplyPermission 检查回复权限
	CheckReplyPermission(ctx context.Context, postID, userID string) (bool, string, error)
}

// MentionRepository 定义提及仓储接口
type MentionRepository interface {
	// CreateMentions 创建提及记录
	CreateMentions(ctx context.Context, postID string, mentionedUserIDs []string) error
	
	// GetMentionedUsers 获取帖子中被提及的用户
	GetMentionedUsers(ctx context.Context, postID string) ([]string, error)
	
	// CheckUserMentioned 检查用户是否被提及
	CheckUserMentioned(ctx context.Context, postID, userID string) (bool, error)
}

// TagRepository 定义标签仓储接口
type TagRepository interface {
	// CreateTags 创建标签记录
	CreateTags(ctx context.Context, postID string, tags []string) error
	
	// GetPostTags 获取帖子标签
	GetPostTags(ctx context.Context, postID string) ([]string, error)
}

// PollRepository 定义投票仓储接口
type PollRepository interface {
	// CreatePoll 创建投票
	CreatePoll(ctx context.Context, postID string, pollData *content_proto.PollData) (*content_proto.Poll, error)
	
	// GetPoll 获取投票信息
	GetPoll(ctx context.Context, pollID string) (*content_proto.Poll, error)
	
	// GetPostPoll 根据帖子ID获取投票
	GetPostPoll(ctx context.Context, postID string) (*content_proto.Poll, error)
}

// ContentRepository 定义内容仓储接口（兼容性）
type ContentRepository interface {
	// CreateContent 创建内容
	CreateContent(ctx context.Context, content *content_proto.Post) error
	
	// GetContent 获取内容
	GetContent(ctx context.Context, contentID string) (*content_proto.Post, error)
	
	// UpdateContent 更新内容
	UpdateContent(ctx context.Context, content *content_proto.Post) error
	
	// DeleteContent 删除内容
	DeleteContent(ctx context.Context, contentID string) error
	
	// ListContent 列出内容
	ListContent(ctx context.Context, userID string, limit, offset int32) ([]*content_proto.Post, int32, error)
}
