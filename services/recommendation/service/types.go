package service

import (
	"time"
)

// RecommendationItem 单个推荐项
type RecommendationItem struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ItemType  string    `json:"item_type"`
	ItemID    string    `json:"item_id"`
	Score     float64   `json:"score"`
	ModelID   string    `json:"model_id"`
	Reason    string    `json:"reason"`
	ExpiresAt time.Time `json:"expires_at"`
	IsClicked bool      `json:"is_clicked"`
	IsViewed  bool      `json:"is_viewed"`
	CreatedAt time.Time `json:"created_at"`
}

// UserFeed 用户个人推荐流
type UserFeed struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Items       []string  `json:"items"`
	LastUpdated time.Time `json:"last_updated"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// ContentFeature 内容特征
type ContentFeature struct {
	ID          string            `json:"id"`
	ContentType string            `json:"content_type"`
	ContentID   string            `json:"content_id"`
	Features    map[string]float64 `json:"features"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// RecommendationModel 推荐模型配置
type RecommendationModel struct {
	ID              string             `json:"id"`
	Name            string             `json:"name"`
	Description     string             `json:"description"`
	Type            string             `json:"type"`
	Parameters      map[string]interface{} `json:"parameters"`
	IsActive        bool               `json:"is_active"`
	Version         string             `json:"version"`
	TrainedAt       *time.Time         `json:"trained_at"`
	AccuracyMetrics map[string]float64 `json:"accuracy_metrics"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

// UserInterest 用户兴趣模型
type UserInterest struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Category  string    `json:"category"`
	Item      string    `json:"item"`
	Score     float64   `json:"score"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserActivityRecord 用户活动记录（用于推荐系统）
type UserActivityRecord struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	ActivityType string    `json:"activity_type"`
	TargetType   string    `json:"target_type"`
	TargetID     string    `json:"target_id"`
	Duration     int       `json:"duration"`
	Weight       float64   `json:"weight"`
	CreatedAt    time.Time `json:"created_at"`
}

// ABTest 推荐A/B测试
type ABTest struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	VariantA     string        `json:"variant_a"`
	VariantB     string        `json:"variant_b"`
	TrafficSplit float64       `json:"traffic_split"`
	IsActive     bool          `json:"is_active"`
	StartDate    time.Time     `json:"start_date"`
	EndDate      *time.Time    `json:"end_date"`
	Metrics      ABTestMetrics `json:"metrics"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

// ABTestMetrics A/B测试指标
type ABTestMetrics struct {
	VariantA     map[string]float64 `json:"variant_a"`
	VariantB     map[string]float64 `json:"variant_b"`
	Improvement  map[string]float64 `json:"improvement"`
	Significance map[string]float64 `json:"significance"`
}

// UserABTestGroup 用户A/B测试分组
type UserABTestGroup struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	TestID    string    `json:"test_id"`
	Variant   string    `json:"variant"`
	CreatedAt time.Time `json:"created_at"`
} 