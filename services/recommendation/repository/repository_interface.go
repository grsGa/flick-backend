package repository

import (
	"context"

	"backend/pkg/models"
)

// Repository 推荐数据仓库接口
type Repository interface {
	// 获取推荐
	GetRecommendations(ctx context.Context, userID string, limit int) ([]models.Recommendation, error)

	// 记录反馈
	RecordFeedback(ctx context.Context, feedback models.Feedback) error
	// 标记为已查看
	MarkAsViewed(ctx context.Context, recommendationID string, userID string) error
	// 标记为已点击
	MarkAsClicked(ctx context.Context, recommendationID string, userID string) error

	// 模型操作
	ListModels(ctx context.Context) ([]models.Model, error)
	GetModel(ctx context.Context, id string) (*models.Model, error)
	CreateModel(ctx context.Context, model models.Model) (string, error)
	UpdateModel(ctx context.Context, model models.Model) error
	DeleteModel(ctx context.Context, id string) error

	// AB测试操作
	ListABTests(ctx context.Context) ([]models.ABTest, error)
	GetABTest(ctx context.Context, id string) (*models.ABTest, error)
	CreateABTest(ctx context.Context, abTest models.ABTest) (string, error)
	UpdateABTest(ctx context.Context, abTest models.ABTest) error
	DeleteABTest(ctx context.Context, id string) error
	GetABTestMetrics(ctx context.Context, id string) (*models.ABTestMetrics, error)
} 