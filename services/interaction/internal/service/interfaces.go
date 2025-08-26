package service

import (
	"context"

	"github.com/flick/backend/services/interaction/proto"
)

// InteractionService 定义互动服务接口
type InteractionService interface {
	// CreateFollow 创建关注
	CreateFollow(ctx context.Context, req *proto.CreateFollowRequest) (*proto.CreateFollowResponse, error)

	// DeleteFollow 删除关注
	DeleteFollow(ctx context.Context, req *proto.DeleteFollowRequest) (*proto.DeleteFollowResponse, error)

	// IsFollowing 检查是否关注
	IsFollowing(ctx context.Context, req *proto.IsFollowingRequest) (*proto.IsFollowingResponse, error)

	// GetFollowers 获取粉丝列表
	GetFollowers(ctx context.Context, req *proto.GetFollowersRequest) (*proto.GetFollowersResponse, error)

	// GetFollowing 获取关注列表
	GetFollowing(ctx context.Context, req *proto.GetFollowingRequest) (*proto.GetFollowingResponse, error)

	// CreateLike 创建点赞
	CreateLike(ctx context.Context, req *proto.CreateLikeRequest) (*proto.CreateLikeResponse, error)

	// DeleteLike 删除点赞
	DeleteLike(ctx context.Context, req *proto.DeleteLikeRequest) (*proto.DeleteLikeResponse, error)

	// IsLiked 检查是否点赞
	IsLiked(ctx context.Context, req *proto.IsLikedRequest) (*proto.IsLikedResponse, error)

	// GetLikes 获取点赞列表
	GetLikes(ctx context.Context, req *proto.GetLikesRequest) (*proto.GetLikesResponse, error)

	// CreateRepost 创建转发
	CreateRepost(ctx context.Context, req *proto.CreateRepostRequest) (*proto.CreateRepostResponse, error)

	// DeleteRepost 删除转发
	DeleteRepost(ctx context.Context, req *proto.DeleteRepostRequest) (*proto.DeleteRepostResponse, error)

	// CreateReply 创建回复
	CreateReply(ctx context.Context, req *proto.CreateReplyRequest) (*proto.CreateReplyResponse, error)

	// DeleteReply 删除回复
	DeleteReply(ctx context.Context, req *proto.DeleteReplyRequest) (*proto.DeleteReplyResponse, error)

	// GetReplies 获取回复列表
	GetReplies(ctx context.Context, req *proto.GetRepliesRequest) (*proto.GetRepliesResponse, error)

	// GetPostStats 获取帖子统计
	GetPostStats(ctx context.Context, req *proto.GetPostStatsRequest) (*proto.GetPostStatsResponse, error)

	// UpdatePostStats 更新帖子统计
	UpdatePostStats(ctx context.Context, req *proto.UpdatePostStatsRequest) (*proto.UpdatePostStatsResponse, error)

	// VotePoll 投票
	VotePoll(ctx context.Context, req *proto.VotePollRequest) (*proto.VotePollResponse, error)

	// CreateBookmark 创建收藏
	CreateBookmark(ctx context.Context, req *proto.CreateBookmarkRequest) (*proto.CreateBookmarkResponse, error)

	// DeleteBookmark 删除收藏
	DeleteBookmark(ctx context.Context, req *proto.DeleteBookmarkRequest) (*proto.DeleteBookmarkResponse, error)

	// IsBookmarked 检查是否收藏
	IsBookmarked(ctx context.Context, req *proto.IsBookmarkedRequest) (*proto.IsBookmarkedResponse, error)

	// GetBookmarks 获取收藏列表
	GetBookmarks(ctx context.Context, req *proto.GetBookmarksRequest) (*proto.GetBookmarksResponse, error)

	// CreateReport 创建举报
	CreateReport(ctx context.Context, req *proto.CreateReportRequest) (*proto.CreateReportResponse, error)
}
