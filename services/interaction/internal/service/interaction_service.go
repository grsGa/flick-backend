package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/flick/backend/services/interaction/internal/repository"
	"github.com/flick/backend/services/interaction/proto"
)

// interactionService 互动服务实现
type InteractionService struct {
	interactionRepo *repository.InteractionRepository
}

// NewInteractionService 创建互动服务实例
func NewInteractionService(interactionRepo *repository.InteractionRepository) *InteractionService {
	return &InteractionService{
		interactionRepo: interactionRepo,
	}
}

// CreateFollow 创建关注
func (s *InteractionService) CreateFollow(ctx context.Context, req *proto.CreateFollowRequest) (*proto.CreateFollowResponse, error) {
	// 检查是否已经关注
	isFollowing, err := s.interactionRepo.IsFollowing(ctx, req.FollowerId, req.FolloweeId)
	if err != nil {
		return &proto.CreateFollowResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to check following status: " + err.Error(),
			},
		}, err
	}

	// 如果已经关注，直接返回成功（幂等操作）
	if isFollowing {
		// 获取现有的关注记录
		existingFollow, err := s.interactionRepo.GetFollow(ctx, req.FollowerId, req.FolloweeId)
		if err != nil {
			return &proto.CreateFollowResponse{
				Error: &proto.Error{
					Code:    500,
					Message: "Failed to get existing follow: " + err.Error(),
				},
			}, err
		}
		
		return &proto.CreateFollowResponse{
			Follow: existingFollow,
		}, nil
	}

	// 创建关注对象
	follow := &proto.Follow{
		Id:         uuid.New().String(),
		FollowerId: req.FollowerId,
		FolloweeId: req.FolloweeId,
		CreatedAt:  time.Now().Format(time.RFC3339),
	}

	// 保存到数据库
	err = s.interactionRepo.CreateFollow(ctx, follow)
	if err != nil {
		return &proto.CreateFollowResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to create follow: " + err.Error(),
			},
		}, err
	}

	return &proto.CreateFollowResponse{
		Follow: follow,
	}, nil
}

// DeleteFollow 删除关注
func (s *InteractionService) DeleteFollow(ctx context.Context, req *proto.DeleteFollowRequest) (*proto.DeleteFollowResponse, error) {
	err := s.interactionRepo.DeleteFollow(ctx, req.FollowerId, req.FolloweeId)
	if err != nil {
		return &proto.DeleteFollowResponse{
			Success: false,
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to delete follow: " + err.Error(),
			},
		}, err
	}

	return &proto.DeleteFollowResponse{
		Success: true,
	}, nil
}

// IsFollowing 检查是否关注
func (s *InteractionService) IsFollowing(ctx context.Context, req *proto.IsFollowingRequest) (*proto.IsFollowingResponse, error) {
	isFollowing, err := s.interactionRepo.IsFollowing(ctx, req.FollowerId, req.FolloweeId)
	if err != nil {
		return &proto.IsFollowingResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to check following status: " + err.Error(),
			},
		}, err
	}

	return &proto.IsFollowingResponse{
		IsFollowing: isFollowing,
	}, nil
}

// GetFollowers 获取粉丝列表
func (s *InteractionService) GetFollowers(ctx context.Context, req *proto.GetFollowersRequest) (*proto.GetFollowersResponse, error) {
	followers, total, err := s.interactionRepo.GetFollowers(ctx, req.UserId, req.Page, req.PageSize)
	if err != nil {
		return &proto.GetFollowersResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to get followers: " + err.Error(),
			},
		}, err
	}

	return &proto.GetFollowersResponse{
		Followers: followers,
		Total:     total,
	}, nil
}

// GetFollowing 获取关注列表
func (s *InteractionService) GetFollowing(ctx context.Context, req *proto.GetFollowingRequest) (*proto.GetFollowingResponse, error) {
	following, total, err := s.interactionRepo.GetFollowing(ctx, req.UserId, req.Page, req.PageSize)
	if err != nil {
		return &proto.GetFollowingResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to get following: " + err.Error(),
			},
		}, err
	}

	return &proto.GetFollowingResponse{
		Following: following,
		Total:     total,
	}, nil
}

// CreateLike 创建点赞
func (s *InteractionService) CreateLike(ctx context.Context, req *proto.CreateLikeRequest) (*proto.CreateLikeResponse, error) {
	// 检查是否已经点赞
	isLiked, err := s.interactionRepo.IsLiked(ctx, req.UserId, req.PostId)
	if err != nil {
		return &proto.CreateLikeResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to check like status: " + err.Error(),
			},
		}, err
	}

	// 如果已经点赞，直接返回
	if isLiked {
		return &proto.CreateLikeResponse{
			Error: &proto.Error{
				Code:    400,
				Message: "Already liked",
			},
		}, nil
	}

	// 创建点赞对象
	like := &proto.Like{
		Id:        uuid.New().String(),
		UserId:    req.UserId,
		PostId:    req.PostId,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	// 保存到数据库
	err = s.interactionRepo.CreateLike(ctx, like)
	if err != nil {
		return &proto.CreateLikeResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to create like: " + err.Error(),
			},
		}, err
	}

	return &proto.CreateLikeResponse{
		Like: like,
	}, nil
}

// DeleteLike 删除点赞
func (s *InteractionService) DeleteLike(ctx context.Context, req *proto.DeleteLikeRequest) (*proto.DeleteLikeResponse, error) {
	err := s.interactionRepo.DeleteLike(ctx, req.UserId, req.PostId)
	if err != nil {
		return &proto.DeleteLikeResponse{
			Success: false,
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to delete like: " + err.Error(),
			},
		}, err
	}

	return &proto.DeleteLikeResponse{
		Success: true,
	}, nil
}

// IsLiked 检查是否点赞
func (s *InteractionService) IsLiked(ctx context.Context, req *proto.IsLikedRequest) (*proto.IsLikedResponse, error) {
	isLiked, err := s.interactionRepo.IsLiked(ctx, req.UserId, req.PostId)
	if err != nil {
		return &proto.IsLikedResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to check like status: " + err.Error(),
			},
		}, err
	}

	return &proto.IsLikedResponse{
		IsLiked: isLiked,
	}, nil
}

// GetLikes 获取点赞列表
func (s *InteractionService) GetLikes(ctx context.Context, req *proto.GetLikesRequest) (*proto.GetLikesResponse, error) {
	likes, total, err := s.interactionRepo.GetLikes(ctx, req.PostId, req.Page, req.PageSize)
	if err != nil {
		return &proto.GetLikesResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to get likes: " + err.Error(),
			},
		}, err
	}

	return &proto.GetLikesResponse{
		Likes: likes,
		Total: total,
	}, nil
}

// CreateRepost 创建转发
func (s *InteractionService) CreateRepost(ctx context.Context, req *proto.CreateRepostRequest) (*proto.CreateRepostResponse, error) {
	// 创建转发对象
	repost := &proto.Repost{
		Id:        uuid.New().String(),
		UserId:    req.UserId,
		PostId:    req.PostId,
		Reply:     req.Reply,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	// 保存到数据库
	err := s.interactionRepo.CreateRepost(ctx, repost)
	if err != nil {
		return &proto.CreateRepostResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to create repost: " + err.Error(),
			},
		}, err
	}

	return &proto.CreateRepostResponse{
		Repost: repost,
	}, nil
}

// DeleteRepost 删除转发
func (s *InteractionService) DeleteRepost(ctx context.Context, req *proto.DeleteRepostRequest) (*proto.DeleteRepostResponse, error) {
	err := s.interactionRepo.DeleteRepost(ctx, req.UserId, req.PostId)
	if err != nil {
		return &proto.DeleteRepostResponse{
			Success: false,
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to delete repost: " + err.Error(),
			},
		}, err
	}

	return &proto.DeleteRepostResponse{
		Success: true,
	}, nil
}

// CreateReport 创建举报
func (s *InteractionService) CreateReport(ctx context.Context, req *proto.CreateReportRequest) (*proto.CreateReportResponse, error) {
	// 创建举报对象
	report := &proto.Report{
		Id:         uuid.New().String(),
		UserId:     req.UserId,
		TargetId:   req.TargetId,
		TargetType: req.TargetType,
		Reason:     req.Reason,
		CreatedAt:  time.Now().Format(time.RFC3339),
	}

	// 保存到数据库
	err := s.interactionRepo.CreateReport(ctx, report)
	if err != nil {
		return &proto.CreateReportResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to create report: " + err.Error(),
			},
		}, err
	}

	return &proto.CreateReportResponse{
		Report: report,
	}, nil
}

// IsBookmarked 检查是否收藏
func (s *InteractionService) IsBookmarked(ctx context.Context, req *proto.IsBookmarkedRequest) (*proto.IsBookmarkedResponse, error) {
	isBookmarked, err := s.interactionRepo.IsBookmarked(ctx, req.UserId, req.PostId)
	if err != nil {
		return &proto.IsBookmarkedResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to check bookmark status: " + err.Error(),
			},
		}, err
	}

	return &proto.IsBookmarkedResponse{
		IsBookmarked: isBookmarked,
	}, nil
}

// GetBookmarks 获取收藏列表
func (s *InteractionService) GetBookmarks(ctx context.Context, req *proto.GetBookmarksRequest) (*proto.GetBookmarksResponse, error) {
	bookmarks, nextCursor, hasMore, err := s.interactionRepo.GetBookmarks(ctx, req.UserId, req.Limit, req.Cursor)
	if err != nil {
		return &proto.GetBookmarksResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to get bookmarks: " + err.Error(),
			},
		}, err
	}

	return &proto.GetBookmarksResponse{
		Bookmarks:  bookmarks,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

// CreateReply 创建回复
func (s *InteractionService) CreateReply(ctx context.Context, req *proto.CreateReplyRequest) (*proto.CreateReplyResponse, error) {
	reply := &proto.Reply{
		Id:            uuid.New().String(),
		PostId:        req.PostId,
		UserId:        req.UserId,
		Content:       req.Content,
		ParentReplyId: req.ParentReplyId,
		CreatedAt:     time.Now().Format(time.RFC3339),
		UpdatedAt:     time.Now().Format(time.RFC3339),
	}

	err := s.interactionRepo.CreateReply(ctx, reply)
	if err != nil {
		return &proto.CreateReplyResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to create reply: " + err.Error(),
			},
		}, err
	}

	return &proto.CreateReplyResponse{
		Reply: reply,
	}, nil
}

// DeleteReply 删除回复
func (s *InteractionService) DeleteReply(ctx context.Context, req *proto.DeleteReplyRequest) (*proto.DeleteReplyResponse, error) {
	err := s.interactionRepo.DeleteReply(ctx, req.ReplyId, req.UserId)
	if err != nil {
		return &proto.DeleteReplyResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to delete reply: " + err.Error(),
			},
		}, err
	}

	return &proto.DeleteReplyResponse{
		Success: true,
	}, nil
}

// GetReplies 获取回复列表
func (s *InteractionService) GetReplies(ctx context.Context, req *proto.GetRepliesRequest) (*proto.GetRepliesResponse, error) {
	replies, nextCursor, hasMore, err := s.interactionRepo.GetReplies(ctx, req.PostId, req.Limit, req.Cursor)
	if err != nil {
		return &proto.GetRepliesResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to get replies: " + err.Error(),
			},
		}, err
	}

	return &proto.GetRepliesResponse{
		Replies:    replies,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

// GetPostStats 获取帖子统计
func (s *InteractionService) GetPostStats(ctx context.Context, req *proto.GetPostStatsRequest) (*proto.GetPostStatsResponse, error) {
	stats, err := s.interactionRepo.GetPostStats(ctx, req.PostId)
	if err != nil {
		return &proto.GetPostStatsResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to get post stats: " + err.Error(),
			},
		}, err
	}

	return &proto.GetPostStatsResponse{
		Stats: stats,
	}, nil
}

// UpdatePostStats 更新帖子统计
func (s *InteractionService) UpdatePostStats(ctx context.Context, req *proto.UpdatePostStatsRequest) (*proto.UpdatePostStatsResponse, error) {
	stats, err := s.interactionRepo.UpdatePostStats(ctx, req.PostId, req.Action, req.Delta)
	if err != nil {
		return &proto.UpdatePostStatsResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to update post stats: " + err.Error(),
			},
		}, err
	}

	return &proto.UpdatePostStatsResponse{
		Stats: stats,
	}, nil
}

// VotePoll 投票
func (s *InteractionService) VotePoll(ctx context.Context, req *proto.VotePollRequest) (*proto.VotePollResponse, error) {
	vote := &proto.PollVote{
		Id:           uuid.New().String(),
		PollId:       req.PollId,
		PollOptionId: req.PollOptionId,
		UserId:       req.UserId,
		CreatedAt:    time.Now().Format(time.RFC3339),
	}

	err := s.interactionRepo.VotePoll(ctx, vote)
	if err != nil {
		return &proto.VotePollResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to vote poll: " + err.Error(),
			},
		}, err
	}

	return &proto.VotePollResponse{
		Vote: vote,
	}, nil
}

// CreateBookmark 创建收藏
func (s *InteractionService) CreateBookmark(ctx context.Context, req *proto.CreateBookmarkRequest) (*proto.CreateBookmarkResponse, error) {
	// 检查是否已经收藏
	isBookmarked, err := s.interactionRepo.IsBookmarked(ctx, req.UserId, req.PostId)
	if err != nil {
		return &proto.CreateBookmarkResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to check bookmark status: " + err.Error(),
			},
		}, err
	}

	if isBookmarked {
		return &proto.CreateBookmarkResponse{
			Error: &proto.Error{
				Code:    400,
				Message: "Already bookmarked",
			},
		}, nil
	}

	bookmark := &proto.Bookmark{
		Id:        uuid.New().String(),
		UserId:    req.UserId,
		PostId:    req.PostId,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	err = s.interactionRepo.CreateBookmark(ctx, bookmark)
	if err != nil {
		return &proto.CreateBookmarkResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to create bookmark: " + err.Error(),
			},
		}, err
	}

	return &proto.CreateBookmarkResponse{
		Bookmark: bookmark,
	}, nil
}

// DeleteBookmark 删除收藏
func (s *InteractionService) DeleteBookmark(ctx context.Context, req *proto.DeleteBookmarkRequest) (*proto.DeleteBookmarkResponse, error) {
	err := s.interactionRepo.DeleteBookmark(ctx, req.UserId, req.PostId)
	if err != nil {
		return &proto.DeleteBookmarkResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to delete bookmark: " + err.Error(),
			},
		}, err
	}

	return &proto.DeleteBookmarkResponse{
		Success: true,
	}, nil
}

// LikePost 点赞帖子 (高级接口，包含统计更新)
func (s *InteractionService) LikePost(ctx context.Context, req *proto.LikePostRequest) (*proto.LikePostResponse, error) {
	// 检查是否已经点赞
	isLiked, err := s.interactionRepo.IsLiked(ctx, req.UserId, req.PostId)
	if err != nil {
		return &proto.LikePostResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to check like status: " + err.Error(),
			},
		}, err
	}

	// 如果已经点赞，返回当前状态
	if isLiked {
		// 获取当前点赞数
		stats, err := s.interactionRepo.GetPostStats(ctx, req.PostId)
		if err != nil {
			return &proto.LikePostResponse{
				IsLiked:   true,
				LikeCount: 0, // 默认值
			}, nil
		}
		return &proto.LikePostResponse{
			IsLiked:   true,
			LikeCount: stats.LikeCount,
		}, nil
	}

	// 创建点赞记录
	like := &proto.Like{
		Id:        uuid.New().String(),
		UserId:    req.UserId,
		PostId:    req.PostId,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	err = s.interactionRepo.CreateLike(ctx, like)
	if err != nil {
		return &proto.LikePostResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to create like: " + err.Error(),
			},
		}, err
	}

	// 更新帖子统计
	stats, err := s.interactionRepo.UpdatePostStats(ctx, req.PostId, "like", 1)
	if err != nil {
		// 即使统计更新失败，点赞操作已成功，返回成功状态
		return &proto.LikePostResponse{
			IsLiked:   true,
			LikeCount: 1, // 至少有当前用户的点赞
		}, nil
	}

	return &proto.LikePostResponse{
		IsLiked:   true,
		LikeCount: stats.LikeCount,
	}, nil
}

// UnlikePost 取消点赞帖子 (高级接口，包含统计更新)
func (s *InteractionService) UnlikePost(ctx context.Context, req *proto.UnlikePostRequest) (*proto.UnlikePostResponse, error) {
	// 检查是否已经点赞
	isLiked, err := s.interactionRepo.IsLiked(ctx, req.UserId, req.PostId)
	if err != nil {
		return &proto.UnlikePostResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to check like status: " + err.Error(),
			},
		}, err
	}

	// 如果没有点赞，返回当前状态
	if !isLiked {
		// 获取当前点赞数
		stats, err := s.interactionRepo.GetPostStats(ctx, req.PostId)
		if err != nil {
			return &proto.UnlikePostResponse{
				IsLiked:   false,
				LikeCount: 0, // 默认值
			}, nil
		}
		return &proto.UnlikePostResponse{
			IsLiked:   false,
			LikeCount: stats.LikeCount,
		}, nil
	}

	// 删除点赞记录
	err = s.interactionRepo.DeleteLike(ctx, req.UserId, req.PostId)
	if err != nil {
		return &proto.UnlikePostResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to delete like: " + err.Error(),
			},
		}, err
	}

	// 更新帖子统计
	stats, err := s.interactionRepo.UpdatePostStats(ctx, req.PostId, "unlike", -1)
	if err != nil {
		// 即使统计更新失败，取消点赞操作已成功，返回成功状态
		return &proto.UnlikePostResponse{
			IsLiked:   false,
			LikeCount: 0, // 保守估计
		}, nil
	}

	return &proto.UnlikePostResponse{
		IsLiked:   false,
		LikeCount: stats.LikeCount,
	}, nil
}

