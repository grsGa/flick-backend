package repository

import (
	"context"
	"github.com/flick/backend/services/recommendation/proto"
)

// RecommendationRepository 定义推荐仓储接口
type RecommendationRepository interface {
	// GetRecommendations 获取推荐内容
	GetRecommendations(ctx context.Context, userID string, limit int32) ([]*proto.RecommendationItem, error)
	
	// RecordUserAction 记录用户行为
	RecordUserAction(ctx context.Context, userID, postID, actionType string, weight float64) error
	
	// UpdateUserInterest 更新用户兴趣
	UpdateUserInterest(ctx context.Context, userID string, interests map[string]float64) error
	
	// GetUserInterests 获取用户兴趣
	GetUserInterests(ctx context.Context, userID string) (map[string]float64, error)
	
	// GetPostTags 获取帖子标签
	GetPostTags(ctx context.Context, postID string) ([]string, error)
}
