package repository

import (
	"context"

	"github.com/flick/backend/services/notification/proto"
)

// NotificationRepository 定义通知仓储接口
type NotificationRepository interface {
	// CreateNotification 创建通知
	CreateNotification(ctx context.Context, notification *proto.Notification) error

	// GetNotificationByID 根据ID获取通知
	GetNotificationByID(ctx context.Context, id string) (*proto.Notification, error)

	// ListNotifications 列出用户通知
	ListNotifications(ctx context.Context, userID string, unreadOnly bool, page, pageSize int32) ([]*proto.Notification, int32, error)

	// MarkAsRead 标记通知为已读
	MarkAsRead(ctx context.Context, notificationID string) error

	// MarkAllAsRead 标记所有通知为已读
	MarkAllAsRead(ctx context.Context, userID string) (int32, error)

	// DeleteNotification 删除通知
	DeleteNotification(ctx context.Context, id string) error

	// GetUnreadCount 获取未读通知数
	GetUnreadCount(ctx context.Context, userID string) (int32, error)
}
