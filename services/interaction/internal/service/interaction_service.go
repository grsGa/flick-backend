package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/flick/backend/services/interaction/internal/repository"
	"github.com/flick/backend/services/interaction/proto"
)

// interactionService 互动服务实现
type interactionService struct {
	interactionRepo repository.InteractionRepository
}

// NewInteractionService 创建互动服务实例
func NewInteractionService(interactionRepo repository.InteractionRepository) InteractionService {
	return &interactionService{
		interactionRepo: interactionRepo,
	}
}

// CreateFollow 创建关注
func (s *interactionService) CreateFollow(ctx context.Context, req *proto.CreateFollowRequest) (*proto.CreateFollowResponse, error) {
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
func (s *interactionService) DeleteFollow(ctx context.Context, req *proto.DeleteFollowRequest) (*proto.DeleteFollowResponse, error) {
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
func (s *interactionService) IsFollowing(ctx context.Context, req *proto.IsFollowingRequest) (*proto.IsFollowingResponse, error) {
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
func (s *interactionService) GetFollowers(ctx context.Context, req *proto.GetFollowersRequest) (*proto.GetFollowersResponse, error) {
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
func (s *interactionService) GetFollowing(ctx context.Context, req *proto.GetFollowingRequest) (*proto.GetFollowingResponse, error) {
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
func (s *interactionService) CreateLike(ctx context.Context, req *proto.CreateLikeRequest) (*proto.CreateLikeResponse, error) {
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
func (s *interactionService) DeleteLike(ctx context.Context, req *proto.DeleteLikeRequest) (*proto.DeleteLikeResponse, error) {
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
func (s *interactionService) IsLiked(ctx context.Context, req *proto.IsLikedRequest) (*proto.IsLikedResponse, error) {
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
func (s *interactionService) GetLikes(ctx context.Context, req *proto.GetLikesRequest) (*proto.GetLikesResponse, error) {
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
func (s *interactionService) CreateRepost(ctx context.Context, req *proto.CreateRepostRequest) (*proto.CreateRepostResponse, error) {
	// 创建转发对象
	repost := &proto.Repost{
		Id:        uuid.New().String(),
		UserId:    req.UserId,
		PostId:    req.PostId,
		Comment:   req.Comment,
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
func (s *interactionService) DeleteRepost(ctx context.Context, req *proto.DeleteRepostRequest) (*proto.DeleteRepostResponse, error) {
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
func (s *interactionService) CreateReport(ctx context.Context, req *proto.CreateReportRequest) (*proto.CreateReportResponse, error) {
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

// CreateComment 创建评论
func (s *interactionService) CreateComment(ctx context.Context, req *proto.CreateCommentRequest) (*proto.CreateCommentResponse, error) {
	comment := &proto.Comment{
		Id:              uuid.New().String(),
		PostId:          req.PostId,
		UserId:          req.UserId,
		Content:         req.Content,
		ParentCommentId: req.ParentCommentId,
		CreatedAt:       time.Now().Format(time.RFC3339),
		UpdatedAt:       time.Now().Format(time.RFC3339),
	}

	err := s.interactionRepo.CreateComment(ctx, comment)
	if err != nil {
		return &proto.CreateCommentResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to create comment: " + err.Error(),
			},
		}, err
	}

	return &proto.CreateCommentResponse{
		Comment: comment,
	}, nil
}

// DeleteComment 删除评论
func (s *interactionService) DeleteComment(ctx context.Context, req *proto.DeleteCommentRequest) (*proto.DeleteCommentResponse, error) {
	err := s.interactionRepo.DeleteComment(ctx, req.CommentId, req.UserId)
	if err != nil {
		return &proto.DeleteCommentResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to delete comment: " + err.Error(),
			},
		}, err
	}

	return &proto.DeleteCommentResponse{
		Success: true,
	}, nil
}

// GetComments 获取评论列表
func (s *interactionService) GetComments(ctx context.Context, req *proto.GetCommentsRequest) (*proto.GetCommentsResponse, error) {
	comments, nextCursor, hasMore, err := s.interactionRepo.GetComments(ctx, req.PostId, req.Limit, req.Cursor)
	if err != nil {
		return &proto.GetCommentsResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to get comments: " + err.Error(),
			},
		}, err
	}

	return &proto.GetCommentsResponse{
		Comments:   comments,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

// GetPostStats 获取帖子统计
func (s *interactionService) GetPostStats(ctx context.Context, req *proto.GetPostStatsRequest) (*proto.GetPostStatsResponse, error) {
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
func (s *interactionService) UpdatePostStats(ctx context.Context, req *proto.UpdatePostStatsRequest) (*proto.UpdatePostStatsResponse, error) {
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
func (s *interactionService) VotePoll(ctx context.Context, req *proto.VotePollRequest) (*proto.VotePollResponse, error) {
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
func (s *interactionService) CreateBookmark(ctx context.Context, req *proto.CreateBookmarkRequest) (*proto.CreateBookmarkResponse, error) {
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
func (s *interactionService) DeleteBookmark(ctx context.Context, req *proto.DeleteBookmarkRequest) (*proto.DeleteBookmarkResponse, error) {
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

// IsBookmarked 检查是否收藏
func (s *interactionService) IsBookmarked(ctx context.Context, req *proto.IsBookmarkedRequest) (*proto.IsBookmarkedResponse, error) {
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
func (s *interactionService) GetBookmarks(ctx context.Context, req *proto.GetBookmarksRequest) (*proto.GetBookmarksResponse, error) {
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
