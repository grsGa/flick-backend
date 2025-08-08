package repository

import (
	"context"
	"backend/services/bookmark/proto"
)

// BookmarkRepository 定义书签仓储接口
type BookmarkRepository interface {
	// CreateBookmark 创建书签
	CreateBookmark(ctx context.Context, bookmark *proto.Bookmark) error
	
	// DeleteBookmark 删除书签
	DeleteBookmark(ctx context.Context, userID, postID string) error
	
	// IsBookmarked 检查是否已收藏
	IsBookmarked(ctx context.Context, userID, postID string) (bool, error)
	
	// ListBookmarks 获取用户书签列表
	ListBookmarks(ctx context.Context, userID string, page, pageSize int32) ([]*proto.Bookmark, int32, error)
}