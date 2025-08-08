package service

import (
	"context"
	"backend/services/bookmark/proto"
)

// BookmarkService 定义书签服务接口
type BookmarkService interface {
	// CreateBookmark 创建书签
	CreateBookmark(ctx context.Context, req *proto.CreateBookmarkRequest) (*proto.CreateBookmarkResponse, error)
	
	// DeleteBookmark 删除书签
	DeleteBookmark(ctx context.Context, req *proto.DeleteBookmarkRequest) (*proto.DeleteBookmarkResponse, error)
	
	// IsBookmarked 检查是否已收藏
	IsBookmarked(ctx context.Context, req *proto.IsBookmarkedRequest) (*proto.IsBookmarkedResponse, error)
	
	// ListBookmarks 获取用户书签列表
	ListBookmarks(ctx context.Context, req *proto.ListBookmarksRequest) (*proto.ListBookmarksResponse, error)
}