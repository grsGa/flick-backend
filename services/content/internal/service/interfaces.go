package service

import (
	"context"
	"backend/services/content/proto"
)

// ContentService 定义内容服务接口
type ContentService interface {
	// GetContent 获取内容
	GetContent(ctx context.Context, req *proto.GetContentRequest) (*proto.GetContentResponse, error)
	
	// CreateContent 创建内容
	CreateContent(ctx context.Context, req *proto.CreateContentRequest) (*proto.CreateContentResponse, error)
	
	// UpdateContent 更新内容
	UpdateContent(ctx context.Context, req *proto.UpdateContentRequest) (*proto.UpdateContentResponse, error)
	
	// DeleteContent 删除内容
	DeleteContent(ctx context.Context, req *proto.DeleteContentRequest) (*proto.DeleteContentResponse, error)
	
	// ListContent 列出内容
	ListContent(ctx context.Context, req *proto.ListContentRequest) (*proto.ListContentResponse, error)
}