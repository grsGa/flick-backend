package server

import (
	"context"
	"net"

	"github.com/flick/backend/services/notification/internal/service"
	"github.com/flick/backend/services/notification/proto"
	"google.golang.org/grpc"
)

// grpcServer gRPC服务实现
type grpcServer struct {
	proto.UnimplementedNotificationServiceServer
	notificationService service.NotificationService
}

// NewGRPCServer 创建gRPC服务实例
func NewGRPCServer(notificationService service.NotificationService) *grpcServer {
	return &grpcServer{
		notificationService: notificationService,
	}
}

// CreateNotification 实现创建通知接口
func (s *grpcServer) CreateNotification(ctx context.Context, req *proto.CreateNotificationRequest) (*proto.CreateNotificationResponse, error) {
	return s.notificationService.CreateNotification(ctx, req)
}

// ListNotifications 实现获取用户通知列表接口
func (s *grpcServer) ListNotifications(ctx context.Context, req *proto.ListNotificationsRequest) (*proto.ListNotificationsResponse, error) {
	return s.notificationService.ListNotifications(ctx, req)
}

// MarkAsRead 实现标记通知为已读接口
func (s *grpcServer) MarkAsRead(ctx context.Context, req *proto.MarkAsReadRequest) (*proto.MarkAsReadResponse, error) {
	return s.notificationService.MarkAsRead(ctx, req)
}

// MarkAllAsRead 实现标记所有通知为已读接口
func (s *grpcServer) MarkAllAsRead(ctx context.Context, req *proto.MarkAllAsReadRequest) (*proto.MarkAllAsReadResponse, error) {
	return s.notificationService.MarkAllAsRead(ctx, req)
}

// DeleteNotification 实现删除通知接口
func (s *grpcServer) DeleteNotification(ctx context.Context, req *proto.DeleteNotificationRequest) (*proto.DeleteNotificationResponse, error) {
	return s.notificationService.DeleteNotification(ctx, req)
}

// GetUnreadCount 实现获取未读通知数接口
func (s *grpcServer) GetUnreadCount(ctx context.Context, req *proto.GetUnreadCountRequest) (*proto.GetUnreadCountResponse, error) {
	return s.notificationService.GetUnreadCount(ctx, req)
}

// Run 启动gRPC服务
func (s *grpcServer) Run(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	proto.RegisterNotificationServiceServer(grpcServer, s)

	return grpcServer.Serve(lis)
}
