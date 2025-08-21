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

// postRepository 帖子仓储实现
type postRepository struct {
	db          *gorm.DB
	userClient  user_proto.UserServiceClient
	mediaClient media_proto.MediaServiceClient
}

// NewPostRepository 创建帖子仓储实例
func NewPostRepository() PostRepository {
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
		db:          database.GetDB(),
		userClient:  userClient,
		mediaClient: mediaClient,
	}
}

// CreatePost 创建帖子
func (r *postRepository) CreatePost(ctx context.Context, req *content_proto.CreatePostRequest) (*content_proto.Post, error) {
	fmt.Printf("[Content Repository] CreatePost called with:\n")
	fmt.Printf("  UserID: %s\n", req.UserId)
	fmt.Printf("  Content: %s\n", req.Content)
	fmt.Printf("  MediaUrls: %v\n", req.MediaUrls)

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

	// 设置父帖子ID（回复）
	if req.ParentId != "" {
		post.ParentID = &req.ParentId
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
			
			// 重试逻辑：最多等待6秒让Media服务完成variants处理
			maxRetries := 3
			retryDelay := 2 * time.Second
			
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

	// 构建返回的帖子对象
	fmt.Printf("[Content Repository] Building response with committed data\n")
	return r.buildPostProto(post, req.MediaUrls, req.MentionedUsers, req.Tags, poll, req.PollData)
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

	return r.buildPostProto(&post, r.getMediaURLs(mediaAttachments), mentionedUsers, tagStrings, poll, nil)
}

// GetUserPosts 获取用户帖子列表
func (r *postRepository) GetUserPosts(ctx context.Context, userID, requestingUserID string, limit int32, cursor string) ([]*content_proto.Post, string, bool, error) {
	fmt.Printf("[Content Repository] GetUserPosts called for userID: %s, requestingUserID: %s, limit: %d\n", userID, requestingUserID, limit)

	var posts []models.Post
	query := r.db.Where("user_id = ? AND deleted_at IS NULL", userID).
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
		nextCursor = posts[len(posts)-1].ID // 使用post ID作为cursor，与处理逻辑一致
	}

	fmt.Printf("[Content Repository] GetUserPosts result: posts=%d, hasMore=%t, nextCursor=%s\n", len(posts), hasMore, nextCursor)
	return postList, nextCursor, hasMore, nil
}

// GetTimeline 获取时间线帖子
func (r *postRepository) GetTimeline(ctx context.Context, userID string, limit int32, cursor string) ([]*content_proto.Post, string, bool, error) {
	fmt.Printf("[Content Repository] GetTimeline called for userID: %s, limit: %d\n", userID, limit)

	// 简化实现：获取所有公开帖子，实际应该根据关注关系过滤
	var posts []models.Post
	query := r.db.Where("visibility = 'public' AND deleted_at IS NULL").
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

	return postList, nextCursor, hasMore, nil
}

// DeletePost 删除帖子（软删除）
func (r *postRepository) DeletePost(ctx context.Context, id, userID string) error {
	result := r.db.Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).
		Update("deleted_at", time.Now())

	if result.Error != nil {
		return fmt.Errorf("failed to delete post: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("post not found or not authorized")
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
func (r *postRepository) buildPostProto(post *models.Post, mediaURLs []string, mentionedUsers []string, tags []string, poll *models.Poll, pollData *content_proto.PollData) (*content_proto.Post, error) {
	// 获取实际的媒体附件从数据库
	var dbMediaAttachments []models.MediaAttachment
	r.db.Where("post_id = ? AND deleted_at IS NULL", post.ID).Find(&dbMediaAttachments)

	// 构建媒体附件 - 使用数据库中的实际数据并获取variants信息
	mediaAttachments := make([]*content_proto.MediaAttachment, len(dbMediaAttachments))
	for i, media := range dbMediaAttachments {
		// 从Media服务获取完整的媒体信息包括variants
		variants := r.getMediaVariants(media.ID)

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

	// 构建统计信息（临时数据）
	stats := &content_proto.PostStats{
		LikeCount:    0,
		CommentCount: 0,
		RepostCount:  0,
		ViewCount:    0,
	}

	// 构建作者信息（从用户服务获取真实数据）
	fmt.Printf("[Content Repository] About to call fetchUserInfo for user: %s\n", post.UserID)
	author := r.fetchUserInfo(context.Background(), post.UserID)
	fmt.Printf("[Content Repository] fetchUserInfo returned author: %+v\n", author)

	return &content_proto.Post{
		Id:               post.ID,
		UserId:           post.UserID,
		Content:          post.Content,
		Visibility:       post.Visibility,
		ReplyPermission:  post.ReplyPermission,
		ParentId:         r.stringPtrToString(post.ParentID),
		RepostId:         r.stringPtrToString(post.RepostID),
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
	resp, err := r.userClient.GetUser(ctx, &user_proto.GetUserRequest{
		UserId: userID,
	})

	if err != nil {
		fmt.Printf("[Content Repository] Failed to fetch user info: %v, using fallback data\n", err)
		return &content_proto.Author{
			Id:          userID,
			Username:    "user_" + userID[:8],
			DisplayName: "User " + userID[:8],
			AvatarUrl:   "",
			IsVerified:  false,
		}
	}

	if resp.Error != nil {
		fmt.Printf("[Content Repository] User service returned error: %s, using fallback data\n", resp.Error.Message)
		return &content_proto.Author{
			Id:          userID,
			Username:    "user_" + userID[:8],
			DisplayName: "User " + userID[:8],
			AvatarUrl:   "",
			IsVerified:  false,
		}
	}

	if resp.User == nil {
		fmt.Printf("[Content Repository] User service returned nil user, using fallback data\n")
		return &content_proto.Author{
			Id:          userID,
			Username:    "user_" + userID[:8],
			DisplayName: "User " + userID[:8],
			AvatarUrl:   "",
			IsVerified:  false,
		}
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
func (r *postRepository) getMediaVariants(mediaID string) *content_proto.MediaVariants {
	if r.mediaClient == nil {
		log.Printf("[Content Repository] Media client not available, returning nil variants for media ID: %s", mediaID)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 调用Media服务获取文件信息
	resp, err := r.mediaClient.GetFile(ctx, &media_proto.GetFileRequest{
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

	log.Printf("[Content Repository] Successfully retrieved variants for media ID: %s", mediaID)
	return variants
}
