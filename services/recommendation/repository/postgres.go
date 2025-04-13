package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"backend/pkg/models"
)

var (
	ErrNotFound        = errors.New("记录未找到")
	ErrModelNotFound   = errors.New("推荐模型不存在")
	ErrFeatureNotFound = errors.New("内容特征不存在")
	ErrABTestNotFound  = errors.New("A/B测试不存在")
)

// PostgresRepository PostgresSQL实现推荐仓库接口
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository 创建PostgresSQL仓库实例
func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

// GetRecommendations 获取推荐
func (r *PostgresRepository) GetRecommendations(ctx context.Context, userID string, limit int) ([]models.Recommendation, error) {
	// 简化实现
	return []models.Recommendation{}, nil
}

// RecordFeedback 记录反馈
func (r *PostgresRepository) RecordFeedback(ctx context.Context, feedback models.Feedback) error {
	// 简化实现
	return nil
}

// MarkAsViewed 标记为已查看
func (r *PostgresRepository) MarkAsViewed(ctx context.Context, recommendationID string, userID string) error {
	// 简化实现
	return nil
}

// MarkAsClicked 标记为已点击
func (r *PostgresRepository) MarkAsClicked(ctx context.Context, recommendationID string, userID string) error {
	// 简化实现
	return nil
}

// ListModels 列出所有模型
func (r *PostgresRepository) ListModels(ctx context.Context) ([]models.Model, error) {
	// 简化实现
	return []models.Model{}, nil
}

// GetModel 获取单个模型
func (r *PostgresRepository) GetModel(ctx context.Context, id string) (*models.Model, error) {
	// 简化实现
	return nil, nil
}

// CreateModel 创建新模型
func (r *PostgresRepository) CreateModel(ctx context.Context, model models.Model) (string, error) {
	// 简化实现
	return "", nil
}

// UpdateModel 更新模型
func (r *PostgresRepository) UpdateModel(ctx context.Context, model models.Model) error {
	// 简化实现
	return nil
}

// DeleteModel 删除模型
func (r *PostgresRepository) DeleteModel(ctx context.Context, id string) error {
	// 简化实现
	return nil
}

// ListABTests 列出所有AB测试
func (r *PostgresRepository) ListABTests(ctx context.Context) ([]models.ABTest, error) {
	// 简化实现
	return []models.ABTest{}, nil
}

// GetABTest 获取单个AB测试
func (r *PostgresRepository) GetABTest(ctx context.Context, id string) (*models.ABTest, error) {
	// 简化实现
	return nil, nil
}

// CreateABTest 创建新AB测试
func (r *PostgresRepository) CreateABTest(ctx context.Context, abTest models.ABTest) (string, error) {
	// 简化实现
	return "", nil
}

// UpdateABTest 更新AB测试
func (r *PostgresRepository) UpdateABTest(ctx context.Context, abTest models.ABTest) error {
	// 简化实现
	return nil
}

// DeleteABTest 删除AB测试
func (r *PostgresRepository) DeleteABTest(ctx context.Context, id string) error {
	// 简化实现
	return nil
}

// GetABTestMetrics 获取AB测试指标
func (r *PostgresRepository) GetABTestMetrics(ctx context.Context, id string) (*models.ABTestMetrics, error) {
	// 简化实现
	return nil, nil
}

// CreateUserInterest 创建用户兴趣
func (r *PostgresRepository) CreateUserInterest(ctx context.Context, interest *models.UserInterest) error {
	return r.db.WithContext(ctx).Create(interest).Error
}

// GetUserInterestByID 通过ID获取用户兴趣
func (r *PostgresRepository) GetUserInterestByID(ctx context.Context, interestID string) (*models.UserInterest, error) {
	var interest models.UserInterest
	err := r.db.WithContext(ctx).Where("id = ?", interestID).First(&interest).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &interest, nil
}

// GetUserInterests 获取用户所有兴趣
func (r *PostgresRepository) GetUserInterests(ctx context.Context, userID string) ([]*models.UserInterest, error) {
	var interests []*models.UserInterest
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&interests).Error
	if err != nil {
		return nil, err
	}
	return interests, nil
}

// UpdateUserInterest 更新用户兴趣
func (r *PostgresRepository) UpdateUserInterest(ctx context.Context, interest *models.UserInterest) error {
	return r.db.WithContext(ctx).Save(interest).Error
}

// DeleteUserInterest 删除用户兴趣
func (r *PostgresRepository) DeleteUserInterest(ctx context.Context, interestID string) error {
	return r.db.WithContext(ctx).Delete(&models.UserInterest{}, "id = ?", interestID).Error
}

// CreateUserActivity 创建用户活动记录
func (r *PostgresRepository) CreateUserActivity(ctx context.Context, activity interface{}) error {
	// 将activity转换为models.UserActivityRecord
	var modelActivity *models.UserActivityRecord

	switch a := activity.(type) {
	case *models.UserActivityRecord:
		modelActivity = a
	case *SimpleUserActivityRecord:
		modelActivity = &models.UserActivityRecord{
			ID:           a.ID,
			UserID:       a.UserID,
			ActivityType: a.ActivityType,
			TargetType:   a.TargetType,
			TargetID:     a.TargetID,
			Duration:     a.Duration,
			Weight:       a.Weight,
			CreatedAt:    a.CreatedAt,
		}
	default:
		return errors.New("不支持的活动记录类型")
	}

	return r.db.WithContext(ctx).Create(modelActivity).Error
}

// GetUserActivities 获取用户活动记录
func (r *PostgresRepository) GetUserActivities(ctx context.Context, userID string, startTime, endTime time.Time, offset, limit int) ([]*models.UserActivityRecord, int64, error) {
	var activities []*models.UserActivityRecord
	var count int64

	query := r.db.WithContext(ctx).Model(&models.UserActivityRecord{})

	if !startTime.IsZero() {
		query = query.Where("created_at >= ?", startTime)
	}

	if !endTime.IsZero() {
		query = query.Where("created_at <= ?", endTime)
	}

	err := query.Where("user_id = ?", userID).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&activities).Error
	if err != nil {
		return nil, 0, err
	}

	return activities, count, nil
}

// GetRecentUserActivities 获取用户最近活动记录
func (r *PostgresRepository) GetRecentUserActivities(ctx context.Context, userID string, limit int) ([]*models.UserActivityRecord, error) {
	var activities []*models.UserActivityRecord
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&activities).Error
	if err != nil {
		return nil, err
	}
	return activities, nil
}

// CreateRecommendationItem 创建推荐项
func (r *PostgresRepository) CreateRecommendationItem(ctx context.Context, item interface{}) error {
	// 将item转换为models.RecommendationItem
	var modelItem *models.RecommendationItem

	switch i := item.(type) {
	case *models.RecommendationItem:
		modelItem = i
	case *SimpleRecommendationItem:
		modelItem = &models.RecommendationItem{
			ID:        i.ID,
			UserID:    i.UserID,
			ItemType:  i.ItemType,
			ItemID:    i.ItemID,
			Score:     i.Score,
			ModelID:   i.ModelID,
			Reason:    i.Reason,
			ExpiresAt: i.ExpiresAt,
			IsClicked: i.IsClicked,
			IsViewed:  i.IsViewed,
			CreatedAt: i.CreatedAt,
		}
	default:
		return errors.New("不支持的推荐项类型")
	}

	return r.db.WithContext(ctx).Create(modelItem).Error
}

// GetRecommendationItemByID 通过ID获取推荐项
func (r *PostgresRepository) GetRecommendationItemByID(ctx context.Context, id string) (*models.RecommendationItem, error) {
	var item models.RecommendationItem
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// GetUserRecommendations 获取用户推荐
func (r *PostgresRepository) GetUserRecommendations(ctx context.Context, userID string, contentType string, limit int) ([]*models.RecommendationItem, error) {
	var items []*models.RecommendationItem
	query := r.db.WithContext(ctx).
		Where("user_id = ? AND item_type = ? AND expires_at > ? AND is_clicked = ?",
			userID, contentType, time.Now(), false).
		Order("score DESC").
		Limit(limit)

	err := query.Find(&items).Error
	if err != nil {
		return nil, err
	}

	return items, nil
}

// UpdateRecommendationItem 更新推荐项
func (r *PostgresRepository) UpdateRecommendationItem(ctx context.Context, item *models.RecommendationItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

// DeleteRecommendationItem 删除推荐项
func (r *PostgresRepository) DeleteRecommendationItem(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.RecommendationItem{}, "id = ?", id).Error
}

// DeleteExpiredRecommendations 删除过期的推荐
func (r *PostgresRepository) DeleteExpiredRecommendations(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&models.RecommendationItem{})
	return result.RowsAffected, result.Error
}

// CreateRecommendationModel 创建推荐模型
func (r *PostgresRepository) CreateRecommendationModel(ctx context.Context, model *models.RecommendationModel) error {
	return r.db.WithContext(ctx).Create(model).Error
}

// GetRecommendationModel 获取推荐模型
func (r *PostgresRepository) GetRecommendationModel(ctx context.Context, id string) (*models.RecommendationModel, error) {
	var model models.RecommendationModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrModelNotFound
	}
	if err != nil {
		return nil, err
	}
	return &model, nil
}

// GetRecommendationModelByName 通过名称获取推荐模型
func (r *PostgresRepository) GetRecommendationModelByName(ctx context.Context, name string) (*models.RecommendationModel, error) {
	var model models.RecommendationModel
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrModelNotFound
	}
	if err != nil {
		return nil, err
	}
	return &model, nil
}

// UpdateRecommendationModel 更新推荐模型
func (r *PostgresRepository) UpdateRecommendationModel(ctx context.Context, model *models.RecommendationModel) error {
	return r.db.WithContext(ctx).Save(model).Error
}

// DeleteRecommendationModel 删除推荐模型
func (r *PostgresRepository) DeleteRecommendationModel(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.RecommendationModel{}, "id = ?", id).Error
}

// ListRecommendationModels 列出推荐模型
func (r *PostgresRepository) ListRecommendationModels(ctx context.Context, isActive bool) ([]*models.RecommendationModel, error) {
	var models []*models.RecommendationModel
	query := r.db.WithContext(ctx)

	if isActive {
		query = query.Where("is_active = ?", true)
	}

	err := query.Find(&models).Error
	if err != nil {
		return nil, err
	}
	return models, nil
}

// CreateContentFeature 创建内容特征
func (r *PostgresRepository) CreateContentFeature(ctx context.Context, feature *models.ContentFeature) error {
	return r.db.WithContext(ctx).Create(feature).Error
}

// GetContentFeature 获取内容特征
func (r *PostgresRepository) GetContentFeature(ctx context.Context, contentType string, contentID string) (*models.ContentFeature, error) {
	var feature models.ContentFeature
	err := r.db.WithContext(ctx).
		Where("content_type = ? AND content_id = ?", contentType, contentID).
		First(&feature).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrFeatureNotFound
	}
	if err != nil {
		return nil, err
	}
	return &feature, nil
}

// UpdateContentFeature 更新内容特征
func (r *PostgresRepository) UpdateContentFeature(ctx context.Context, feature *models.ContentFeature) error {
	return r.db.WithContext(ctx).Save(feature).Error
}

// DeleteContentFeature 删除内容特征
func (r *PostgresRepository) DeleteContentFeature(ctx context.Context, contentType string, contentID string) error {
	return r.db.WithContext(ctx).
		Where("content_type = ? AND content_id = ?", contentType, contentID).
		Delete(&models.ContentFeature{}).Error
}

// GetSimilarContentFeatures 获取相似内容特征
// 注意：这是一个简化实现，实际应用中通常需要向量相似度搜索
// 可能需要使用PostgresSQL的扩展如pg_vector或外部服务如Milvus、Faiss等
func (r *PostgresRepository) GetSimilarContentFeatures(ctx context.Context, contentType string, featureVector models.FeatureVector, limit int) ([]*models.ContentFeature, error) {
	// 这里是简化实现，仅返回相同类型的内容
	// 实际应用中应当基于特征向量计算相似度
	var contentFeatures []*models.ContentFeature
	err := r.db.WithContext(ctx).
		Where("content_type = ?", contentType).
		Limit(limit).
		Find(&contentFeatures).Error
	if err != nil {
		return nil, err
	}
	return contentFeatures, nil
}

// GetUserFeed 获取用户推荐流
func (r *PostgresRepository) GetUserFeed(ctx context.Context, userID string) (*models.UserFeed, error) {
	var feed models.UserFeed
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&feed).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil // 返回nil而不是错误，表示Feed不存在
	}
	if err != nil {
		return nil, err
	}
	return &feed, nil
}

// CreateUserFeed 创建用户推荐流
func (r *PostgresRepository) CreateUserFeed(ctx context.Context, feed interface{}) error {
	// 将feed转换为models.UserFeed
	var modelFeed *models.UserFeed

	switch f := feed.(type) {
	case *models.UserFeed:
		modelFeed = f
	case *SimpleUserFeed:
		modelFeed = &models.UserFeed{
			ID:          f.ID,
			UserID:      f.UserID,
			Items:       f.Items,
			LastUpdated: f.LastUpdated,
			ExpiresAt:   f.ExpiresAt,
			CreatedAt:   f.CreatedAt,
		}
	default:
		return errors.New("不支持的用户Feed类型")
	}

	return r.db.WithContext(ctx).Create(modelFeed).Error
}

// UpdateUserFeed 更新用户推荐流
func (r *PostgresRepository) UpdateUserFeed(ctx context.Context, feed *models.UserFeed) error {
	return r.db.WithContext(ctx).Save(feed).Error
}

// CreateUserABTestGroup 创建用户A/B测试分组
func (r *PostgresRepository) CreateUserABTestGroup(ctx context.Context, group *models.UserABTestGroup) error {
	return r.db.WithContext(ctx).Create(group).Error
}

// GetUserABTestGroup 获取用户A/B测试分组
func (r *PostgresRepository) GetUserABTestGroup(ctx context.Context, userID string, testID string) (*models.UserABTestGroup, error) {
	var group models.UserABTestGroup
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND test_id = ?", userID, testID).
		First(&group).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil // 返回nil而不是错误，表示分组不存在
	}
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// GetUserABTestGroups 获取用户所有A/B测试分组
func (r *PostgresRepository) GetUserABTestGroups(ctx context.Context, userID string) ([]*models.UserABTestGroup, error) {
	var groups []*models.UserABTestGroup
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&groups).Error
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// GetTrendingContent 获取热门内容
func (r *PostgresRepository) GetTrendingContent(ctx context.Context, contentType string, timeRange time.Duration, limit int) ([]*models.RecommendationItem, error) {
	startTime := time.Now().Add(-timeRange)

	// 这里是简化实现，实际应基于多个指标如点击率、分享数、评论数等计算热门内容
	// 可以采用加权评分或自定义算法

	var items []*models.RecommendationItem
	err := r.db.WithContext(ctx).
		Model(&models.RecommendationItem{}).
		Where("item_type = ? AND created_at > ?", contentType, startTime).
		Order("score DESC").
		Limit(limit).
		Find(&items).Error
	if err != nil {
		return nil, err
	}

	return items, nil
}

// GetRecommendationMetrics 获取推荐指标
func (r *PostgresRepository) GetRecommendationMetrics(ctx context.Context, modelID string, startTime, endTime time.Time) (map[string]float64, error) {
	// 指标计算（简化实现）
	metrics := make(map[string]float64)

	// 计算点击率(CTR)
	var totalItems, clickedItems int64

	if err := r.db.WithContext(ctx).
		Model(&models.RecommendationItem{}).
		Where("model_id = ? AND created_at BETWEEN ? AND ?", modelID, startTime, endTime).
		Count(&totalItems).Error; err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).
		Model(&models.RecommendationItem{}).
		Where("model_id = ? AND is_clicked = ? AND created_at BETWEEN ? AND ?",
			modelID, true, startTime, endTime).
		Count(&clickedItems).Error; err != nil {
		return nil, err
	}

	if totalItems > 0 {
		metrics["ctr"] = float64(clickedItems) / float64(totalItems)
	} else {
		metrics["ctr"] = 0
	}

	// 其他指标如转化率、留存率等需要结合具体业务逻辑计算

	return metrics, nil
}

// GetUserInteractionStats 获取用户交互统计
func (r *PostgresRepository) GetUserInteractionStats(ctx context.Context, userID string, startTime, endTime time.Time) (map[string]float64, error) {
	stats := make(map[string]float64)

	// 计算各类型活动数量
	activityTypes := []string{"view", "like", "comment", "share", "click"}

	for _, actType := range activityTypes {
		var count int64
		if err := r.db.WithContext(ctx).
			Model(&models.UserActivityRecord{}).
			Where("user_id = ? AND activity_type = ? AND created_at BETWEEN ? AND ?",
				userID, actType, startTime, endTime).
			Count(&count).Error; err != nil {
			return nil, err
		}
		stats[actType+"_count"] = float64(count)
	}

	// 计算平均停留时间
	var totalDuration float64
	var durationCount int64

	rows, err := r.db.WithContext(ctx).
		Model(&models.UserActivityRecord{}).
		Select("duration").
		Where("user_id = ? AND duration > 0 AND created_at BETWEEN ? AND ?",
			userID, startTime, endTime).
		Rows()
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			// 记录关闭错误，但不覆盖可能存在的其他错误
			if err == nil {
				err = cerr
			}
		}
	}()

	for rows.Next() {
		var duration int
		if err := rows.Scan(&duration); err != nil {
			return nil, err
		}
		totalDuration += float64(duration)
		durationCount++
	}

	// 检查rows.Next()循环中是否有错误
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if durationCount > 0 {
		stats["avg_duration"] = totalDuration / float64(durationCount)
	} else {
		stats["avg_duration"] = 0
	}

	return stats, nil
}
