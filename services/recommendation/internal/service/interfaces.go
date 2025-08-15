package service

import (
	"context"
	"github.com/flick/backend/services/recommendation/proto"
)

// RecommendationService 定义推荐服务接口
type RecommendationService interface {
	// GetRecommendations 获取推荐内容
	GetRecommendations(ctx context.Context, req *proto.GetRecommendationsRequest) (*proto.GetRecommendationsResponse, error)
	
	// RecordUserAction 记录用户行为
	RecordUserAction(ctx context.Context, req *proto.RecordUserActionRequest) (*proto.RecordUserActionResponse, error)
	
	// UpdateUserInterest 更新用户兴趣
	UpdateUserInterest(ctx context.Context, req *proto.UpdateUserInterestRequest) (*proto.UpdateUserInterestResponse, error)
}
