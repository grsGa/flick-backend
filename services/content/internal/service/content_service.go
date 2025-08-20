package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	
	"github.com/google/uuid"
	"github.com/flick/backend/services/content/internal/repository"
	"github.com/flick/backend/services/content/proto"
)

// postService 帖子服务实现
type postService struct {
	postRepo repository.PostRepository
}

// NewPostService 创建帖子服务实例
func NewPostService(postRepo repository.PostRepository) PostService {
	return &postService{
		postRepo: postRepo,
	}
}

// CreatePost 创建帖子
func (s *postService) CreatePost(ctx context.Context, req *proto.CreatePostRequest) (*proto.CreatePostResponse, error) {
	fmt.Printf("[Content Service] CreatePost called with:\n")
	fmt.Printf("  UserID: %s\n", req.UserId)
	fmt.Printf("  Content: %s\n", req.Content)
	fmt.Printf("  MediaUrls: %v\n", req.MediaUrls)
	fmt.Printf("  ParentId: %s\n", req.ParentId)
	
	// 验证请求参数
	if err := s.validateCreatePostRequest(req); err != nil {
		fmt.Printf("[Content Service] Validation failed: %v\n", err)
		return &proto.CreatePostResponse{
			Error: &proto.Error{
				Code:    400,
				Message: "Validation failed: " + err.Error(),
			},
		}, err
	}

	// 如果是回复，检查回复权限
	if req.ParentId != "" {
		canReply, _, err := s.postRepo.CheckReplyPermission(ctx, req.ParentId, req.UserId)
		if err != nil {
			return &proto.CreatePostResponse{
				Error: &proto.Error{
					Code:    500,
					Message: "Failed to check reply permission: " + err.Error(),
				},
			}, err
		}
		if !canReply {
			return &proto.CreatePostResponse{
				Error: &proto.Error{
					Code:    403,
					Message: "You don't have permission to reply to this post",
				},
			}, errors.New("reply permission denied")
		}
	}

	// 创建帖子
	fmt.Printf("[Content Service] Calling repository CreatePost\n")
	post, err := s.postRepo.CreatePost(ctx, req)
	if err != nil {
		fmt.Printf("[Content Service] Repository CreatePost failed: %v\n", err)
		return &proto.CreatePostResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to create post: " + err.Error(),
			},
		}, err
	}

	fmt.Printf("[Content Service] Post created successfully: %+v\n", post)

	// TODO: 发布帖子创建事件给interaction-service
	// s.publishPostCreatedEvent(post)

	return &proto.CreatePostResponse{
		Post: post,
	}, nil
}

// GetPost 获取帖子
func (s *postService) GetPost(ctx context.Context, req *proto.GetPostRequest) (*proto.GetPostResponse, error) {
	post, err := s.postRepo.GetPost(ctx, req.PostId, req.RequestingUserId)
	if err != nil {
		return &proto.GetPostResponse{
			Error: &proto.Error{
				Code:    404,
				Message: "Post not found: " + err.Error(),
			},
		}, err
	}

	return &proto.GetPostResponse{
		Post: post,
	}, nil
}

// GetUserPosts 获取用户帖子列表
func (s *postService) GetUserPosts(ctx context.Context, req *proto.GetUserPostsRequest) (*proto.GetUserPostsResponse, error) {
	posts, nextCursor, hasMore, err := s.postRepo.GetUserPosts(ctx, req.UserId, req.RequestingUserId, req.Limit, req.Cursor)
	if err != nil {
		return &proto.GetUserPostsResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to get user posts: " + err.Error(),
			},
		}, err
	}

	return &proto.GetUserPostsResponse{
		Posts: posts,
		NextCursor: nextCursor,
		HasMore: hasMore,
	}, nil
}

// GetTimeline 获取时间线
func (s *postService) GetTimeline(ctx context.Context, req *proto.GetTimelineRequest) (*proto.GetTimelineResponse, error) {
	posts, nextCursor, hasMore, err := s.postRepo.GetTimeline(ctx, req.UserId, req.Limit, req.Cursor)
	if err != nil {
		return &proto.GetTimelineResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to get timeline: " + err.Error(),
			},
		}, err
	}

	return &proto.GetTimelineResponse{
		Posts: posts,
		NextCursor: nextCursor,
		HasMore: hasMore,
	}, nil
}

// DeletePost 删除帖子
func (s *postService) DeletePost(ctx context.Context, req *proto.DeletePostRequest) (*proto.DeletePostResponse, error) {
	err := s.postRepo.DeletePost(ctx, req.PostId, req.UserId)
	if err != nil {
		return &proto.DeletePostResponse{
			Success: false,
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to delete post: " + err.Error(),
			},
		}, err
	}

	return &proto.DeletePostResponse{
		Success: true,
	}, nil
}

// CheckReplyPermission 检查回复权限
func (s *postService) CheckReplyPermission(ctx context.Context, req *proto.CheckReplyPermissionRequest) (*proto.CheckReplyPermissionResponse, error) {
	canReply, reason, err := s.postRepo.CheckReplyPermission(ctx, req.PostId, req.UserId)
	if err != nil {
		return &proto.CheckReplyPermissionResponse{
			CanReply: false,
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to check reply permission: " + err.Error(),
			},
		}, err
	}

	return &proto.CheckReplyPermissionResponse{
		CanReply: canReply,
		Reason: reason,
	}, nil
}

// validateCreatePostRequest 验证创建帖子请求
func (s *postService) validateCreatePostRequest(req *proto.CreatePostRequest) error {
	// 检查用户ID
	if req.UserId == "" {
		return errors.New("user ID is required")
	}

	// 检查内容长度 - 允许纯媒体帖子（无文本内容）
	if len(strings.TrimSpace(req.Content)) == 0 && len(req.MediaUrls) == 0 {
		return errors.New("post must have either content or media")
	}
	if len(req.Content) > 280 {
		return errors.New("post content cannot exceed 280 characters")
	}

	// 检查可见性
	if req.Visibility != "" && req.Visibility != "public" && req.Visibility != "private" && req.Visibility != "followers" {
		return errors.New("invalid visibility value")
	}

	// 检查回复权限
	if req.ReplyPermission != "" && req.ReplyPermission != "EVERYONE" && req.ReplyPermission != "FOLLOWING" && req.ReplyPermission != "MENTIONED_ONLY" {
		return errors.New("invalid reply permission value")
	}

	// 检查媒体数量限制
	if len(req.MediaUrls) > 4 {
		return errors.New("cannot attach more than 4 media files")
	}

	// 检查提及用户数量限制
	if len(req.MentionedUsers) > 10 {
		return errors.New("cannot mention more than 10 users")
	}

	// 检查标签数量限制
	if len(req.Tags) > 5 {
		return errors.New("cannot add more than 5 tags")
	}

	// 检查投票选项
	if req.PollData != nil {
		if len(req.PollData.Options) < 2 {
			return errors.New("poll must have at least 2 options")
		}
		if len(req.PollData.Options) > 4 {
			return errors.New("poll cannot have more than 4 options")
		}
		if req.PollData.DurationMinutes < 5 {
			return errors.New("poll duration must be at least 5 minutes")
		}
		if req.PollData.DurationMinutes > 10080 { // 7 days
			return errors.New("poll duration cannot exceed 7 days")
		}
		for _, option := range req.PollData.Options {
			if len(strings.TrimSpace(option)) == 0 {
				return errors.New("poll option cannot be empty")
			}
			if len(option) > 25 {
				return errors.New("poll option cannot exceed 25 characters")
			}
		}
	}

	return nil
}

// contentService 内容服务实现（保持兼容性）
type contentService struct {
	contentRepo repository.ContentRepository
}

// NewContentService 创建内容服务实例
func NewContentService(contentRepo repository.ContentRepository) ContentService {
	return &contentService{
		contentRepo: contentRepo,
	}
}

// GetContent 获取内容（保持兼容性）
func (s *contentService) GetContent(ctx context.Context, req *proto.GetPostRequest) (*proto.GetPostResponse, error) {
	content, err := s.contentRepo.GetContent(ctx, req.PostId)
	if err != nil {
		return &proto.GetPostResponse{
			Error: &proto.Error{
				Code:    404,
				Message: "Content not found: " + err.Error(),
			},
		}, err
	}
	
	return &proto.GetPostResponse{
		Post: content,
	}, nil
}

// CreateContent 创建内容（保持兼容性）
func (s *contentService) CreateContent(ctx context.Context, req *proto.CreatePostRequest) (*proto.CreatePostResponse, error) {
	// 创建帖子对象
	post := &proto.Post{
		Id:        uuid.New().String(),
		UserId:    req.UserId,
		Content:   req.Content,
		Visibility: req.Visibility,
		ReplyPermission: req.ReplyPermission,
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}
	
	// 保存到数据库
	err := s.contentRepo.CreateContent(ctx, post)
	if err != nil {
		return &proto.CreatePostResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to create content: " + err.Error(),
			},
		}, err
	}
	
	return &proto.CreatePostResponse{
		Post: post,
	}, nil
}

// UpdateContent 更新内容（保持兼容性）
func (s *contentService) UpdateContent(ctx context.Context, req *proto.CreatePostRequest) (*proto.CreatePostResponse, error) {
	// 首先获取现有内容
	content, err := s.contentRepo.GetContent(ctx, req.UserId)
	if err != nil {
		return &proto.CreatePostResponse{
			Error: &proto.Error{
				Code:    404,
				Message: "Content not found: " + err.Error(),
			},
		}, err
	}
	
	// 更新内容字段
	if req.Content != "" {
		content.Content = req.Content
	}
	
	content.UpdatedAt = time.Now().Format(time.RFC3339)
	
	// 保存到数据库
	err = s.contentRepo.UpdateContent(ctx, content)
	if err != nil {
		return &proto.CreatePostResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to update content: " + err.Error(),
			},
		}, err
	}
	
	return &proto.CreatePostResponse{
		Post: content,
	}, nil
}

// DeleteContent 删除内容（保持兼容性）
func (s *contentService) DeleteContent(ctx context.Context, req *proto.DeletePostRequest) (*proto.DeletePostResponse, error) {
	err := s.contentRepo.DeleteContent(ctx, req.PostId)
	if err != nil {
		return &proto.DeletePostResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to delete content: " + err.Error(),
			},
		}, err
	}
	
	return &proto.DeletePostResponse{
		Success: true,
	}, nil
}

// ListContent 列出内容（保持兼容性）
func (s *contentService) ListContent(ctx context.Context, req *proto.GetUserPostsRequest) (*proto.GetUserPostsResponse, error) {
	contents, total, err := s.contentRepo.ListContent(ctx, req.UserId, req.Limit, 20)
	if err != nil {
		return &proto.GetUserPostsResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to list content: " + err.Error(),
			},
		}, err
	}
	
	return &proto.GetUserPostsResponse{
		Posts:   contents,
		HasMore: total > int32(len(contents)),
	}, nil
}
