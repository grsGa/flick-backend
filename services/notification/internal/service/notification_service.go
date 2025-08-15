package service

import (
	"context"
	"time"

	"github.com/flick/backend/services/notification/internal/repository"
	"github.com/flick/backend/services/notification/proto"
	"github.com/google/uuid"
)

// notificationService 通知服务实现
type notificationService struct {
	notificationRepo repository.NotificationRepository
}

// NewNotificationService 创建通知服务实例
func NewNotificationService(notificationRepo repository.NotificationRepository) NotificationService {
	return &notificationService{
		notificationRepo: notificationRepo,
	}
}

// CreateNotification 创建通知
func (s *notificationService) CreateNotification(ctx context.Context, req *proto.CreateNotificationRequest) (*proto.CreateNotificationResponse, error) {
	// 创建通知对象
	notification := &proto.Notification{
		Id:         uuid.New().String(),
		ReceiverId: req.ReceiverId,
		ActorId:    req.ActorId,
		ActionType: req.ActionType,
		TargetType: req.TargetType,
		TargetId:   req.TargetId,
		Content:    req.Content,
		IsRead:     false,
		CreatedAt:  time.Now().Format(time.RFC3339),
	}

	// 保存到数据库
	err := s.notificationRepo.CreateNotification(ctx, notification)
	if err != nil {
		return &proto.CreateNotificationResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to create notification: " + err.Error(),
			},
		}, err
	}

	return &proto.CreateNotificationResponse{
		Notification: notification,
	}, nil
}

// ListNotifications 获取用户通知列表
func (s *notificationService) ListNotifications(ctx context.Context, req *proto.ListNotificationsRequest) (*proto.ListNotificationsResponse, error) {
	notifications, total, err := s.notificationRepo.ListNotifications(ctx, req.UserId, req.UnreadOnly, req.Page, req.PageSize)
	if err != nil {
		return &proto.ListNotificationsResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to list notifications: " + err.Error(),
			},
		}, err
	}

	return &proto.ListNotificationsResponse{
		Notifications: notifications,
		Total:         total,
	}, nil
}

// MarkAsRead 标记通知为已读
func (s *notificationService) MarkAsRead(ctx context.Context, req *proto.MarkAsReadRequest) (*proto.MarkAsReadResponse, error) {
	err := s.notificationRepo.MarkAsRead(ctx, req.NotificationId)
	if err != nil {
		return &proto.MarkAsReadResponse{
			Success: false,
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to mark notification as read: " + err.Error(),
			},
		}, err
	}

	return &proto.MarkAsReadResponse{
		Success: true,
	}, nil
}

// MarkAllAsRead 标记所有通知为已读
func (s *notificationService) MarkAllAsRead(ctx context.Context, req *proto.MarkAllAsReadRequest) (*proto.MarkAllAsReadResponse, error) {
	count, err := s.notificationRepo.MarkAllAsRead(ctx, req.UserId)
	if err != nil {
		return &proto.MarkAllAsReadResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to mark all notifications as read: " + err.Error(),
			},
		}, err
	}

	return &proto.MarkAllAsReadResponse{
		MarkedCount: count,
	}, nil
}

// DeleteNotification 删除通知
func (s *notificationService) DeleteNotification(ctx context.Context, req *proto.DeleteNotificationRequest) (*proto.DeleteNotificationResponse, error) {
	err := s.notificationRepo.DeleteNotification(ctx, req.NotificationId)
	if err != nil {
		return &proto.DeleteNotificationResponse{
			Success: false,
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to delete notification: " + err.Error(),
			},
		}, err
	}

	return &proto.DeleteNotificationResponse{
		Success: true,
	}, nil
}

// GetUnreadCount 获取未读通知数
func (s *notificationService) GetUnreadCount(ctx context.Context, req *proto.GetUnreadCountRequest) (*proto.GetUnreadCountResponse, error) {
	count, err := s.notificationRepo.GetUnreadCount(ctx, req.UserId)
	if err != nil {
		return &proto.GetUnreadCountResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to get unread count: " + err.Error(),
			},
		}, err
	}

	return &proto.GetUnreadCountResponse{
		Count: count,
	}, nil
}
