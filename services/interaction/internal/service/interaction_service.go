package service

import (
	"context"
	"time"

	"github.com/flick/backend/services/interaction/internal/repository"
	"github.com/flick/backend/services/interaction/proto"
	"github.com/google/uuid"
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

	// 如果已经关注，直接返回
	if isFollowing {
		return &proto.CreateFollowResponse{
			Error: &proto.Error{
				Code:    400,
				Message: "Already following",
			},
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
