package service

import (
	"context"

	"github.com/flick/backend/services/notification/proto"
)

// NotificationService 定义通知服务接口
type NotificationService interface {
	// CreateNotification 创建通知
	CreateNotification(ctx context.Context, req *proto.CreateNotificationRequest) (*proto.CreateNotificationResponse, error)

	// ListNotifications 获取用户通知列表
	ListNotifications(ctx context.Context, req *proto.ListNotificationsRequest) (*proto.ListNotificationsResponse, error)

	// MarkAsRead 标记通知为已读
	MarkAsRead(ctx context.Context, req *proto.MarkAsReadRequest) (*proto.MarkAsReadResponse, error)

	// MarkAllAsRead 标记所有通知为已读
	MarkAllAsRead(ctx context.Context, req *proto.MarkAllAsReadRequest) (*proto.MarkAllAsReadResponse, error)

	// DeleteNotification 删除通知
	DeleteNotification(ctx context.Context, req *proto.DeleteNotificationRequest) (*proto.DeleteNotificationResponse, error)

	// GetUnreadCount 获取未读通知数
	GetUnreadCount(ctx context.Context, req *proto.GetUnreadCountRequest) (*proto.GetUnreadCountResponse, error)
}
