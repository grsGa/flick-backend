package repository

import (
	"context"
	"backend/services/content/proto"
)

// ContentRepository 定义内容仓储接口
type ContentRepository interface {
	// GetContentByID 根据ID获取内容
	GetContentByID(ctx context.Context, id string) (*proto.Content, error)
	
	// CreateContent 创建内容
	CreateContent(ctx context.Context, content *proto.Content) error
	
	// UpdateContent 更新内容
	UpdateContent(ctx context.Context, content *proto.Content) error
	
	// DeleteContent 删除内容
	DeleteContent(ctx context.Context, id string) error
	
	// ListContent 列出内容
	ListContent(ctx context.Context, userID string, page, pageSize int32) ([]*proto.Content, int32, error)
}