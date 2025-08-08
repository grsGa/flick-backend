package repository

import (
	"context"
	"errors"
	"time"

	"backend/pkg/database"
	"backend/pkg/models"
	"backend/services/notification/proto"

	"gorm.io/gorm"
)

// notificationRepository 通知仓储实现
type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository 创建通知仓储实例
func NewNotificationRepository() NotificationRepository {
	return &notificationRepository{
		db: database.GetDB(),
	}
}

// CreateNotification 创建通知
func (r *notificationRepository) CreateNotification(ctx context.Context, notification *proto.Notification) error {
	createdAt, _ := time.Parse(time.RFC3339, notification.CreatedAt)
	n := &models.Notification{
		ID:         notification.Id,
		ReceiverID: notification.ReceiverId,
		ActorID:    notification.ActorId,
		ActionType: notification.ActionType,
		TargetType: notification.TargetType,
		TargetID:   notification.TargetId,
		IsRead:     notification.IsRead,
		CreatedAt:  createdAt,
	}

	return r.db.Create(n).Error
}

// GetNotificationByID 根据ID获取通知
func (r *notificationRepository) GetNotificationByID(ctx context.Context, id string) (*proto.Notification, error) {
	var notification models.Notification
	if err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&notification).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("notification not found")
		}
		return nil, err
	}

	return &proto.Notification{
		Id:         notification.ID,
		ReceiverId: notification.ReceiverID,
		ActorId:    notification.ActorID,
		ActionType: notification.ActionType,
		TargetType: notification.TargetType,
		TargetId:   notification.TargetID,
		IsRead:     notification.IsRead,
		CreatedAt:  notification.CreatedAt.Format(time.RFC3339),
	}, nil
}

// ListNotifications 列出用户通知
func (r *notificationRepository) ListNotifications(ctx context.Context, userID string, unreadOnly bool, page, pageSize int32) ([]*proto.Notification, int32, error) {
	var notifications []models.Notification
	var total int64

	// 构建查询
	query := r.db.Model(&models.Notification{}).Where("receiver_id = ? AND deleted_at IS NULL", userID)

	if unreadOnly {
		query = query.Where("is_read = false")
	}

	// 查询总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询列表
	offset := (page - 1) * pageSize
	if err := query.Offset(int(offset)).Limit(int(pageSize)).Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	protoNotifications := make([]*proto.Notification, len(notifications))
	for i, notification := range notifications {
		protoNotifications[i] = &proto.Notification{
			Id:         notification.ID,
			ReceiverId: notification.ReceiverID,
			ActorId:    notification.ActorID,
			ActionType: notification.ActionType,
			TargetType: notification.TargetType,
			TargetId:   notification.TargetID,
			IsRead:     notification.IsRead,
			CreatedAt:  notification.CreatedAt.Format(time.RFC3339),
		}
	}

	return protoNotifications, int32(total), nil
}

// MarkAsRead 标记通知为已读
func (r *notificationRepository) MarkAsRead(ctx context.Context, notificationID string) error {
	return r.db.Model(&models.Notification{}).
		Where("id = ? AND deleted_at IS NULL", notificationID).
		Update("is_read", true).
		Update("updated_at", time.Now()).Error
}

// MarkAllAsRead 标记所有通知为已读
func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID string) (int32, error) {
	result := r.db.Model(&models.Notification{}).
		Where("receiver_id = ? AND is_read = false AND deleted_at IS NULL", userID).
		Update("is_read", true).
		Update("updated_at", time.Now())

	if result.Error != nil {
		return 0, result.Error
	}

	return int32(result.RowsAffected), nil
}

// DeleteNotification 删除通知
func (r *notificationRepository) DeleteNotification(ctx context.Context, id string) error {
	// 软删除通知
	return r.db.Where("id = ?", id).Delete(&models.Notification{}).Error
}

// GetUnreadCount 获取未读通知数
func (r *notificationRepository) GetUnreadCount(ctx context.Context, userID string) (int32, error) {
	var count int64
	err := r.db.Model(&models.Notification{}).
		Where("receiver_id = ? AND is_read = false AND deleted_at IS NULL", userID).
		Count(&count).Error

	return int32(count), err
}
