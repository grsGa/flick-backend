package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// UserInterest 用户兴趣模型
type UserInterest struct {
	ID        string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string         `json:"user_id" gorm:"type:uuid;not null;index"`
	Category  string         `json:"category" gorm:"size:50;not null"` // tag, topic, etc
	Item      string         `json:"item" gorm:"size:100;not null"`
	Score     float64        `json:"score" gorm:"default:0"`
	Source    string         `json:"source" gorm:"size:20;not null"` // explicit, implicit, system
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	User      User           `json:"-" gorm:"foreignKey:UserID"`
}

// UserActivityRecord 用户活动记录（用于推荐系统）
type UserActivityRecord struct {
	ID           string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       string    `json:"user_id" gorm:"type:uuid;not null;index"`
	ActivityType string    `json:"activity_type" gorm:"size:20;not null"` // view, like, comment, etc
	TargetType   string    `json:"target_type" gorm:"size:20;not null"`   // post, user, etc
	TargetID     string    `json:"target_id" gorm:"type:uuid;not null"`
	Duration     int       `json:"duration" gorm:"default:0"` // 停留时间（秒）
	Weight       float64   `json:"weight" gorm:"default:1"`   // 权重系数
	IPAddress    string    `json:"-"`
	DeviceInfo   string    `json:"-"`
	CreatedAt    time.Time `json:"created_at" gorm:"index"`
	User         User      `json:"-" gorm:"foreignKey:UserID"`
}

// RecommendationModel 推荐模型配置
type RecommendationModel struct {
	ID              string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name            string         `json:"name" gorm:"size:50;not null;uniqueIndex"`
	Description     string         `json:"description" gorm:"size:500"`
	Type            string         `json:"type" gorm:"size:20;not null"` // collaborative, content, hybrid
	Parameters      ModelParams    `json:"parameters" gorm:"type:jsonb"`
	IsActive        bool           `json:"is_active" gorm:"default:true"`
	Version         string         `json:"version" gorm:"size:20"`
	TrainedAt       *time.Time     `json:"trained_at"`
	AccuracyMetrics ModelMetrics   `json:"accuracy_metrics" gorm:"type:jsonb"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

// ModelParams 模型参数
type ModelParams map[string]interface{}

// Value 实现driver.Valuer接口
func (m ModelParams) Value() (driver.Value, error) {
	if len(m) == 0 {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan 实现sql.Scanner接口
func (m *ModelParams) Scan(value interface{}) error {
	if value == nil {
		*m = ModelParams{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("类型断言为[]byte失败")
	}

	return json.Unmarshal(bytes, &m)
}

// ModelMetrics 模型评估指标
type ModelMetrics map[string]float64

// Value 实现driver.Valuer接口
func (m ModelMetrics) Value() (driver.Value, error) {
	if len(m) == 0 {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan 实现sql.Scanner接口
func (m *ModelMetrics) Scan(value interface{}) error {
	if value == nil {
		*m = ModelMetrics{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("类型断言为[]byte失败")
	}

	return json.Unmarshal(bytes, &m)
}

// ContentFeature 内容特征
type ContentFeature struct {
	ID          string        `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ContentType string        `json:"content_type" gorm:"size:20;not null"` // post, user
	ContentID   string        `json:"content_id" gorm:"type:uuid;not null;uniqueIndex:idx_content"`
	Features    FeatureVector `json:"features" gorm:"type:jsonb"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// FeatureVector 特征向量
type FeatureVector map[string]float64

// Value 实现driver.Valuer接口
func (f FeatureVector) Value() (driver.Value, error) {
	if len(f) == 0 {
		return nil, nil
	}
	return json.Marshal(f)
}

// Scan 实现sql.Scanner接口
func (f *FeatureVector) Scan(value interface{}) error {
	if value == nil {
		*f = FeatureVector{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("类型断言为[]byte失败")
	}

	return json.Unmarshal(bytes, &f)
}

// RecommendationItem 单个推荐项
type RecommendationItem struct {
	ID        string              `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string              `json:"user_id" gorm:"type:uuid;not null;index"`
	ItemType  string              `json:"item_type" gorm:"size:20;not null"` // post, user
	ItemID    string              `json:"item_id" gorm:"type:uuid;not null"`
	Score     float64             `json:"score"`
	ModelID   string              `json:"model_id" gorm:"type:uuid"`
	Reason    string              `json:"reason" gorm:"size:200"` // 推荐原因
	ExpiresAt time.Time           `json:"expires_at" gorm:"index"`
	IsClicked bool                `json:"is_clicked" gorm:"default:false"`
	IsViewed  bool                `json:"is_viewed" gorm:"default:false"`
	CreatedAt time.Time           `json:"created_at"`
	User      User                `json:"-" gorm:"foreignKey:UserID"`
	Model     RecommendationModel `json:"-" gorm:"foreignKey:ModelID"`
}

// UserFeed 用户个人推荐流
type UserFeed struct {
	ID          string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      string    `json:"user_id" gorm:"type:uuid;not null;uniqueIndex"`
	Items       []string  `json:"items" gorm:"type:uuid[]"` // 推荐项ID数组
	LastUpdated time.Time `json:"last_updated"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
	User        User      `json:"-" gorm:"foreignKey:UserID"`
}

// Recommendation 表示一个推荐项
type Recommendation struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	ContentID string     `json:"content_id"`
	ModelID   string     `json:"model_id"`
	Score     float64    `json:"score"`
	Content   *Content   `json:"content,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Viewed    bool       `json:"viewed"`
	Clicked   bool       `json:"clicked"`
	ViewedAt  *time.Time `json:"viewed_at,omitempty"`
	ClickedAt *time.Time `json:"clicked_at,omitempty"`
}

// Content 表示推荐的内容
type Content struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	ContentType string    `json:"content_type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Feedback 表示用户对推荐的反馈
type Feedback struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	RecommendationID string    `json:"recommendation_id"`
	Rating           int       `json:"rating"`
	FeedbackType     string    `json:"feedback_type"`
	Comment          string    `json:"comment,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// Model 表示推荐模型
type Model struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     string    `json:"version"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ABTest 表示A/B测试
type ABTest struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	Status            string    `json:"status"`
	ControlModelID    string    `json:"control_model_id"`
	ExperimentModelID string    `json:"experiment_model_id"`
	TrafficPercentage float64   `json:"traffic_percentage"`
	StartDate         time.Time `json:"start_date"`
	EndDate           time.Time `json:"end_date"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// ABTestMetrics 表示A/B测试指标
type ABTestMetrics struct {
	ABTestID              string  `json:"ab_test_id"`
	ControlImpressions    int64   `json:"control_impressions"`
	ControlViews          int64   `json:"control_views"`
	ControlClicks         int64   `json:"control_clicks"`
	ExperimentImpressions int64   `json:"experiment_impressions"`
	ExperimentViews       int64   `json:"experiment_views"`
	ExperimentClicks      int64   `json:"experiment_clicks"`
	ControlViewRate       float64 `json:"control_view_rate"`
	ControlClickRate      float64 `json:"control_click_rate"`
	ExperimentViewRate    float64 `json:"experiment_view_rate"`
	ExperimentClickRate   float64 `json:"experiment_click_rate"`
	ViewLift              float64 `json:"view_lift"`
	ClickLift             float64 `json:"click_lift"`
}

// UserABTestGroup 用户A/B测试分组
type UserABTestGroup struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string    `json:"user_id" gorm:"type:uuid;not null;index"`
	TestID    string    `json:"test_id" gorm:"type:uuid;not null"`
	Variant   string    `json:"variant" gorm:"size:1;not null"` // "A" 或 "B"
	CreatedAt time.Time `json:"created_at"`
	User      User      `json:"-" gorm:"foreignKey:UserID"`
	ABTest    ABTest    `json:"-" gorm:"foreignKey:TestID"`
}
