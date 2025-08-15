package service

import (
	"context"
	"github.com/flick/backend/services/search/internal/repository"
	"github.com/flick/backend/services/search/proto"
)

// searchService 搜索服务实现
type searchService struct {
	searchRepo repository.SearchRepository
}

// NewSearchService 创建搜索服务实例
func NewSearchService(searchRepo repository.SearchRepository) SearchService {
	return &searchService{
		searchRepo: searchRepo,
	}
}

// SearchContent 搜索内容
func (s *searchService) SearchContent(ctx context.Context, req *proto.SearchContentRequest) (*proto.SearchContentResponse, error) {
	items, total, err := s.searchRepo.SearchContent(ctx, req.Query, req.Page, req.PageSize, req.SortBy)
	if err != nil {
		return &proto.SearchContentResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to search content: " + err.Error(),
			},
		}, err
	}
	
	return &proto.SearchContentResponse{
		Items: items,
		Total: total,
	}, nil
}

// SearchUsers 搜索用户
func (s *searchService) SearchUsers(ctx context.Context, req *proto.SearchUsersRequest) (*proto.SearchUsersResponse, error) {
	items, total, err := s.searchRepo.SearchUsers(ctx, req.Query, req.Page, req.PageSize)
	if err != nil {
		return &proto.SearchUsersResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to search users: " + err.Error(),
			},
		}, err
	}
	
	return &proto.SearchUsersResponse{
		Items: items,
		Total: total,
	}, nil
}

// SearchHashtags 搜索标签
func (s *searchService) SearchHashtags(ctx context.Context, req *proto.SearchHashtagsRequest) (*proto.SearchHashtagsResponse, error) {
	items, total, err := s.searchRepo.SearchHashtags(ctx, req.Query, req.Page, req.PageSize)
	if err != nil {
		return &proto.SearchHashtagsResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to search hashtags: " + err.Error(),
			},
		}, err
	}
	
	return &proto.SearchHashtagsResponse{
		Items: items,
		Total: total,
	}, nil
}
