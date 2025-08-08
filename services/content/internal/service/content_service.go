package service

import (
	"context"
	"time"
	
	"github.com/google/uuid"
	"backend/services/content/internal/repository"
	"backend/services/content/proto"
)

// contentService 内容服务实现
type contentService struct {
	contentRepo repository.ContentRepository
}

// NewContentService 创建内容服务实例
func NewContentService(contentRepo repository.ContentRepository) ContentService {
	return &contentService{
		contentRepo: contentRepo,
	}
}

// GetContent 获取内容
func (s *contentService) GetContent(ctx context.Context, req *proto.GetContentRequest) (*proto.GetContentResponse, error) {
	content, err := s.contentRepo.GetContentByID(ctx, req.ContentId)
	if err != nil {
		return &proto.GetContentResponse{
			Error: &proto.Error{
				Code:    404,
				Message: "Content not found: " + err.Error(),
			},
		}, err
	}
	
	return &proto.GetContentResponse{
		Content: content,
	}, nil
}

// CreateContent 创建内容
func (s *contentService) CreateContent(ctx context.Context, req *proto.CreateContentRequest) (*proto.CreateContentResponse, error) {
	// 创建内容对象
	content := &proto.Content{
		Id:        uuid.New().String(),
		UserId:    req.UserId,
		Title:     req.Title,
		Body:      req.Body,
		MediaFiles: req.MediaFiles,
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}
	
	// 保存到数据库
	err := s.contentRepo.CreateContent(ctx, content)
	if err != nil {
		return &proto.CreateContentResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to create content: " + err.Error(),
			},
		}, err
	}
	
	return &proto.CreateContentResponse{
		Content: content,
	}, nil
}

// UpdateContent 更新内容
func (s *contentService) UpdateContent(ctx context.Context, req *proto.UpdateContentRequest) (*proto.UpdateContentResponse, error) {
	// 首先获取现有内容
	content, err := s.contentRepo.GetContentByID(ctx, req.ContentId)
	if err != nil {
		return &proto.UpdateContentResponse{
			Error: &proto.Error{
				Code:    404,
				Message: "Content not found: " + err.Error(),
			},
		}, err
	}
	
	// 更新内容字段
	if req.Title != "" {
		content.Title = req.Title
	}
	
	if req.Body != "" {
		content.Body = req.Body
	}
	
	content.MediaFiles = req.MediaFiles
	content.UpdatedAt = time.Now().Format(time.RFC3339)
	
	// 保存到数据库
	err = s.contentRepo.UpdateContent(ctx, content)
	if err != nil {
		return &proto.UpdateContentResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to update content: " + err.Error(),
			},
		}, err
	}
	
	return &proto.UpdateContentResponse{
		Content: content,
	}, nil
}

// DeleteContent 删除内容
func (s *contentService) DeleteContent(ctx context.Context, req *proto.DeleteContentRequest) (*proto.DeleteContentResponse, error) {
	err := s.contentRepo.DeleteContent(ctx, req.ContentId)
	if err != nil {
		return &proto.DeleteContentResponse{
			Success: false,
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to delete content: " + err.Error(),
			},
		}, err
	}
	
	return &proto.DeleteContentResponse{
		Success: true,
	}, nil
}

// ListContent 列出内容
func (s *contentService) ListContent(ctx context.Context, req *proto.ListContentRequest) (*proto.ListContentResponse, error) {
	contents, total, err := s.contentRepo.ListContent(ctx, req.UserId, req.Page, req.PageSize)
	if err != nil {
		return &proto.ListContentResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to list content: " + err.Error(),
			},
		}, err
	}
	
	return &proto.ListContentResponse{
		Contents: contents,
		Total:    total,
	}, nil
}