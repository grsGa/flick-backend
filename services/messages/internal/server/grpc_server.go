package server

import (
	"context"
	"net"

	"github.com/flick/backend/services/messages/internal/service"
	"github.com/flick/backend/services/messages/proto"
	"google.golang.org/grpc"
)

// grpcServer gRPC服务实现
type GRPCServer struct {
	proto.UnimplementedMessageServiceServer
	messageService *service.MessageService
}

// NewGRPCServer 创建gRPC服务实例
func NewGRPCServer(messageService *service.MessageService) *GRPCServer {
	return &GRPCServer{
		messageService: messageService,
	}
}

// CreateConversation 实现创建会话接口
func (s *GRPCServer) CreateConversation(ctx context.Context, req *proto.CreateConversationRequest) (*proto.CreateConversationResponse, error) {
	return s.messageService.CreateConversation(ctx, req)
}

// ListConversations 实现获取会话列表接口
func (s *GRPCServer) ListConversations(ctx context.Context, req *proto.ListConversationsRequest) (*proto.ListConversationsResponse, error) {
	return s.messageService.ListConversations(ctx, req)
}

// GetConversation 实现获取会话详情接口
func (s *GRPCServer) GetConversation(ctx context.Context, req *proto.GetConversationRequest) (*proto.GetConversationResponse, error) {
	return s.messageService.GetConversation(ctx, req)
}

// SendMessage 实现发送消息接口
func (s *GRPCServer) SendMessage(ctx context.Context, req *proto.SendMessageRequest) (*proto.SendMessageResponse, error) {
	return s.messageService.SendMessage(ctx, req)
}

// ListMessages 实现获取消息列表接口
func (s *GRPCServer) ListMessages(ctx context.Context, req *proto.ListMessagesRequest) (*proto.ListMessagesResponse, error) {
	return s.messageService.ListMessages(ctx, req)
}

// MarkAsRead 实现标记消息为已读接口
func (s *GRPCServer) MarkAsRead(ctx context.Context, req *proto.MarkAsReadRequest) (*proto.MarkAsReadResponse, error) {
	return s.messageService.MarkAsRead(ctx, req)
}

// DeleteConversation 实现删除会话接口
func (s *GRPCServer) DeleteConversation(ctx context.Context, req *proto.DeleteConversationRequest) (*proto.DeleteConversationResponse, error) {
	return s.messageService.DeleteConversation(ctx, req)
}

// Run 启动gRPC服务
func (s *GRPCServer) Run(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	proto.RegisterMessageServiceServer(grpcServer, s)

	return grpcServer.Serve(lis)
}

