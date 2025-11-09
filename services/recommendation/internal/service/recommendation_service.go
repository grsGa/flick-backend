package service

import (
	"context"

	"github.com/flick/backend/services/recommendation/internal/repository"
	"github.com/flick/backend/services/recommendation/proto"
)

// recommendationService 推荐服务实现
type RecommendationService struct {
	recommendationRepo *repository.RecommendationRepository
}

// NewRecommendationService 创建推荐服务实例
func NewRecommendationService(recommendationRepo *repository.RecommendationRepository) *RecommendationService {
	return &RecommendationService{
		recommendationRepo: recommendationRepo,
	}
}

// GetRecommendations 获取推荐内容
func (s *RecommendationService) GetRecommendations(ctx context.Context, req *proto.GetRecommendationsRequest) (*proto.GetRecommendationsResponse, error) {
	items, err := s.recommendationRepo.GetRecommendations(ctx, req.UserId, req.Limit)
	if err != nil {
		return &proto.GetRecommendationsResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to get recommendations: " + err.Error(),
			},
		}, err
	}

	return &proto.GetRecommendationsResponse{
		Items: items,
	}, nil
}

// RecordUserAction 记录用户行为
func (s *RecommendationService) RecordUserAction(ctx context.Context, req *proto.RecordUserActionRequest) (*proto.RecordUserActionResponse, error) {
	err := s.recommendationRepo.RecordUserAction(ctx, req.UserId, req.PostId, req.ActionType, req.Weight)
	if err != nil {
		return &proto.RecordUserActionResponse{
			Success: false,
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to record user action: " + err.Error(),
			},
		}, err
	}

	return &proto.RecordUserActionResponse{
		Success: true,
	}, nil
}

// UpdateUserInterest 更新用户兴趣
func (s *RecommendationService) UpdateUserInterest(ctx context.Context, req *proto.UpdateUserInterestRequest) (*proto.UpdateUserInterestResponse, error) {
	err := s.recommendationRepo.UpdateUserInterest(ctx, req.UserId, req.Interests)
	if err != nil {
		return &proto.UpdateUserInterestResponse{
			Success: false,
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to update user interest: " + err.Error(),
			},
		}, err
	}

	return &proto.UpdateUserInterestResponse{
		Success: true,
	}, nil
}

