package service

import (
	"context"
	"backend/services/search/proto"
)

// SearchService 定义搜索服务接口
type SearchService interface {
	// SearchContent 搜索内容
	SearchContent(ctx context.Context, req *proto.SearchContentRequest) (*proto.SearchContentResponse, error)
	
	// SearchUsers 搜索用户
	SearchUsers(ctx context.Context, req *proto.SearchUsersRequest) (*proto.SearchUsersResponse, error)
	
	// SearchHashtags 搜索标签
	SearchHashtags(ctx context.Context, req *proto.SearchHashtagsRequest) (*proto.SearchHashtagsResponse, error)
}