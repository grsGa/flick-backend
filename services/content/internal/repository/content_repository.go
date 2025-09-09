package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"

	"github.com/flick/backend/pkg/database"
	"github.com/flick/backend/pkg/models"
	content_proto "github.com/flick/backend/services/content/proto"
	media_proto "github.com/flick/backend/services/media/proto"
	user_proto "github.com/flick/backend/services/user/proto"
)

// EventPublisher 事件发布接口
type EventPublisher interface {
	PublishReplyDeleted(ctx context.Context, eventData map[string]interface{}) error
	PublishReplyCreated(ctx context.Context, eventData map[string]interface{}) error
}

// postRepository 帖子仓储实现
type postRepository struct {
	db             *gorm.DB
	userClient     user_proto.UserServiceClient
	mediaClient    media_proto.MediaServiceClient
	eventPublisher EventPublisher
}

// NewPostRepository 创建帖子仓储实例
func NewPostRepository(eventPublisher EventPublisher) PostRepository {
	fmt.Printf("[Content Repository] Initializing PostRepository with user service integration\n")

	// 连接用户服务
	fmt.Printf("[Content Repository] Attempting to connect to user-service:50051\n")
	userConn, err := grpc.Dial("user-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("[Content Repository] Failed to connect to user service: %v\n", err)
	}

	// 连接媒体服务
	fmt.Printf("[Content Repository] Attempting to connect to media-service:50054\n")
	mediaConn, err2 := grpc.Dial("media-service:50054", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err2 != nil {
		fmt.Printf("[Content Repository] Failed to connect to media service: %v\n", err2)
	}

	var userClient user_proto.UserServiceClient
	var mediaClient media_proto.MediaServiceClient

	if err == nil {
		userClient = user_proto.NewUserServiceClient(userConn)
		fmt.Printf("[Content Repository] Successfully connected to user service\n")
	}

	if err2 == nil {
		mediaClient = media_proto.NewMediaServiceClient(mediaConn)
		fmt.Printf("[Content Repository] Successfully connected to media service\n")
	}

	return &postRepository{
		db:             database.GetDB(),
		userClient:     userClient,
		mediaClient:    mediaClient,
		eventPublisher: eventPublisher,
	}
}

// CreatePost 创建帖子
func (r *postRepository) CreatePost(ctx context.Context, req *content_proto.CreatePostRequest) (*content_proto.Post, error) {
	fmt.Printf("[Content Repository] CreatePost called with:\n")
	fmt.Printf("  UserID: %s\n", req.UserId)
	fmt.Printf("  Content: %s\n", req.Content)
	fmt.Printf("  MediaUrls: %v\n", req.MediaUrls)

	// SECURITY: Validate user exists before creating post
	fmt.Printf("[Content Repository] Validating user exists: %s\n", req.UserId)
	author := r.fetchUserInfo(ctx, req.UserId)
	if author == nil {
		fmt.Printf("[Content Repository] User %s does not exist, rejecting post creation\n", req.UserId)
		return nil, fmt.Errorf("user does not exist: %s", req.UserId)
	}
	fmt.Printf("[Content Repository] User validation passed: %s (%s)\n", author.Username, author.DisplayName)

	// 开始事务
	fmt.Printf("[Content Repository] Starting database transaction\n")
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[Content Repository] Transaction panic, rolling back\n")
			tx.Rollback()
		}
	}()

	// 创建帖子
	visibility := req.Visibility
	if visibility == "" {
		visibility = "public" // 默认为公开
	} else {
		// 将大写值转换为小写以匹配数据库约束
		switch strings.ToUpper(visibility) {
		case "PUBLIC":
			visibility = "public"
		case "PRIVATE":
			visibility = "private"
		case "FOLLOWERS":
			visibility = "followers"
		}
	}

	replyPermission := req.ReplyPermission
	if replyPermission == "" {
		replyPermission = "EVERYONE" // 默认允许所有人回复
	}

	post := &models.Post{
		UserID:          req.UserId,
		Content:         req.Content,
		Visibility:      visibility,
		ReplyPermission: replyPermission,
		HasMedia:        len(req.MediaUrls) > 0,
		HasPoll:         req.PollData != nil,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	fmt.Printf("[Content Repository] Created post model with UserID: %s, Content: %s\n", post.UserID, post.Content)

	// 处理回复逻辑
	if req.ParentId != "" {
		post.ParentID = &req.ParentId
		post.IsReply = true

		// 获取父帖子信息来确定回复层级和根帖子ID
		var parentPost models.Post
		if err := tx.Where("id = ?", req.ParentId).First(&parentPost).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to find parent post: %w", err)
		}

		// 设置回复层级和根帖子ID - 强制执行两级回复系统
		if parentPost.IsReply {
			// 对回复的回复：需要区分是对一级回复还是二级回复的回复
			if parentPost.ReplyLevel == 1 {
				// 对一级回复的回复：设为level 2
				post.ReplyLevel = 2
				post.RootID = parentPost.RootID
				// ParentID指向被回复的一级回复
				// post.ParentID 已经设置为 req.ParentId

				// 保存原始被回复的用户信息到ReplyMention表
				originalReplyToUserID := parentPost.UserID
				fmt.Printf("[Content Repository] Creating level 2 reply to level 1 - RootID: %v, ParentID: %v, OriginalReplyTo: %s\n", post.RootID, post.ParentID, originalReplyToUserID)
			} else {
				// 对二级回复的回复：仍然设为level 2，但ParentID指向被回复的二级回复的ParentID（即一级回复）
				post.ReplyLevel = 2
				post.RootID = parentPost.RootID
				// 关键：ParentID应该指向被回复的二级回复的ParentID，这样可以正确分组显示
				post.ParentID = parentPost.ParentID

				// 保存被回复的二级回复用户信息到ReplyMention表
				originalReplyToUserID := parentPost.UserID
				fmt.Printf("[Content Repository] Creating level 2 reply to level 2 - RootID: %v, ParentID: %v (grouped under same level 1), OriginalReplyTo: %s\n", post.RootID, post.ParentID, originalReplyToUserID)
			}

			post.Content = req.Content
		} else {
			// 对原帖的回复：level 1
			post.ReplyLevel = 1
			post.RootID = &req.ParentId
			fmt.Printf("[Content Repository] Creating level 1 reply - RootID: %v, ParentID: %v\n", post.RootID, post.ParentID)
		}

		// 验证回复层级（应该永远不会超过2）
		if post.ReplyLevel > 2 {
			tx.Rollback()
			return nil, fmt.Errorf("reply level cannot exceed 2 - current level: %d", post.ReplyLevel)
		}

		fmt.Printf("[Content Repository] Final reply structure - Level: %d, ParentID: %v, RootID: %v\n", post.ReplyLevel, post.ParentID, post.RootID)
	}

	// 设置转发帖子ID
	if req.RepostId != "" {
		post.RepostID = &req.RepostId
	}

	fmt.Printf("[Content Repository] Inserting post into database\n")
	if err := tx.Create(post).Error; err != nil {
		fmt.Printf("[Content Repository] Database insert failed: %v\n", err)
		tx.Rollback()
		return nil, fmt.Errorf("failed to create post: %w", err)
	}
	fmt.Printf("[Content Repository] Post inserted successfully with ID: %s\n", post.ID)

	// 如果是二级回复，创建ReplyMention记录来保存原始被回复的用户信息
	if post.ReplyLevel == 2 && req.ParentId != "" {
		// 获取被回复的帖子信息（可能是一级回复或二级回复）
		var repliedPost models.Post
		if err := tx.Where("id = ?", req.ParentId).First(&repliedPost).Error; err == nil {
			var mentionedUserID string

			if repliedPost.ReplyLevel == 1 {
				// 回复一级回复：保存一级回复的用户ID
				mentionedUserID = repliedPost.UserID
				fmt.Printf("[Content Repository] Reply to level 1 - mentioning user: %s\n", mentionedUserID)
			} else if repliedPost.ReplyLevel == 2 {
				// 回复二级回复：保存二级回复的用户ID
				mentionedUserID = repliedPost.UserID
				fmt.Printf("[Content Repository] Reply to level 2 - mentioning user: %s\n", mentionedUserID)
			}

			if mentionedUserID != "" {
				replyMention := models.ReplyMention{
					ReplyID:         post.ID,
					MentionedUserID: mentionedUserID,
					CreatedAt:       time.Now(),
				}
				if err := tx.Create(&replyMention).Error; err != nil {
					fmt.Printf("[Content Repository] Warning: Failed to create reply mention: %v\n", err)
					// 不回滚，因为这不是关键错误
				} else {
					fmt.Printf("[Content Repository] Created reply mention for reply %s -> user %s\n", post.ID, mentionedUserID)
				}
			}
		}
	}

	// 创建媒体附件
	if len(req.MediaUrls) > 0 {
		for _, mediaURL := range req.MediaUrls {
			// 从URL提取文件ID (URL格式: http://host/bucket/posts/userID/fileID/filename)
			parts := strings.Split(mediaURL, "/")
			if len(parts) < 6 {
				fmt.Printf("[Content Repository] CreatePost - Invalid media URL format: %s\n", mediaURL)
				continue
			}

			// 提取文件ID (倒数第二个路径段)
			fileID := parts[len(parts)-2]
			fmt.Printf("[Content Repository] CreatePost - Extracted file ID from URL: %s -> %s\n", mediaURL, fileID)

			// 查询Media服务获取完整的媒体信息，带重试逻辑等待variants处理完成
			mediaReq := &media_proto.GetFileRequest{
				FileId: fileID,
			}

			var mediaResp *media_proto.GetFileResponse
			var err error

			// 重试逻辑：针对视频处理的更长等待时间
			maxRetries := 10
			retryDelay := 5 * time.Second

			for attempt := 0; attempt <= maxRetries; attempt++ {
				if attempt > 0 {
					fmt.Printf("[Content Repository] CreatePost - Waiting %v for media processing (attempt %d/%d)\n", retryDelay, attempt, maxRetries)
					time.Sleep(retryDelay)
				}

				mediaResp, err = r.mediaClient.GetFile(ctx, mediaReq)
				if err == nil && mediaResp.File != nil && mediaResp.File.Variants != nil {
					fmt.Printf("[Content Repository] CreatePost - Successfully got variants on attempt %d\n", attempt+1)
					break
				}

				if attempt < maxRetries {
					fmt.Printf("[Content Repository] CreatePost - No variants yet on attempt %d, retrying...\n", attempt+1)
				}
			}
			if err != nil {
				fmt.Printf("[Content Repository] CreatePost - Failed to get media from media service: %v\n", err)
				// 如果无法从Media服务获取，则创建基本记录
				mediaType := "image"
				mimeType := "image/jpeg"

				if strings.Contains(mediaURL, ".mp4") {
					mediaType = "video"
					mimeType = "video/mp4"
				} else if strings.Contains(mediaURL, ".webm") {
					mediaType = "video"
					mimeType = "video/webm"
				} else if strings.Contains(mediaURL, ".mov") {
					mediaType = "video"
					mimeType = "video/quicktime"
				} else if strings.Contains(mediaURL, ".png") {
					mimeType = "image/png"
				} else if strings.Contains(mediaURL, ".gif") {
					mimeType = "image/gif"
				} else if strings.Contains(mediaURL, ".webp") {
					mimeType = "image/webp"
				}

				media := &models.MediaAttachment{
					PostID:    &post.ID,
					UserID:    req.UserId,
					URL:       mediaURL,
					Type:      mediaType,
					MimeType:  mimeType,
					Status:    "active",
					CreatedAt: time.Now(),
				}

				if err := tx.Create(media).Error; err != nil {
					fmt.Printf("[Content Repository] CreatePost - Failed to create media attachment: %v\n", err)
					tx.Rollback()
					return nil, fmt.Errorf("failed to create media attachment: %w", err)
				}
				fmt.Printf("[Content Repository] CreatePost - Created fallback media attachment with ID: %s\n", media.ID)
				continue
			}

			if mediaResp.Error != nil {
				fmt.Printf("[Content Repository] CreatePost - Media service returned error: %s\n", mediaResp.Error.Message)
				continue
			}

			// 使用Media服务返回的完整信息创建媒体附件记录
			mediaFile := mediaResp.File

			// 转换variants到数据库格式
			var variants models.MediaVariants
			if mediaFile.Variants != nil {
				if mediaFile.Variants.Thumbnail != nil {
					variants.Thumbnail = &models.MediaVariant{
						URL:    mediaFile.Variants.Thumbnail.Url,
						Width:  mediaFile.Variants.Thumbnail.Width,
						Height: mediaFile.Variants.Thumbnail.Height,
						Size:   mediaFile.Variants.Thumbnail.Size,
					}
				}
				if mediaFile.Variants.Small != nil {
					variants.Small = &models.MediaVariant{
						URL:    mediaFile.Variants.Small.Url,
						Width:  mediaFile.Variants.Small.Width,
						Height: mediaFile.Variants.Small.Height,
						Size:   mediaFile.Variants.Small.Size,
					}
				}
				if mediaFile.Variants.Medium != nil {
					variants.Medium = &models.MediaVariant{
						URL:    mediaFile.Variants.Medium.Url,
						Width:  mediaFile.Variants.Medium.Width,
						Height: mediaFile.Variants.Medium.Height,
						Size:   mediaFile.Variants.Medium.Size,
					}
				}
				if mediaFile.Variants.Large != nil {
					variants.Large = &models.MediaVariant{
						URL:    mediaFile.Variants.Large.Url,
						Width:  mediaFile.Variants.Large.Width,
						Height: mediaFile.Variants.Large.Height,
						Size:   mediaFile.Variants.Large.Size,
					}
				}
				if mediaFile.Variants.Original != nil {
					variants.Original = &models.MediaVariant{
						URL:    mediaFile.Variants.Original.Url,
						Width:  mediaFile.Variants.Original.Width,
						Height: mediaFile.Variants.Original.Height,
						Size:   mediaFile.Variants.Original.Size,
					}
				}
				// 处理视频特有的variants
				if mediaFile.Variants.Preview != nil {
					variants.Preview = &models.MediaVariant{
						URL:    mediaFile.Variants.Preview.Url,
						Width:  mediaFile.Variants.Preview.Width,
						Height: mediaFile.Variants.Preview.Height,
						Size:   mediaFile.Variants.Preview.Size,
					}
				}
				if mediaFile.Variants.LowRes != nil {
					variants.LowRes = &models.MediaVariant{
						URL:    mediaFile.Variants.LowRes.Url,
						Width:  mediaFile.Variants.LowRes.Width,
						Height: mediaFile.Variants.LowRes.Height,
						Size:   mediaFile.Variants.LowRes.Size,
					}
				}
				if mediaFile.Variants.MidRes != nil {
					variants.MidRes = &models.MediaVariant{
						URL:    mediaFile.Variants.MidRes.Url,
						Width:  mediaFile.Variants.MidRes.Width,
						Height: mediaFile.Variants.MidRes.Height,
						Size:   mediaFile.Variants.MidRes.Size,
					}
				}
				if mediaFile.Variants.HighRes != nil {
					variants.HighRes = &models.MediaVariant{
						URL:    mediaFile.Variants.HighRes.Url,
						Width:  mediaFile.Variants.HighRes.Width,
						Height: mediaFile.Variants.HighRes.Height,
						Size:   mediaFile.Variants.HighRes.Size,
					}
				}
			}

			var processedAt *time.Time
			if mediaFile.ProcessedAt != "" {
				if t, err := time.Parse(time.RFC3339, mediaFile.ProcessedAt); err == nil {
					processedAt = &t
				}
			}

			// 更新现有的媒体记录而不是创建新的
			updateData := map[string]interface{}{
				"post_id":      &post.ID,
				"filename":     mediaFile.Filename,
				"url":          mediaFile.Url,
				"type":         mediaFile.Type,
				"mime_type":    mediaFile.MimeType,
				"size":         mediaFile.Size,
				"status":       mediaFile.Status,
				"width":        mediaFile.Width,
				"height":       mediaFile.Height,
				"duration":     mediaFile.Duration,
				"variants":     variants,
				"alt_text":     &mediaFile.AltText,
				"processed_at": processedAt,
				"updated_at":   time.Now(),
			}

			fmt.Printf("[Content Repository] CreatePost - Updating existing media attachment: ID=%s, URL=%s, HasVariants=%t\n",
				mediaFile.Id, mediaURL, mediaFile.Variants != nil)

			if err := tx.Model(&models.MediaAttachment{}).Where("id = ?", mediaFile.Id).Updates(updateData).Error; err != nil {
				fmt.Printf("[Content Repository] CreatePost - Failed to update media attachment: %v\n", err)
				tx.Rollback()
				return nil, fmt.Errorf("failed to update media attachment: %w", err)
			}
			fmt.Printf("[Content Repository] CreatePost - Successfully updated media attachment with ID: %s\n", mediaFile.Id)
		}
	}

	// 创建提及记录
	if len(req.MentionedUsers) > 0 {
		for _, userID := range req.MentionedUsers {
			mention := &models.PostMention{
				PostID:          post.ID,
				MentionedUserID: userID,
				CreatedAt:       time.Now(),
			}
			if err := tx.Create(mention).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create mention: %w", err)
			}
		}
	}

	// 创建标签记录
	if len(req.Tags) > 0 {
		for _, tag := range req.Tags {
			postTag := &models.PostTag{
				PostID:    post.ID,
				Tag:       tag,
				CreatedAt: time.Now(),
			}
			if err := tx.Create(postTag).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create tag: %w", err)
			}
		}
	}

	// 创建投票
	var poll *models.Poll
	if req.PollData != nil {
		expiresAt := time.Now().Add(time.Duration(req.PollData.DurationMinutes) * time.Minute)
		poll = &models.Poll{
			PostID:          post.ID,
			Question:        req.PollData.Question,
			DurationMinutes: int(req.PollData.DurationMinutes),
			ExpiresAt:       expiresAt,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		if err := tx.Create(poll).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create poll: %w", err)
		}

		// 创建投票选项
		for i, optionText := range req.PollData.Options {
			option := &models.PollOption{
				PollID:    poll.ID,
				Text:      optionText,
				Position:  i + 1,
				VoteCount: 0,
				CreatedAt: time.Now(),
			}
			if err := tx.Create(option).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create poll option: %w", err)
			}
		}
	}

	// 提交事务
	fmt.Printf("[Content Repository] Committing transaction\n")
	if err := tx.Commit().Error; err != nil {
		fmt.Printf("[Content Repository] Transaction commit failed: %v\n", err)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	fmt.Printf("[Content Repository] Transaction committed successfully\n")

	// 如果创建的是回复，发布回复创建事件以更新父帖子的回复计数
	if post.IsReply && r.eventPublisher != nil {
		eventData := map[string]interface{}{
			"reply_id":   post.ID,
			"parent_id":  post.ParentID,
			"root_id":    post.RootID,
			"user_id":    post.UserID,
			"level":      post.ReplyLevel,
			"created_at": post.CreatedAt.Format(time.RFC3339),
		}
		
		if err := r.eventPublisher.PublishReplyCreated(ctx, eventData); err != nil {
			// 记录错误但不影响创建操作
			fmt.Printf("[Content Repository] Failed to publish reply created event: %v\n", err)
		}
	}

	// 构建返回的帖子对象
	fmt.Printf("[Content Repository] Building response with committed data\n")
	return r.buildPostProto(ctx, post, req.MediaUrls, req.MentionedUsers, req.Tags, poll, req.PollData)
}

// GetPost 根据ID获取帖子
func (r *postRepository) GetPost(ctx context.Context, postID, requestingUserID string) (*content_proto.Post, error) {
	var post models.Post
	if err := r.db.Where("id = ? AND deleted_at IS NULL", postID).First(&post).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("post not found")
		}
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	// 获取媒体附件
	var mediaAttachments []models.MediaAttachment
	fmt.Printf("[Content Repository] GetPost - Querying media attachments for post ID: %s\n", post.ID)
	if err := r.db.Where("post_id = ? AND deleted_at IS NULL", post.ID).Find(&mediaAttachments).Error; err != nil {
		fmt.Printf("[Content Repository] GetPost - Failed to fetch media attachments: %v\n", err)
	}
	fmt.Printf("[Content Repository] GetPost - Found %d media attachments for post %s\n", len(mediaAttachments), post.ID)
	for i, media := range mediaAttachments {
		fmt.Printf("[Content Repository] GetPost - Media %d: ID=%s, PostID=%v, URL=%s\n", i, media.ID, media.PostID, media.URL)
	}

	// 查询提及信息 (如果表存在)
	var mentions []models.PostMention
	if r.db.Migrator().HasTable(&models.PostMention{}) {
		if err := r.db.Where("post_id = ?", post.ID).Find(&mentions).Error; err != nil {
			fmt.Printf("[Content Repository] Failed to fetch mentions: %v\n", err)
		}
	} else {
		fmt.Printf("[Content Repository] PostMention table does not exist, skipping mentions query\n")
	}

	// 查询标签信息 (如果表存在)
	var tags []models.PostTag
	if r.db.Migrator().HasTable(&models.PostTag{}) {
		if err := r.db.Where("post_id = ?", post.ID).Find(&tags).Error; err != nil {
			fmt.Printf("[Content Repository] Failed to fetch tags: %v\n", err)
		}
	} else {
		fmt.Printf("[Content Repository] PostTag table does not exist, skipping tags query\n")
	}

	// 获取提及用户
	mentionedUsers := make([]string, len(mentions))
	for i, mention := range mentions {
		mentionedUsers[i] = mention.MentionedUserID
	}

	// 获取标签
	tagStrings := make([]string, len(tags))
	for i, tag := range tags {
		tagStrings[i] = tag.Tag
	}

	// 获取投票信息
	var poll *models.Poll
	if post.HasPoll {
		if err := r.db.Where("post_id = ?", post.ID).First(&poll).Error; err == nil {
			var options []models.PollOption
			r.db.Where("poll_id = ?", poll.ID).Order("position").Find(&options)

			pollOptions := make([]*content_proto.PollOption, len(options))
			for i, option := range options {
				pollOptions[i] = &content_proto.PollOption{
					Id:        option.ID,
					Text:      option.Text,
					VoteCount: int32(option.VoteCount),
				}
			}
		}
	}

	return r.buildPostProto(ctx, &post, r.getMediaURLs(mediaAttachments), mentionedUsers, tagStrings, poll, nil)
}

// GetUserPosts 获取用户帖子列表
func (r *postRepository) GetUserPosts(ctx context.Context, userID, requestingUserID string, limit int32, cursor string) ([]*content_proto.Post, string, bool, error) {
	fmt.Printf("[Content Repository] GetUserPosts called for userID: %s, requestingUserID: %s, limit: %d\n", userID, requestingUserID, limit)

	var posts []models.Post
	query := r.db.Where("user_id = ? AND deleted_at IS NULL AND is_reply = false", userID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(int(limit))
	}
	if cursor != "" {
		// cursor是post ID，需要找到该post的created_at时间
		var cursorPost models.Post
		if err := r.db.Where("id = ?", cursor).First(&cursorPost).Error; err == nil {
			query = query.Where("created_at < ?", cursorPost.CreatedAt)
			fmt.Printf("[Content Repository] Using cursor post %s with created_at: %s\n", cursor, cursorPost.CreatedAt.Format(time.RFC3339))
		} else {
			fmt.Printf("[Content Repository] Failed to find cursor post %s: %v, ignoring cursor\n", cursor, err)
			// 如果找不到cursor post，忽略cursor继续查询，而不是失败
		}
	}

	// 添加SQL调试
	fmt.Printf("[Content Repository] Executing query to find posts for user: %s\n", userID)

	if err := query.Find(&posts).Error; err != nil {
		fmt.Printf("[Content Repository] Query failed: %v\n", err)
		return nil, "", false, fmt.Errorf("failed to get user posts: %w", err)
	}

	fmt.Printf("[Content Repository] Found %d posts for user: %s\n", len(posts), userID)
	for i, post := range posts {
		fmt.Printf("[Content Repository] Post %d: ID=%s, Content=%s, CreatedAt=%s\n",
			i, post.ID, post.Content[:min(50, len(post.Content))], post.CreatedAt.Format(time.RFC3339))
	}

	postList, err := r.buildPostList(posts)
	if err != nil {
		fmt.Printf("[Content Repository] Failed to build post list: %v\n", err)
		return nil, "", false, err
	}

	fmt.Printf("[Content Repository] Built %d posts successfully\n", len(postList))

	// 生成下一页cursor（简化处理）
	nextCursor := ""
	// 只有当返回的帖子数等于请求的limit时，才可能有更多帖子
	// 如果返回的帖子数少于limit，说明已经到底了
	hasMore := limit > 0 && len(posts) == int(limit)
	if hasMore && len(posts) > 0 {
		nextCursor = posts[len(posts)-1].CreatedAt.Format(time.RFC3339)
	}

	fmt.Printf("[Content Repository] GetUserPosts result: posts=%d, hasMore=%t, nextCursor=%s\n", len(posts), hasMore, nextCursor)
	return postList, nextCursor, hasMore, nil
}

// GetTimeline 获取时间线帖子 (For you feed - 显示所有公开帖子)
func (r *postRepository) GetTimeline(ctx context.Context, userID string, limit int32, cursor string) ([]*content_proto.Post, string, bool, error) {
	fmt.Printf("[Content Repository] GetTimeline (For you) called for userID: %s, limit: %d\n", userID, limit)

	// For you feed显示所有公开帖子，包括用户自己的帖子
	var posts []models.Post
	query := r.db.Where("visibility = 'public' AND deleted_at IS NULL AND is_reply = false").
		Order("created_at DESC")

	// 如果limit为-1，表示获取所有帖子，不应用限制
	// 如果limit为0或正数，应用相应的限制
	if limit > 0 {
		query = query.Limit(int(limit))
	}
	if cursor != "" {
		// cursor是post ID，需要找到该post的created_at时间
		var cursorPost models.Post
		if err := r.db.Where("id = ?", cursor).First(&cursorPost).Error; err == nil {
			query = query.Where("created_at < ?", cursorPost.CreatedAt)
			fmt.Printf("[Content Repository] Using cursor post %s with created_at: %s\n", cursor, cursorPost.CreatedAt.Format(time.RFC3339))
		} else {
			fmt.Printf("[Content Repository] Failed to find cursor post %s: %v\n", cursor, err)
		}
	}

	fmt.Printf("[Content Repository] Executing timeline query\n")
	if err := query.Find(&posts).Error; err != nil {
		fmt.Printf("[Content Repository] Timeline query failed: %v\n", err)
		return nil, "", false, fmt.Errorf("failed to get timeline: %w", err)
	}

	fmt.Printf("[Content Repository] Timeline found %d posts\n", len(posts))
	for i, post := range posts {
		fmt.Printf("[Content Repository] Timeline Post %d: ID=%s, UserID=%s, Content=%s, CreatedAt=%s\n",
			i, post.ID, post.UserID, post.Content[:min(50, len(post.Content))], post.CreatedAt.Format(time.RFC3339))
	}

	postList, err := r.buildPostList(posts)
	if err != nil {
		fmt.Printf("[Content Repository] Failed to build timeline post list: %v\n", err)
		return nil, "", false, err
	}

	fmt.Printf("[Content Repository] Built %d timeline posts successfully\n", len(postList))

	// 生成下一页cursor（简化处理）
	nextCursor := ""
	// 如果limit为-1（获取所有帖子），则没有更多页面
	// 否则，如果返回的帖子数等于limit，说明可能还有更多
	hasMore := limit > 0 && len(posts) == int(limit)
	if hasMore && len(posts) > 0 {
		nextCursor = posts[len(posts)-1].CreatedAt.Format(time.RFC3339)
	}

	fmt.Printf("[Content Repository] GetTimeline result: posts=%d, hasMore=%t, nextCursor=%s\n", len(posts), hasMore, nextCursor)
	return postList, nextCursor, hasMore, nil
}

// GetFollowingTimeline 获取关注用户时间线帖子 (Following feed)
func (r *postRepository) GetFollowingTimeline(ctx context.Context, userID string, limit int32, cursor string) ([]*content_proto.Post, string, bool, error) {
	fmt.Printf("[Content Repository] GetFollowingTimeline called for userID: %s, limit: %d\n", userID, limit)

	// 获取用户关注的用户列表
	followingUserIDs, err := r.getFollowingUserIDs(ctx, userID)
	if err != nil {
		fmt.Printf("[Content Repository] Failed to get following users: %v\n", err)
		// 如果获取关注列表失败，返回空结果
		return []*content_proto.Post{}, "", false, nil
	}

	fmt.Printf("[Content Repository] User %s is following %d users: %v\n", userID, len(followingUserIDs), followingUserIDs)

	// 如果用户没有关注任何人，返回空结果
	if len(followingUserIDs) == 0 {
		fmt.Printf("[Content Repository] User %s is not following anyone, returning empty timeline\n", userID)
		return []*content_proto.Post{}, "", false, nil
	}

	// 获取关注用户的帖子，排除回复内容，只显示原创帖子
	var posts []models.Post
	query := r.db.Where("visibility = 'public' AND deleted_at IS NULL AND is_reply = false AND user_id IN ?", followingUserIDs).
		Order("created_at DESC")

	// 如果limit为-1，表示获取所有帖子，不应用限制
	// 如果limit为0或正数，应用相应的限制
	if limit > 0 {
		query = query.Limit(int(limit))
	}
	if cursor != "" {
		// cursor是post ID，需要找到该post的created_at时间
		var cursorPost models.Post
		if err := r.db.Where("id = ?", cursor).First(&cursorPost).Error; err == nil {
			query = query.Where("created_at < ?", cursorPost.CreatedAt)
			fmt.Printf("[Content Repository] Using cursor post %s with created_at: %s\n", cursor, cursorPost.CreatedAt.Format(time.RFC3339))
		} else {
			fmt.Printf("[Content Repository] Failed to find cursor post %s: %v\n", cursor, err)
		}
	}

	fmt.Printf("[Content Repository] Executing following timeline query\n")
	if err := query.Find(&posts).Error; err != nil {
		fmt.Printf("[Content Repository] Following timeline query failed: %v\n", err)
		return nil, "", false, fmt.Errorf("failed to get following timeline: %w", err)
	}

	fmt.Printf("[Content Repository] Following timeline found %d posts\n", len(posts))
	for i, post := range posts {
		fmt.Printf("[Content Repository] Following Timeline Post %d: ID=%s, UserID=%s, Content=%s, CreatedAt=%s\n",
			i, post.ID, post.UserID, post.Content[:min(50, len(post.Content))], post.CreatedAt.Format(time.RFC3339))
	}

	postList, err := r.buildPostList(posts)
	if err != nil {
		fmt.Printf("[Content Repository] Failed to build following timeline post list: %v\n", err)
		return nil, "", false, err
	}

	fmt.Printf("[Content Repository] Built %d following timeline posts successfully\n", len(postList))

	// 生成下一页cursor（简化处理）
	nextCursor := ""
	hasMore := limit > 0 && len(posts) == int(limit)
	if hasMore && len(posts) > 0 {
		nextCursor = posts[len(posts)-1].CreatedAt.Format(time.RFC3339)
	}

	fmt.Printf("[Content Repository] GetFollowingTimeline result: posts=%d, hasMore=%t, nextCursor=%s\n", len(posts), hasMore, nextCursor)
	return postList, nextCursor, hasMore, nil
}

// getFollowingUserIDs 获取用户关注的用户ID列表
func (r *postRepository) getFollowingUserIDs(ctx context.Context, userID string) ([]string, error) {
	// 检查用户服务客户端是否可用
	if r.userClient == nil {
		return nil, fmt.Errorf("user service client not available")
	}

	req := &user_proto.GetFollowingRequest{
		UserId: userID,
		First:  1000, // 获取最多1000个关注用户
		After:  "",
	}

	resp, err := r.userClient.GetFollowing(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get following users: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("user service error: %s", resp.Error.Message)
	}

	var followingIDs []string
	for _, user := range resp.Users {
		followingIDs = append(followingIDs, user.Id)
	}

	return followingIDs, nil
}

// DeletePost 删除帖子（软删除）
func (r *postRepository) DeletePost(ctx context.Context, id, userID string) error {
	// 首先获取要删除的帖子信息，用于后续的计数更新
	var post models.Post
	if err := r.db.Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).First(&post).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("post not found or not authorized")
		}
		return fmt.Errorf("failed to get post: %w", err)
	}

	// 执行软删除
	result := r.db.Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).
		Update("deleted_at", time.Now())

	if result.Error != nil {
		return fmt.Errorf("failed to delete post: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("post not found or not authorized")
	}

	// 如果删除的是回复，需要发布回复删除事件以更新父帖子的回复计数
	if post.IsReply && r.eventPublisher != nil {
		eventData := map[string]interface{}{
			"reply_id":   post.ID,
			"parent_id":  post.ParentID,
			"root_id":    post.RootID,
			"user_id":    post.UserID,
			"level":      post.ReplyLevel,
			"deleted_at": time.Now().Format(time.RFC3339),
		}
		
		if err := r.eventPublisher.PublishReplyDeleted(ctx, eventData); err != nil {
			// 记录错误但不影响删除操作
			fmt.Printf("[Content Repository] Failed to publish reply deleted event: %v\n", err)
		}
	}

	return nil
}

// CheckReplyPermission 检查回复权限
func (r *postRepository) CheckReplyPermission(ctx context.Context, postID, userID string) (bool, string, error) {
	var post models.Post
	if err := r.db.Where("id = ? AND deleted_at IS NULL", postID).First(&post).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, "post not found", nil
		}
		return false, "failed to get post", err
	}

	// 帖子作者总是可以回复
	if post.UserID == userID {
		return true, "", nil
	}

	switch post.ReplyPermission {
	case "EVERYONE":
		return true, "", nil
	case "FOLLOWING":
		// 需要检查关注关系，这里简化处理
		// 实际应该调用用户服务检查关注关系
		return true, "", nil // 临时返回true
	case "MENTIONED_ONLY":
		// 检查用户是否被提及
		var count int64
		r.db.Model(&models.PostMention{}).
			Where("post_id = ? AND mentioned_user_id = ?", postID, userID).
			Count(&count)
		if count > 0 {
			return true, "", nil
		}
		return false, "user not mentioned", nil
	default:
		return false, "invalid reply permission", nil
	}
}

// buildPostProto 构建帖子Proto对象
func (r *postRepository) buildPostProto(ctx context.Context, post *models.Post, mediaURLs []string, mentionedUsers []string, tags []string, poll *models.Poll, pollData *content_proto.PollData) (*content_proto.Post, error) {
	// 获取实际的媒体附件从数据库
	var dbMediaAttachments []models.MediaAttachment
	r.db.Where("post_id = ? AND deleted_at IS NULL", post.ID).Find(&dbMediaAttachments)

	// 构建媒体附件 - 使用数据库中的实际数据并获取variants信息
	mediaAttachments := make([]*content_proto.MediaAttachment, len(dbMediaAttachments))
	for i, media := range dbMediaAttachments {
		// 从Media服务获取完整的媒体信息包括variants
		variants := r.getMediaVariants(ctx, media.ID)

		mediaAttachments[i] = &content_proto.MediaAttachment{
			Id:       media.ID,
			Url:      media.URL,
			Type:     media.Type,
			MimeType: media.MimeType,
			Width:    int32(media.Width),
			Height:   int32(media.Height),
			Variants: variants,
		}
	}

	// 构建投票信息
	var pollProto *content_proto.Poll
	if poll != nil {
		// 获取投票选项
		var options []models.PollOption
		r.db.Where("poll_id = ?", poll.ID).Order("position").Find(&options)

		pollOptions := make([]*content_proto.PollOption, len(options))
		for i, option := range options {
			pollOptions[i] = &content_proto.PollOption{
				Id:        option.ID,
				Text:      option.Text,
				VoteCount: int32(option.VoteCount),
			}
		}

		pollProto = &content_proto.Poll{
			Id:              poll.ID,
			Question:        poll.Question,
			Options:         pollOptions,
			DurationMinutes: int32(poll.DurationMinutes),
			ExpiresAt:       poll.ExpiresAt.Format(time.RFC3339),
		}
	} else if pollData != nil {
		// 用于创建时的临时投票数据
		pollOptions := make([]*content_proto.PollOption, len(pollData.Options))
		for i, optionText := range pollData.Options {
			pollOptions[i] = &content_proto.PollOption{
				Id:        strconv.Itoa(i + 1),
				Text:      optionText,
				VoteCount: 0,
			}
		}

		expiresAt := time.Now().Add(time.Duration(pollData.DurationMinutes) * time.Minute)
		pollProto = &content_proto.Poll{
			Id:              post.ID + "_poll", // 临时ID
			Question:        pollData.Question,
			Options:         pollOptions,
			DurationMinutes: pollData.DurationMinutes,
			ExpiresAt:       expiresAt.Format(time.RFC3339),
		}
	}

	// 构建统计信息
	var replyCount int64
	if post.IsReply && post.ReplyLevel == 2 {
		// 对于二级回复：始终返回0，因为对二级回复的回复会归类到父级一级回复下
		replyCount = 0
	} else if post.IsReply && post.ReplyLevel == 1 {
		// 对于一级回复：统计所有以它为父级的二级回复
		r.db.Model(&models.Post{}).Where("parent_id = ? AND is_reply = true AND deleted_at IS NULL", post.ID).Count(&replyCount)
	} else {
		// 对于原帖：统计所有回复（一级和二级）
		r.db.Model(&models.Post{}).Where("(parent_id = ? OR root_id = ?) AND is_reply = true AND deleted_at IS NULL", post.ID, post.ID).Count(&replyCount)
	}

	stats := &content_proto.PostStats{
		LikeCount:   0,
		ReplyCount:  int32(replyCount),
		RepostCount: 0,
		ViewCount:   0,
	}

	// 构建作者信息（从用户服务获取真实数据）
	fmt.Printf("[Content Repository] About to call fetchUserInfo for user: %s\n", post.UserID)
	author := r.fetchUserInfo(context.Background(), post.UserID)
	fmt.Printf("[Content Repository] fetchUserInfo returned author: %+v\n", author)
	
	// Security check: reject if user does not exist
	if author == nil {
		fmt.Printf("[Content Repository] User %s does not exist, rejecting post creation\n", post.UserID)
		return nil, fmt.Errorf("user does not exist: %s", post.UserID)
	}

	return &content_proto.Post{
		Id:               post.ID,
		UserId:           post.UserID,
		Content:          post.Content,
		Visibility:       post.Visibility,
		ReplyPermission:  post.ReplyPermission,
		ParentId:         r.stringPtrToString(post.ParentID),
		RootId:           r.stringPtrToString(post.RootID),
		RepostId:         r.stringPtrToString(post.RepostID),
		IsReply:          post.IsReply,
		ReplyLevel:       int32(post.ReplyLevel),
		MediaAttachments: mediaAttachments,
		MentionedUsers:   mentionedUsers,
		Tags:             tags,
		Poll:             pollProto,
		Stats:            stats,
		Author:           author,
		CreatedAt:        post.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        post.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// buildPostList 构建帖子列表
func (r *postRepository) buildPostList(posts []models.Post) ([]*content_proto.Post, error) {
	result := make([]*content_proto.Post, len(posts))
	for i, post := range posts {
		// 获取每个帖子的详细信息
		postProto, err := r.GetPost(context.Background(), post.ID, "")
		if err != nil {
			return nil, fmt.Errorf("failed to build post %s: %w", post.ID, err)
		}
		result[i] = postProto
	}
	return result, nil
}

// getMediaURLs 从媒体附件中提取URL列表
func (r *postRepository) getMediaURLs(attachments []models.MediaAttachment) []string {
	urls := make([]string, len(attachments))
	for i, attachment := range attachments {
		urls[i] = attachment.URL
	}
	return urls
}

// stringPtrToString 将字符串指针转换为字符串
func (r *postRepository) stringPtrToString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// fetchUserInfo 从用户服务获取用户信息
func (r *postRepository) fetchUserInfo(ctx context.Context, userID string) *content_proto.Author {
	// 如果没有用户客户端，返回临时数据
	if r.userClient == nil {
		fmt.Printf("[Content Repository] No user client available, using fallback data for user: %s\n", userID)
		return &content_proto.Author{
			Id:          userID,
			Username:    "user_" + userID[:8],
			DisplayName: "User " + userID[:8],
			AvatarUrl:   "",
			IsVerified:  false,
		}
	}

	// 调用用户服务获取用户信息
	fmt.Printf("[Content Repository] Fetching user info for: %s\n", userID)
	
	// Create timeout context for user service call
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	resp, err := r.userClient.GetUser(timeoutCtx, &user_proto.GetUserRequest{
		UserId: userID,
	})

	if err != nil {
		fmt.Printf("[Content Repository] Failed to fetch user info: %v\n", err)
		fmt.Printf("[Content Repository] Error details: %T - %+v\n", err, err)
		return nil // Return nil instead of fallback data for security
	}

	if resp.Error != nil {
		fmt.Printf("[Content Repository] User service returned error: %s, user does not exist\n", resp.Error.Message)
		return nil // Return nil instead of fallback data for security
	}

	if resp.User == nil {
		fmt.Printf("[Content Repository] User service returned nil user, user does not exist\n")
		return nil // Return nil instead of fallback data for security
	}

	// 返回真实用户数据
	fmt.Printf("[Content Repository] Successfully fetched user info: %s (%s)\n", resp.User.Username, resp.User.DisplayName)
	return &content_proto.Author{
		Id:          resp.User.Id,
		Username:    resp.User.Username,
		DisplayName: resp.User.DisplayName,
		AvatarUrl:   resp.User.AvatarUrl,
		IsVerified:  resp.User.IsVerified,
	}
}

// 保留旧的内容仓储方法以保持兼容性
type contentRepository struct {
	db *gorm.DB
}

func NewContentRepository() ContentRepository {
	return &contentRepository{
		db: database.GetDB(),
	}
}

// CreateContent 创建内容
func (r *contentRepository) CreateContent(ctx context.Context, content *content_proto.Post) error {
	// 创建帖子
	createdAt, _ := time.Parse(time.RFC3339, content.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, content.UpdatedAt)
	post := &models.Post{
		ID:        content.Id,
		UserID:    content.UserId,
		Content:   content.Content,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	if err := r.db.Create(post).Error; err != nil {
		return err
	}

	// 创建媒体附件
	for _, mediaFile := range content.MediaAttachments {
		media := &models.MediaAttachment{
			ID:     mediaFile.Id,
			PostID: &content.Id, // 使用指针
			URL:    mediaFile.Url,
			Type:   mediaFile.Type,
		}

		if err := r.db.Create(media).Error; err != nil {
			return err
		}
	}

	return nil
}

// GetContent 获取内容
func (r *contentRepository) GetContent(ctx context.Context, contentID string) (*content_proto.Post, error) {
	var post models.Post
	if err := r.db.Where("id = ? AND deleted_at IS NULL", contentID).First(&post).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("content not found")
		}
		return nil, fmt.Errorf("failed to get content: %w", err)
	}

	// 获取媒体附件
	var mediaAttachments []models.MediaAttachment
	r.db.Where("post_id = ? AND deleted_at IS NULL", post.ID).Find(&mediaAttachments)

	mediaFiles := make([]*content_proto.MediaAttachment, len(mediaAttachments))
	for i, media := range mediaAttachments {
		mediaFiles[i] = &content_proto.MediaAttachment{
			Id:   media.ID,
			Url:  media.URL,
			Type: media.Type,
		}
	}

	return &content_proto.Post{
		Id:               post.ID,
		UserId:           post.UserID,
		Content:          post.Content,
		MediaAttachments: mediaFiles,
		CreatedAt:        post.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        post.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// UpdateContent 更新内容
func (r *contentRepository) UpdateContent(ctx context.Context, content *content_proto.Post) error {
	// 更新帖子
	updatedAt, _ := time.Parse(time.RFC3339, content.UpdatedAt)
	post := &models.Post{
		ID:        content.Id,
		Content:   content.Content,
		UpdatedAt: updatedAt,
	}

	if err := r.db.Where("id = ? AND deleted_at IS NULL", content.Id).Updates(post).Error; err != nil {
		return err
	}

	// 更新媒体附件（简化处理，实际应支持增删改）
	for _, mediaFile := range content.MediaAttachments {
		postID := content.Id
		media := &models.MediaAttachment{
			ID:     mediaFile.Id,
			PostID: &postID, // 使用指针
			URL:    mediaFile.Url,
			Type:   mediaFile.Type,
		}

		if err := r.db.Where("id = ?", mediaFile.Id).Updates(media).Error; err != nil {
			return err
		}
	}

	return nil
}

// DeleteContent 删除内容
func (r *contentRepository) DeleteContent(ctx context.Context, id string) error {
	// 软删除帖子
	if err := r.db.Where("id = ?", id).Delete(&models.Post{}).Error; err != nil {
		return err
	}

	// 同时软删除相关的媒体附件
	return r.db.Where("post_id = ?", id).Delete(&models.MediaAttachment{}).Error
}

// ListContent 列出内容
func (r *contentRepository) ListContent(ctx context.Context, userID string, page, pageSize int32) ([]*content_proto.Post, int32, error) {
	var posts []models.Post
	var total int64

	// 查询总数
	if err := r.db.Model(&models.Post{}).Where("user_id = ? AND deleted_at IS NULL", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询帖子列表
	offset := (page - 1) * pageSize
	if err := r.db.Where("user_id = ? AND deleted_at IS NULL", userID).Offset(int(offset)).Limit(int(pageSize)).Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	contents := make([]*content_proto.Post, len(posts))
	for i, post := range posts {
		// 获取媒体附件
		var mediaAttachments []models.MediaAttachment
		r.db.Where("post_id = ? AND deleted_at IS NULL", post.ID).Find(&mediaAttachments)

		mediaFiles := make([]*content_proto.MediaAttachment, len(mediaAttachments))
		for j, media := range mediaAttachments {
			mediaFiles[j] = &content_proto.MediaAttachment{
				Id:   media.ID,
				Url:  media.URL,
				Type: media.Type,
			}
		}

		contents[i] = &content_proto.Post{
			Id:               post.ID,
			UserId:           post.UserID,
			Content:          post.Content,
			MediaAttachments: mediaFiles,
			CreatedAt:        post.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:        post.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return contents, int32(total), nil
}

// getMediaVariants 从Media服务获取媒体variants信息
func (r *postRepository) getMediaVariants(ctx context.Context, mediaID string) *content_proto.MediaVariants {
	if r.mediaClient == nil {
		log.Printf("[Content Repository] Media client not available, returning nil variants for media ID: %s", mediaID)
		return nil
	}

	// Use the provided context with timeout, preserving authentication headers
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	log.Printf("[Content Repository] Calling media service GetFile for media ID: %s", mediaID)

	// 调用Media服务获取文件信息
	resp, err := r.mediaClient.GetFile(timeoutCtx, &media_proto.GetFileRequest{
		FileId: mediaID,
	})
	if err != nil {
		log.Printf("[Content Repository] Failed to get media variants from media service for ID %s: %v", mediaID, err)
		return nil
	}

	if resp.File == nil || resp.File.Variants == nil {
		log.Printf("[Content Repository] No variants found for media ID: %s", mediaID)
		return nil
	}

	// 转换Media服务的variants到Content服务的proto格式
	variants := &content_proto.MediaVariants{}

	if resp.File.Variants.Thumbnail != nil {
		variants.Thumbnail = &content_proto.MediaVariant{
			Url:    resp.File.Variants.Thumbnail.Url,
			Width:  resp.File.Variants.Thumbnail.Width,
			Height: resp.File.Variants.Thumbnail.Height,
			Size:   resp.File.Variants.Thumbnail.Size,
		}
	}

	if resp.File.Variants.Small != nil {
		variants.Small = &content_proto.MediaVariant{
			Url:    resp.File.Variants.Small.Url,
			Width:  resp.File.Variants.Small.Width,
			Height: resp.File.Variants.Small.Height,
			Size:   resp.File.Variants.Small.Size,
		}
	}

	if resp.File.Variants.Medium != nil {
		variants.Medium = &content_proto.MediaVariant{
			Url:    resp.File.Variants.Medium.Url,
			Width:  resp.File.Variants.Medium.Width,
			Height: resp.File.Variants.Medium.Height,
			Size:   resp.File.Variants.Medium.Size,
		}
	}

	if resp.File.Variants.Large != nil {
		variants.Large = &content_proto.MediaVariant{
			Url:    resp.File.Variants.Large.Url,
			Width:  resp.File.Variants.Large.Width,
			Height: resp.File.Variants.Large.Height,
			Size:   resp.File.Variants.Large.Size,
		}
	}

	if resp.File.Variants.Original != nil {
		variants.Original = &content_proto.MediaVariant{
			Url:    resp.File.Variants.Original.Url,
			Width:  resp.File.Variants.Original.Width,
			Height: resp.File.Variants.Original.Height,
			Size:   resp.File.Variants.Original.Size,
		}
	}

	// 处理视频特有的variants
	if resp.File.Variants.Preview != nil {
		variants.Preview = &content_proto.MediaVariant{
			Url:    resp.File.Variants.Preview.Url,
			Width:  resp.File.Variants.Preview.Width,
			Height: resp.File.Variants.Preview.Height,
			Size:   resp.File.Variants.Preview.Size,
		}
	}

	if resp.File.Variants.LowRes != nil {
		variants.LowRes = &content_proto.MediaVariant{
			Url:    resp.File.Variants.LowRes.Url,
			Width:  resp.File.Variants.LowRes.Width,
			Height: resp.File.Variants.LowRes.Height,
			Size:   resp.File.Variants.LowRes.Size,
		}
	}

	if resp.File.Variants.MidRes != nil {
		variants.MidRes = &content_proto.MediaVariant{
			Url:    resp.File.Variants.MidRes.Url,
			Width:  resp.File.Variants.MidRes.Width,
			Height: resp.File.Variants.MidRes.Height,
			Size:   resp.File.Variants.MidRes.Size,
		}
	}

	if resp.File.Variants.HighRes != nil {
		variants.HighRes = &content_proto.MediaVariant{
			Url:    resp.File.Variants.HighRes.Url,
			Width:  resp.File.Variants.HighRes.Width,
			Height: resp.File.Variants.HighRes.Height,
			Size:   resp.File.Variants.HighRes.Size,
		}
	}

	log.Printf("[Content Repository] Successfully retrieved variants for media ID: %s", mediaID)
	return variants
}

// GetPostReplies 获取帖子的回复列表 - 优化版本
func (r *postRepository) GetPostReplies(ctx context.Context, postID, requestingUserID string, limit int32, cursor string) ([]*content_proto.Post, string, bool, error) {
	fmt.Printf("[Content Repository] GetPostReplies called for postID: %s\n", postID)

	// 优化查询：使用索引友好的查询方式
	// 1. 先获取直接回复 (reply_level = 1, parent_id = postID)
	// 2. 再获取嵌套回复 (reply_level = 2, root_id = postID)
	query := r.db.Where("((parent_id = ? AND reply_level = 1) OR (root_id = ? AND reply_level = 2)) AND is_reply = true AND deleted_at IS NULL", postID, postID).
		Order("reply_level ASC, created_at ASC") // 改为ASC，保持时间顺序

	// 处理分页
	if cursor != "" {
		if cursorTime, err := time.Parse(time.RFC3339, cursor); err == nil {
			query = query.Where("created_at > ?", cursorTime)
		}
	}

	if limit > 0 {
		query = query.Limit(int(limit + 1)) // +1 用于检查是否有更多数据
	}

	var posts []models.Post
	if err := query.Find(&posts).Error; err != nil {
		return nil, "", false, fmt.Errorf("failed to get post replies: %w", err)
	}

	// 检查是否有更多数据
	hasMore := false
	nextCursor := ""
	if len(posts) > int(limit) {
		hasMore = true
		posts = posts[:limit] // 移除多余的记录
		nextCursor = posts[len(posts)-1].CreatedAt.Format(time.RFC3339)
	}

	// 转换为proto格式
	var protoPosts []*content_proto.Post
	for _, post := range posts {
		protoPost, err := r.buildPostProtoSimple(ctx, &post, requestingUserID)
		if err != nil {
			fmt.Printf("[Content Repository] Failed to build proto for post %s: %v\n", post.ID, err)
			continue
		}
		protoPosts = append(protoPosts, protoPost)
	}

	fmt.Printf("[Content Repository] GetPostReplies returning %d replies\n", len(protoPosts))
	return protoPosts, nextCursor, hasMore, nil
}

// GetConversationThread 获取完整对话线程
func (r *postRepository) GetConversationThread(ctx context.Context, rootID, requestingUserID string, limit int32, cursor string) ([]*content_proto.Post, string, bool, error) {
	fmt.Printf("[Content Repository] GetConversationThread called for rootID: %s\n", rootID)

	// 获取根帖子和所有相关回复
	query := r.db.Where("(id = ? OR root_id = ?) AND deleted_at IS NULL", rootID, rootID).
		Order("reply_level ASC, created_at ASC")

	// 处理分页
	if cursor != "" {
		if cursorTime, err := time.Parse(time.RFC3339, cursor); err == nil {
			query = query.Where("created_at > ?", cursorTime)
		}
	}

	if limit > 0 {
		query = query.Limit(int(limit + 1)) // +1 用于检查是否有更多数据
	}

	var posts []models.Post
	if err := query.Find(&posts).Error; err != nil {
		return nil, "", false, fmt.Errorf("failed to get conversation thread: %w", err)
	}

	// 检查是否有更多数据
	hasMore := false
	nextCursor := ""
	if len(posts) > int(limit) {
		hasMore = true
		posts = posts[:limit] // 移除多余的记录
		nextCursor = posts[len(posts)-1].CreatedAt.Format(time.RFC3339)
	}

	// 转换为proto格式
	var protoPosts []*content_proto.Post
	for _, post := range posts {
		protoPost, err := r.buildPostProtoSimple(ctx, &post, requestingUserID)
		if err != nil {
			fmt.Printf("[Content Repository] Failed to build proto for post %s: %v\n", post.ID, err)
			continue
		}
		protoPosts = append(protoPosts, protoPost)
	}

	fmt.Printf("[Content Repository] GetConversationThread returning %d posts\n", len(protoPosts))
	return protoPosts, nextCursor, hasMore, nil
}

// DeleteReply 删除回复
func (r *postRepository) DeleteReply(ctx context.Context, replyID, userID string) error {
	fmt.Printf("[Content Repository] DeleteReply called for replyID: %s by user: %s\n", replyID, userID)

	// 检查回复是否存在且属于该用户
	var reply models.Post
	if err := r.db.Where("id = ? AND user_id = ? AND is_reply = true AND deleted_at IS NULL", replyID, userID).First(&reply).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("reply not found or permission denied")
		}
		return fmt.Errorf("failed to find reply: %w", err)
	}

	// 软删除回复
	now := time.Now()
	if err := r.db.Model(&reply).Update("deleted_at", now).Error; err != nil {
		return fmt.Errorf("failed to delete reply: %w", err)
	}

	fmt.Printf("[Content Repository] Reply %s deleted successfully\n", replyID)
	return nil
}

// buildPostProtoSimple 构建帖子Proto对象（简化版本，用于回复查询）
func (r *postRepository) buildPostProtoSimple(ctx context.Context, post *models.Post, requestingUserID string) (*content_proto.Post, error) {
	// 获取实际的媒体附件从数据库
	var dbMediaAttachments []models.MediaAttachment
	r.db.Where("post_id = ? AND deleted_at IS NULL", post.ID).Find(&dbMediaAttachments)

	// 构建媒体附件
	mediaAttachments := make([]*content_proto.MediaAttachment, len(dbMediaAttachments))
	for i, media := range dbMediaAttachments {
		variants := r.getMediaVariants(ctx, media.ID)
		mediaAttachments[i] = &content_proto.MediaAttachment{
			Id:       media.ID,
			Url:      media.URL,
			Type:     media.Type,
			MimeType: media.MimeType,
			Width:    int32(media.Width),
			Height:   int32(media.Height),
			Variants: variants,
		}
	}

	// 获取提及的用户
	var mentions []models.PostMention
	r.db.Where("post_id = ?", post.ID).Find(&mentions)
	mentionedUsers := make([]string, len(mentions))
	for i, mention := range mentions {
		mentionedUsers[i] = mention.MentionedUserID
	}

	// 获取标签
	var tags []models.PostTag
	r.db.Where("post_id = ?", post.ID).Find(&tags)
	tagStrings := make([]string, len(tags))
	for i, tag := range tags {
		tagStrings[i] = tag.Tag
	}

	// 获取投票信息
	var poll *content_proto.Poll
	if post.HasPoll {
		var dbPoll models.Poll
		if err := r.db.Where("post_id = ?", post.ID).First(&dbPoll).Error; err == nil {
			var options []models.PollOption
			r.db.Where("poll_id = ?", dbPoll.ID).Order("position").Find(&options)

			pollOptions := make([]*content_proto.PollOption, len(options))
			for i, option := range options {
				pollOptions[i] = &content_proto.PollOption{
					Id:        option.ID,
					Text:      option.Text,
					VoteCount: int32(option.VoteCount),
				}
			}

			poll = &content_proto.Poll{
				Id:              dbPoll.ID,
				Question:        dbPoll.Question,
				Options:         pollOptions,
				DurationMinutes: int32(dbPoll.DurationMinutes),
				ExpiresAt:       dbPoll.ExpiresAt.Format(time.RFC3339),
			}
		}
	}

	// 构建统计信息
	var replyCount int64
	if post.IsReply && post.ReplyLevel == 2 {
		// 对于二级回复：始终返回0，因为对二级回复的回复会归类到父级一级回复下
		replyCount = 0
	} else if post.IsReply && post.ReplyLevel == 1 {
		// 对于一级回复：统计所有以它为父级的二级回复
		r.db.Model(&models.Post{}).Where("parent_id = ? AND is_reply = true AND deleted_at IS NULL", post.ID).Count(&replyCount)
	} else {
		// 对于原帖：统计所有回复（一级和二级）
		r.db.Model(&models.Post{}).Where("(parent_id = ? OR root_id = ?) AND is_reply = true AND deleted_at IS NULL", post.ID, post.ID).Count(&replyCount)
	}

	stats := &content_proto.PostStats{
		LikeCount:   0,
		ReplyCount:  int32(replyCount),
		RepostCount: 0,
		ViewCount:   0,
	}

	// 构建作者信息
	author := r.fetchUserInfo(context.Background(), post.UserID)
	
	// Security check: reject if user does not exist
	if author == nil {
		fmt.Printf("[Content Repository] User %s does not exist in buildPostProtoSimple\n", post.UserID)
		return nil, fmt.Errorf("user does not exist: %s", post.UserID)
	}

	return &content_proto.Post{
		Id:               post.ID,
		UserId:           post.UserID,
		Content:          post.Content,
		Visibility:       post.Visibility,
		ReplyPermission:  post.ReplyPermission,
		ParentId:         r.stringPtrToString(post.ParentID),
		RootId:           r.stringPtrToString(post.RootID),
		RepostId:         r.stringPtrToString(post.RepostID),
		IsReply:          post.IsReply,
		ReplyLevel:       int32(post.ReplyLevel),
		HasMedia:         post.HasMedia,
		HasPoll:          post.HasPoll,
		MediaAttachments: mediaAttachments,
		MentionedUsers:   mentionedUsers,
		Tags:             tagStrings,
		Poll:             poll,
		Stats:            stats,
		Author:           author,
		CreatedAt:        post.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        post.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// GetReplyMention 获取回复提及信息
func (r *postRepository) GetReplyMention(ctx context.Context, replyID string) (*content_proto.ReplyMention, error) {
	db := database.GetDB()

	var replyMention models.ReplyMention
	if err := db.Where("reply_id = ?", replyID).First(&replyMention).Error; err != nil {
		return nil, fmt.Errorf("failed to get reply mention: %w", err)
	}

	return &content_proto.ReplyMention{
		Id:              replyMention.ID,
		ReplyId:         replyMention.ReplyID,
		MentionedUserId: replyMention.MentionedUserID,
		CreatedAt:       replyMention.CreatedAt.Format(time.RFC3339),
	}, nil
}
