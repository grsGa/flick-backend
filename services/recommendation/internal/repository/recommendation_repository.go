package repository

import (
	"context"
	"time"

	"github.com/flick/backend/pkg/database"
	"github.com/flick/backend/pkg/models"
	"github.com/flick/backend/services/recommendation/proto"
	"gorm.io/gorm"
)

// recommendationRepository 推荐仓储实现
type recommendationRepository struct {
	db *gorm.DB
}

// NewRecommendationRepository 创建推荐仓储实例
func NewRecommendationRepository() RecommendationRepository {
	return &recommendationRepository{
		db: database.GetDB(),
	}
}

// GetRecommendations 获取推荐内容
func (r *recommendationRepository) GetRecommendations(ctx context.Context, userID string, limit int32) ([]*proto.RecommendationItem, error) {
	var feedItems []models.FeedItem

	// 简化实现：直接从feed_items表获取推荐内容
	// 实际推荐算法会更复杂，可能涉及协同过滤、内容推荐等
	if err := r.db.Where("user_id = ?", userID).
		Order("rank_score DESC").
		Limit(int(limit)).
		Find(&feedItems).Error; err != nil {
		return nil, err
	}

	items := make([]*proto.RecommendationItem, len(feedItems))
	for i, item := range feedItems {
		items[i] = &proto.RecommendationItem{
			Id:        item.ID,
			UserId:    item.UserID,
			PostId:    item.PostID,
			Score:     item.RankScore,
			CreatedAt: item.CreatedAt.Format(time.RFC3339),
		}
	}

	return items, nil
}

// RecordUserAction 记录用户行为
func (r *recommendationRepository) RecordUserAction(ctx context.Context, userID, postID, actionType string, weight float64) error {
	// 在实际实现中，这里会记录用户行为并更新推荐模型
	// 简化处理，直接返回nil
	return nil
}

// UpdateUserInterest 更新用户兴趣
func (r *recommendationRepository) UpdateUserInterest(ctx context.Context, userID string, interests map[string]float64) error {
	// 在实际实现中，这里会更新用户兴趣模型
	// 简化处理，直接返回nil
	return nil
}

// GetUserInterests 获取用户兴趣
func (r *recommendationRepository) GetUserInterests(ctx context.Context, userID string) (map[string]float64, error) {
	// 在实际实现中，这里会获取用户兴趣数据
	// 简化处理，返回空map
	return make(map[string]float64), nil
}

// GetPostTags 获取帖子标签
func (r *recommendationRepository) GetPostTags(ctx context.Context, postID string) ([]string, error) {
	// 在实际实现中，这里会获取帖子的标签信息
	// 简化处理，返回空切片
	return []string{}, nil
}
