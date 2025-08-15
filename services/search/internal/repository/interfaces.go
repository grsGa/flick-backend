package repository

import (
	"context"
	"github.com/flick/backend/services/search/proto"
)

// SearchRepository 定义搜索仓储接口
type SearchRepository interface {
	// SearchContent 搜索内容
	SearchContent(ctx context.Context, query string, page, pageSize int32, sortBy string) ([]*proto.SearchResultItem, int32, error)
	
	// SearchUsers 搜索用户
	SearchUsers(ctx context.Context, query string, page, pageSize int32) ([]*proto.SearchResultItem, int32, error)
	
	// SearchHashtags 搜索标签
	SearchHashtags(ctx context.Context, query string, page, pageSize int32) ([]*proto.SearchResultItem, int32, error)
}
