package service

import (
	"context"
	"github.com/flick/backend/services/content/proto"
)

// PostService 定义帖子服务接口
type PostService interface {
	// CreatePost 创建帖子
	CreatePost(ctx context.Context, req *proto.CreatePostRequest) (*proto.CreatePostResponse, error)
	
	// GetPost 获取帖子
	GetPost(ctx context.Context, req *proto.GetPostRequest) (*proto.GetPostResponse, error)
	
	// GetUserPosts 获取用户帖子列表
	GetUserPosts(ctx context.Context, req *proto.GetUserPostsRequest) (*proto.GetUserPostsResponse, error)
	
	// GetTimeline 获取时间线
	GetTimeline(ctx context.Context, req *proto.GetTimelineRequest) (*proto.GetTimelineResponse, error)
	
	// DeletePost 删除帖子
	DeletePost(ctx context.Context, req *proto.DeletePostRequest) (*proto.DeletePostResponse, error)
	
	// CheckReplyPermission 检查回复权限
	CheckReplyPermission(ctx context.Context, req *proto.CheckReplyPermissionRequest) (*proto.CheckReplyPermissionResponse, error)
}

// ContentService 定义内容服务接口（保持兼容性）
type ContentService interface {
	// GetContent 获取内容
	GetContent(ctx context.Context, req *proto.GetPostRequest) (*proto.GetPostResponse, error)
	
	// CreateContent 创建内容
	CreateContent(ctx context.Context, req *proto.CreatePostRequest) (*proto.CreatePostResponse, error)
	
	// UpdateContent 更新内容
	UpdateContent(ctx context.Context, req *proto.CreatePostRequest) (*proto.CreatePostResponse, error)
	
	// DeleteContent 删除内容
	DeleteContent(ctx context.Context, req *proto.DeletePostRequest) (*proto.DeletePostResponse, error)
	
	// ListContent 列出内容
	ListContent(ctx context.Context, req *proto.GetUserPostsRequest) (*proto.GetUserPostsResponse, error)
}
