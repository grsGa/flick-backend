package repository

import (
	"context"
	"backend/services/interaction/proto"
)

// InteractionRepository 定义互动仓储接口
type InteractionRepository interface {
	// CreateFollow 创建关注
	CreateFollow(ctx context.Context, follow *proto.Follow) error
	
	// DeleteFollow 删除关注
	DeleteFollow(ctx context.Context, followerID, followeeID string) error
	
	// IsFollowing 检查是否关注
	IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error)
	
	// GetFollowers 获取粉丝列表
	GetFollowers(ctx context.Context, userID string, page, pageSize int32) ([]*proto.Follow, int32, error)
	
	// GetFollowing 获取关注列表
	GetFollowing(ctx context.Context, userID string, page, pageSize int32) ([]*proto.Follow, int32, error)
	
	// CreateLike 创建点赞
	CreateLike(ctx context.Context, like *proto.Like) error
	
	// DeleteLike 删除点赞
	DeleteLike(ctx context.Context, userID, postID string) error
	
	// IsLiked 检查是否点赞
	IsLiked(ctx context.Context, userID, postID string) (bool, error)
	
	// GetLikes 获取点赞列表
	GetLikes(ctx context.Context, postID string, page, pageSize int32) ([]*proto.Like, int32, error)
	
	// CreateRepost 创建转发
	CreateRepost(ctx context.Context, repost *proto.Repost) error
	
	// DeleteRepost 删除转发
	DeleteRepost(ctx context.Context, userID, postID string) error
	
	// CreateReport 创建举报
	CreateReport(ctx context.Context, report *proto.Report) error
}