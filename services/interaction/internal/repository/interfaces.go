package repository

import (
	"context"

	"github.com/flick/backend/services/interaction/proto"
)

// InteractionRepository 定义互动仓储接口
type InteractionRepository interface {
	// CreateFollow 创建关注
	CreateFollow(ctx context.Context, follow *proto.Follow) error

	// DeleteFollow 删除关注
	DeleteFollow(ctx context.Context, followerID, followeeID string) error

	// IsFollowing 检查是否关注
	IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error)

	// GetFollow 获取关注记录
	GetFollow(ctx context.Context, followerID, followeeID string) (*proto.Follow, error)

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

	// CreateComment 创建评论
	CreateComment(ctx context.Context, comment *proto.Comment) error

	// DeleteComment 删除评论
	DeleteComment(ctx context.Context, commentID, userID string) error

	// GetComments 获取评论列表
	GetComments(ctx context.Context, postID string, limit int32, cursor string) ([]*proto.Comment, string, bool, error)

	// GetPostStats 获取帖子统计
	GetPostStats(ctx context.Context, postID string) (*proto.PostStats, error)

	// UpdatePostStats 更新帖子统计
	UpdatePostStats(ctx context.Context, postID, action string, delta int32) (*proto.PostStats, error)

	// VotePoll 投票
	VotePoll(ctx context.Context, vote *proto.PollVote) error

	// CreateBookmark 创建收藏
	CreateBookmark(ctx context.Context, bookmark *proto.Bookmark) error

	// DeleteBookmark 删除收藏
	DeleteBookmark(ctx context.Context, userID, postID string) error

	// IsBookmarked 检查是否收藏
	IsBookmarked(ctx context.Context, userID, postID string) (bool, error)

	// GetBookmarks 获取收藏列表
	GetBookmarks(ctx context.Context, userID string, limit int32, cursor string) ([]*proto.Bookmark, string, bool, error)

	// CreateReport 创建举报
	CreateReport(ctx context.Context, report *proto.Report) error
}
