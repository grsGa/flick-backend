package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend/pkg/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrNotificationNotFound = errors.New("通知未找到")
	ErrTemplateNotFound     = errors.New("通知模板未找到")
	ErrDeviceNotFound       = errors.New("设备未找到")
)

// PostgresRepository 是NotificationRepository的PostgresSQL实现
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository 创建一个新的PostgresSQL通知仓库
func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

// CreateNotification 创建新通知
func (r *PostgresRepository) CreateNotification(ctx context.Context, notification *models.Notification) error {
	if notification.ID == "" {
		notification.ID = uuid.New().String()
	}

	now := time.Now()
	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = now
	}
	if notification.UpdatedAt.IsZero() {
		notification.UpdatedAt = now
	}

	err := r.db.WithContext(ctx).Create(notification).Error
	if err != nil {
		return fmt.Errorf("创建通知失败: %w", err)
	}

	// 更新用户未读通知计数
	if err := r.updateUnreadCount(ctx, notification.UserID); err != nil {
		return fmt.Errorf("更新未读通知计数失败: %w", err)
	}

	return nil
}

// CreateBulkNotifications 批量创建通知
func (r *PostgresRepository) CreateBulkNotifications(ctx context.Context, notifications []*models.Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	err := r.db.WithContext(ctx).Create(&notifications).Error
	if err != nil {
		return fmt.Errorf("批量创建通知失败: %w", err)
	}

	// 获取所有接收者ID
	userIDs := make(map[string]struct{})
	for _, n := range notifications {
		userIDs[n.UserID] = struct{}{}
	}

	// 更新每个用户的未读通知计数
	for userID := range userIDs {
		if err := r.updateUnreadCount(ctx, userID); err != nil {
			return fmt.Errorf("更新用户 %s 的未读通知计数失败: %w", userID, err)
		}
	}

	return nil
}

// GetNotificationByID 通过ID获取通知
func (r *PostgresRepository) GetNotificationByID(ctx context.Context, id string) (*models.Notification, error) {
	var notification models.Notification
	err := r.db.WithContext(ctx).
		Preload("Actor").
		Where("id = ?", id).
		First(&notification).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotificationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("获取通知失败: %w", err)
	}

	return &notification, nil
}

// GetUserNotifications 获取用户通知列表
func (r *PostgresRepository) GetUserNotifications(ctx context.Context, userID string, page, pageSize int) ([]*models.Notification, int64, error) {
	offset := (page - 1) * pageSize

	var total int64
	if err := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取通知数量失败: %w", err)
	}

	notifications := []*models.Notification{}
	if err := r.db.WithContext(ctx).
		Preload("Actor").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&notifications).Error; err != nil {
		return nil, 0, fmt.Errorf("获取通知列表失败: %w", err)
	}

	return notifications, total, nil
}

// GetUserNotificationsByType 按类型获取用户通知
func (r *PostgresRepository) GetUserNotificationsByType(ctx context.Context, userID string, notificationType string, page, pageSize int) ([]*models.Notification, int64, error) {
	offset := (page - 1) * pageSize

	var total int64
	if err := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ? AND notification_type = ?", userID, notificationType).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取通知数量失败: %w", err)
	}

	notifications := []*models.Notification{}
	if err := r.db.WithContext(ctx).
		Preload("Actor").
		Where("user_id = ? AND notification_type = ?", userID, notificationType).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&notifications).Error; err != nil {
		return nil, 0, fmt.Errorf("获取通知列表失败: %w", err)
	}

	return notifications, total, nil
}

// GetUnreadNotificationCount 获取用户未读通知数量
func (r *PostgresRepository) GetUnreadNotificationCount(ctx context.Context, userID string) (int64, error) {
	var count models.NotificationCount
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&count).Error

	// 如果未找到记录，返回0
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("获取未读通知计数失败: %w", err)
	}

	return int64(count.UnreadCount), nil
}

// MarkNotificationAsRead 标记单个通知为已读
func (r *PostgresRepository) MarkNotificationAsRead(ctx context.Context, notificationID string) error {
	var notification models.Notification
	// 先查找通知，获取用户ID
	err := r.db.WithContext(ctx).
		Where("id = ?", notificationID).
		First(&notification).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotificationNotFound
	}
	if err != nil {
		return fmt.Errorf("获取通知失败: %w", err)
	}

	// 更新状态为已读
	now := time.Now()
	err = r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id = ? AND status = ?", notificationID, models.NotificationStatusUnread).
		Updates(map[string]interface{}{
			"status":  models.NotificationStatusRead,
			"read_at": now,
		}).Error
	if err != nil {
		return fmt.Errorf("标记通知为已读失败: %w", err)
	}

	// 更新未读计数
	if err := r.updateUnreadCount(ctx, notification.UserID); err != nil {
		return fmt.Errorf("更新未读通知计数失败: %w", err)
	}

	return nil
}

// MarkNotificationsAsRead 批量标记通知为已读
func (r *PostgresRepository) MarkNotificationsAsRead(ctx context.Context, notificationIDs []string) error {
	if len(notificationIDs) == 0 {
		return nil
	}

	// 查找涉及的所有用户ID
	var notifications []models.Notification
	err := r.db.WithContext(ctx).
		Where("id IN ? AND status = ?", notificationIDs, models.NotificationStatusUnread).
		Find(&notifications).Error
	if err != nil {
		return fmt.Errorf("获取通知列表失败: %w", err)
	}

	// 如果没有找到未读通知，直接返回
	if len(notifications) == 0 {
		return nil
	}

	// 收集所有用户ID
	userIDs := make(map[string]struct{})
	for _, n := range notifications {
		userIDs[n.UserID] = struct{}{}
	}

	// 批量更新通知状态
	now := time.Now()
	err = r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id IN ? AND status = ?", notificationIDs, models.NotificationStatusUnread).
		Updates(map[string]interface{}{
			"status":  models.NotificationStatusRead,
			"read_at": now,
		}).Error
	if err != nil {
		return fmt.Errorf("批量标记通知为已读失败: %w", err)
	}

	// 为每个用户更新未读计数
	for userID := range userIDs {
		if err := r.updateUnreadCount(ctx, userID); err != nil {
			return fmt.Errorf("更新用户 %s 的未读通知计数失败: %w", userID, err)
		}
	}

	return nil
}

// MarkAllNotificationsAsRead 标记用户所有通知为已读
func (r *PostgresRepository) MarkAllNotificationsAsRead(ctx context.Context, userID string) error {
	now := time.Now()
	err := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ? AND status = ?", userID, models.NotificationStatusUnread).
		Updates(map[string]interface{}{
			"status":  models.NotificationStatusRead,
			"read_at": now,
		}).Error
	if err != nil {
		return fmt.Errorf("标记所有通知为已读失败: %w", err)
	}

	// 更新未读计数为0
	err = r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Assign(models.NotificationCount{
			UserID:      userID,
			UnreadCount: 0,
			UpdatedAt:   time.Now(),
		}).
		FirstOrCreate(&models.NotificationCount{}).Error
	if err != nil {
		return fmt.Errorf("重置未读通知计数失败: %w", err)
	}

	return nil
}

// DeleteNotification 删除通知
func (r *PostgresRepository) DeleteNotification(ctx context.Context, notificationID string) error {
	var notification models.Notification
	// 先查找通知，获取用户ID
	err := r.db.WithContext(ctx).
		Where("id = ?", notificationID).
		First(&notification).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotificationNotFound
	}
	if err != nil {
		return fmt.Errorf("获取通知失败: %w", err)
	}

	// 软删除通知
	err = r.db.WithContext(ctx).Delete(&models.Notification{}, notificationID).Error
	if err != nil {
		return fmt.Errorf("删除通知失败: %w", err)
	}

	// 如果删除的是未读通知，更新计数
	if notification.Status == models.NotificationStatusUnread {
		if err := r.updateUnreadCount(ctx, notification.UserID); err != nil {
			return fmt.Errorf("更新未读通知计数失败: %w", err)
		}
	}

	return nil
}

// DeleteUserNotifications 删除用户所有通知
func (r *PostgresRepository) DeleteUserNotifications(ctx context.Context, userID string) error {
	// 软删除用户的所有通知
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.Notification{}).Error
	if err != nil {
		return fmt.Errorf("删除用户通知失败: %w", err)
	}

	// 更新未读计数为0
	err = r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Assign(models.NotificationCount{
			UserID:      userID,
			UnreadCount: 0,
			UpdatedAt:   time.Now(),
		}).
		FirstOrCreate(&models.NotificationCount{}).Error
	if err != nil {
		return fmt.Errorf("重置未读通知计数失败: %w", err)
	}

	return nil
}

// GetUserNotificationSettings 获取用户通知设置
func (r *PostgresRepository) GetUserNotificationSettings(ctx context.Context, userID string) (*models.NotificationPreference, error) {
	var settings models.NotificationPreference
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&settings).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 创建默认设置
		settings = models.NotificationPreference{
			UserID:            userID,
			LikeEnabled:       true,
			CommentEnabled:    true,
			MentionEnabled:    true,
			FollowEnabled:     true,
			SystemEnabled:     true,
			MessageEnabled:    true,
			ShareEnabled:      true,
			EmailEnabled:      true,
			PushEnabled:       true,
			QuietHoursEnabled: false,
			QuietHoursStart:   "22:00",
			QuietHoursEnd:     "08:00",
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		err := r.CreateUserNotificationSettings(ctx, &settings)
		if err != nil {
			return nil, fmt.Errorf("创建默认通知设置失败: %w", err)
		}

		return &settings, nil
	}

	if err != nil {
		return nil, fmt.Errorf("获取通知设置失败: %w", err)
	}

	return &settings, nil
}

// CreateUserNotificationSettings 创建用户通知设置
func (r *PostgresRepository) CreateUserNotificationSettings(ctx context.Context, settings *models.NotificationPreference) error {
	err := r.db.WithContext(ctx).Create(settings).Error
	if err != nil {
		return fmt.Errorf("创建通知设置失败: %w", err)
	}
	return nil
}

// UpdateUserNotificationSettings 更新用户通知设置
func (r *PostgresRepository) UpdateUserNotificationSettings(ctx context.Context, settings *models.NotificationPreference) error {
	err := r.db.WithContext(ctx).Save(settings).Error
	if err != nil {
		return fmt.Errorf("更新通知设置失败: %w", err)
	}
	return nil
}

// CreateUserDevice 创建用户设备
func (r *PostgresRepository) CreateUserDevice(ctx context.Context, device *models.UserDevice) error {
	if device.ID == "" {
		device.ID = uuid.New().String()
	}
	
	return r.db.WithContext(ctx).Create(device).Error
}

// GetUserDevices 获取用户所有设备
func (r *PostgresRepository) GetUserDevices(ctx context.Context, userID string) ([]*models.UserDevice, error) {
	var devices []*models.UserDevice
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&devices).Error
	if err != nil {
		return nil, fmt.Errorf("获取用户设备失败: %w", err)
	}

	return devices, nil
}

// GetUserDeviceByToken 通过令牌获取用户设备
func (r *PostgresRepository) GetUserDeviceByToken(ctx context.Context, userID string, deviceToken string) (*models.UserDevice, error) {
	var device models.UserDevice
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND device_token = ?", userID, deviceToken).
		First(&device).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrDeviceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("获取用户设备失败: %w", err)
	}

	return &device, nil
}

// UpdateUserDevice 更新用户设备信息
func (r *PostgresRepository) UpdateUserDevice(ctx context.Context, device *models.UserDevice) error {
	err := r.db.WithContext(ctx).Save(device).Error
	if err != nil {
		return fmt.Errorf("更新用户设备失败: %w", err)
	}
	return nil
}

// DeleteUserDevice 删除用户设备
func (r *PostgresRepository) DeleteUserDevice(ctx context.Context, userID, deviceToken string) error {
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND device_token = ?", userID, deviceToken).
		Delete(&models.UserDevice{}).Error
	if err != nil {
		return fmt.Errorf("删除用户设备失败: %w", err)
	}
	
	return nil
}

// CreateNotificationTemplate 创建通知模板
func (r *PostgresRepository) CreateNotificationTemplate(ctx context.Context, template *models.NotificationTemplate) error {
	// 检查是否已存在同类型模板
	var existingTemplate models.NotificationTemplate
	err := r.db.WithContext(ctx).
		Where("type = ?", template.Type).
		First(&existingTemplate).Error

	// 如果已存在，则更新
	if err == nil {
		template.ID = existingTemplate.ID
		return r.UpdateNotificationTemplate(ctx, template)
	}

	// 如果是其他错误，返回错误
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("检查模板是否存在时出错: %w", err)
	}

	// 创建新模板
	err = r.db.WithContext(ctx).Create(template).Error
	if err != nil {
		return fmt.Errorf("创建通知模板失败: %w", err)
	}

	return nil
}

// GetNotificationTemplate 获取通知模板
func (r *PostgresRepository) GetNotificationTemplate(ctx context.Context, templateType string) (*models.NotificationTemplate, error) {
	var template models.NotificationTemplate
	err := r.db.WithContext(ctx).
		Where("type = ?", templateType).
		First(&template).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTemplateNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("获取通知模板失败: %w", err)
	}

	return &template, nil
}

// UpdateNotificationTemplate 更新通知模板
func (r *PostgresRepository) UpdateNotificationTemplate(ctx context.Context, template *models.NotificationTemplate) error {
	result := r.db.WithContext(ctx).Save(template)
	if result.Error != nil {
		return fmt.Errorf("更新通知模板失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrTemplateNotFound
	}
	return nil
}

// DeleteNotificationTemplate 删除通知模板
func (r *PostgresRepository) DeleteNotificationTemplate(ctx context.Context, templateType string) error {
	result := r.db.WithContext(ctx).
		Where("type = ?", templateType).
		Delete(&models.NotificationTemplate{})

	if result.Error != nil {
		return fmt.Errorf("删除通知模板失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrTemplateNotFound
	}

	return nil
}

// ListNotificationTemplates 获取所有通知模板
func (r *PostgresRepository) ListNotificationTemplates(ctx context.Context) ([]*models.NotificationTemplate, error) {
	var templates []*models.NotificationTemplate
	err := r.db.WithContext(ctx).
		Find(&templates).Error
	if err != nil {
		return nil, fmt.Errorf("获取通知模板列表失败: %w", err)
	}

	return templates, nil
}

// 辅助函数：更新用户未读通知计数
func (r *PostgresRepository) updateUnreadCount(ctx context.Context, userID string) error {
	// 计算未读数量
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ? AND status = ?", userID, models.NotificationStatusUnread).
		Count(&count).Error
	if err != nil {
		return err
	}

	// 更新或创建通知计数记录
	err = r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Assign(models.NotificationCount{
			UserID:      userID,
			UnreadCount: int(count),
			UpdatedAt:   time.Now(),
		}).
		FirstOrCreate(&models.NotificationCount{}).Error

	return err
}

// CreateOrUpdateUserNotificationSettings 创建或更新用户通知设置
func (r *PostgresRepository) CreateOrUpdateUserNotificationSettings(ctx context.Context, settings *models.NotificationPreference) error {
	// 检查是否已存在
	var existingSettings models.NotificationPreference
	err := r.db.WithContext(ctx).
		Where("user_id = ?", settings.UserID).
		First(&existingSettings).Error

	// 如果不存在，创建新设置
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.WithContext(ctx).Create(settings).Error
	}

	// 存在则更新
	if err == nil {
		settings.ID = existingSettings.ID
		return r.db.WithContext(ctx).Save(settings).Error
	}

	return fmt.Errorf("创建或更新通知设置失败: %w", err)
}

// UpdateDeviceNotificationSettings 更新设备通知设置
func (r *PostgresRepository) UpdateDeviceNotificationSettings(ctx context.Context, userID, deviceToken string, enabled bool) error {
	result := r.db.WithContext(ctx).
		Model(&models.UserDevice{}).
		Where("user_id = ? AND device_token = ?", userID, deviceToken).
		Update("notifications_enabled", enabled)
	
	if result.Error != nil {
		return fmt.Errorf("更新设备通知设置失败: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return ErrDeviceNotFound
	}
	
	return nil
}

// CreateOrUpdateNotificationTemplate 创建或更新通知模板
func (r *PostgresRepository) CreateOrUpdateNotificationTemplate(ctx context.Context, template *models.NotificationTemplate) error {
	// 检查是否已存在
	existingTemplate, err := r.GetNotificationTemplate(ctx, string(template.Type))
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			// 创建新模板
			return r.CreateNotificationTemplate(ctx, template)
		}
		return err
	}
	
	// 更新已存在的模板
	template.ID = existingTemplate.ID
	return r.UpdateNotificationTemplate(ctx, template)
}
