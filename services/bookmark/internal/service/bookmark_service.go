package service

import (
	"context"
	"time"
	
	"github.com/google/uuid"
	"backend/services/bookmark/internal/repository"
	"backend/services/bookmark/proto"
)

// bookmarkService 书签服务实现
type bookmarkService struct {
	bookmarkRepo repository.BookmarkRepository
}

// NewBookmarkService 创建书签服务实例
func NewBookmarkService(bookmarkRepo repository.BookmarkRepository) BookmarkService {
	return &bookmarkService{
		bookmarkRepo: bookmarkRepo,
	}
}

// CreateBookmark 创建书签
func (s *bookmarkService) CreateBookmark(ctx context.Context, req *proto.CreateBookmarkRequest) (*proto.CreateBookmarkResponse, error) {
	// 检查是否已经收藏
	isBookmarked, err := s.bookmarkRepo.IsBookmarked(ctx, req.UserId, req.PostId)
	if err != nil {
		return &proto.CreateBookmarkResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to check bookmark status: " + err.Error(),
			},
		}, err
	}
	
	// 如果已经收藏，直接返回
	if isBookmarked {
		return &proto.CreateBookmarkResponse{
			Error: &proto.Error{
				Code:    400,
				Message: "Already bookmarked",
			},
		}, nil
	}
	
	// 创建书签对象
	bookmark := &proto.Bookmark{
		Id:        uuid.New().String(),
		UserId:    req.UserId,
		PostId:    req.PostId,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	
	// 保存到数据库
	err = s.bookmarkRepo.CreateBookmark(ctx, bookmark)
	if err != nil {
		return &proto.CreateBookmarkResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to create bookmark: " + err.Error(),
			},
		}, err
	}
	
	return &proto.CreateBookmarkResponse{
		Bookmark: bookmark,
	}, nil
}

// DeleteBookmark 删除书签
func (s *bookmarkService) DeleteBookmark(ctx context.Context, req *proto.DeleteBookmarkRequest) (*proto.DeleteBookmarkResponse, error) {
	err := s.bookmarkRepo.DeleteBookmark(ctx, req.UserId, req.PostId)
	if err != nil {
		return &proto.DeleteBookmarkResponse{
			Success: false,
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to delete bookmark: " + err.Error(),
			},
		}, err
	}
	
	return &proto.DeleteBookmarkResponse{
		Success: true,
	}, nil
}

// IsBookmarked 检查是否已收藏
func (s *bookmarkService) IsBookmarked(ctx context.Context, req *proto.IsBookmarkedRequest) (*proto.IsBookmarkedResponse, error) {
	isBookmarked, err := s.bookmarkRepo.IsBookmarked(ctx, req.UserId, req.PostId)
	if err != nil {
		return &proto.IsBookmarkedResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to check bookmark status: " + err.Error(),
			},
		}, err
	}
	
	return &proto.IsBookmarkedResponse{
		IsBookmarked: isBookmarked,
	}, nil
}

// ListBookmarks 获取用户书签列表
func (s *bookmarkService) ListBookmarks(ctx context.Context, req *proto.ListBookmarksRequest) (*proto.ListBookmarksResponse, error) {
	bookmarks, total, err := s.bookmarkRepo.ListBookmarks(ctx, req.UserId, req.Page, req.PageSize)
	if err != nil {
		return &proto.ListBookmarksResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to list bookmarks: " + err.Error(),
			},
		}, err
	}
	
	return &proto.ListBookmarksResponse{
		Bookmarks: bookmarks,
		Total:     total,
	}, nil
}