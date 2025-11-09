package repository

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/flick/backend/pkg/database"
	"github.com/flick/backend/pkg/models"
	"github.com/flick/backend/services/interaction/proto"
	"gorm.io/gorm"
)

type InteractionRepository struct {
	db *gorm.DB
}

// NewInteractionRepository 创建互动仓储实例
func NewInteractionRepository() *InteractionRepository {
	return &InteractionRepository{db: database.GetDB()}
}

// NewPostgresRepository 创建PostgreSQL仓储实现（保持向后兼容）
func NewPostgresRepository() *InteractionRepository {
	return &InteractionRepository{db: database.GetDB()}
}

// CreateFollow 创建关注
func (r *InteractionRepository) CreateFollow(ctx context.Context, follow *proto.Follow) error {
	model := &models.Follow{
		FollowerID: follow.FollowerId,
		FolloweeID: follow.FolloweeId,
		CreatedAt:  time.Now(),
	}
	return r.db.WithContext(ctx).Create(model).Error
}

// DeleteFollow 删除关注
func (r *InteractionRepository) DeleteFollow(ctx context.Context, followerID, followeeID string) error {
	return r.db.WithContext(ctx).
		Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
		Delete(&models.Follow{}).Error
}

// IsFollowing 检查是否关�?
func (r *InteractionRepository) IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Follow{}).
		Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
		Count(&count).Error
	return count > 0, err
}

// GetFollow 获取关注记录
func (r *InteractionRepository) GetFollow(ctx context.Context, followerID, followeeID string) (*proto.Follow, error) {
	var follow models.Follow
	err := r.db.WithContext(ctx).
		Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
		First(&follow).Error
	if err != nil {
		return nil, err
	}

	return &proto.Follow{
		Id:         follow.ID,
		FollowerId: follow.FollowerID,
		FolloweeId: follow.FolloweeID,
		CreatedAt:  follow.CreatedAt.Format(time.RFC3339),
	}, nil
}

// GetFollowers 获取粉丝列表
func (r *InteractionRepository) GetFollowers(ctx context.Context, userID string, page, pageSize int32) ([]*proto.Follow, int32, error) {
	var follows []models.Follow
	var total int64

	offset := (page - 1) * pageSize

	// 获取总数
	if err := r.db.WithContext(ctx).
		Model(&models.Follow{}).
		Where("followee_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	if err := r.db.WithContext(ctx).
		Where("followee_id = ?", userID).
		Offset(int(offset)).
		Limit(int(pageSize)).
		Order("created_at DESC").
		Find(&follows).Error; err != nil {
		return nil, 0, err
	}

	// 转换为proto格式
	result := make([]*proto.Follow, len(follows))
	for i, follow := range follows {
		result[i] = &proto.Follow{
			Id:         follow.ID,
			FollowerId: follow.FollowerID,
			FolloweeId: follow.FolloweeID,
			CreatedAt:  follow.CreatedAt.Format(time.RFC3339),
		}
	}

	return result, int32(total), nil
}

// GetFollowing 获取关注列表
func (r *InteractionRepository) GetFollowing(ctx context.Context, userID string, page, pageSize int32) ([]*proto.Follow, int32, error) {
	var follows []models.Follow
	var total int64

	offset := (page - 1) * pageSize

	// 获取总数
	if err := r.db.WithContext(ctx).
		Model(&models.Follow{}).
		Where("follower_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	if err := r.db.WithContext(ctx).
		Where("follower_id = ?", userID).
		Offset(int(offset)).
		Limit(int(pageSize)).
		Order("created_at DESC").
		Find(&follows).Error; err != nil {
		return nil, 0, err
	}

	// 转换为proto格式
	result := make([]*proto.Follow, len(follows))
	for i, follow := range follows {
		result[i] = &proto.Follow{
			Id:         follow.ID,
			FollowerId: follow.FollowerID,
			FolloweeId: follow.FolloweeID,
			CreatedAt:  follow.CreatedAt.Format(time.RFC3339),
		}
	}

	return result, int32(total), nil
}

// CreateLike 创建点赞
func (r *InteractionRepository) CreateLike(ctx context.Context, like *proto.Like) error {
	model := &models.Like{
		UserID:    like.UserId,
		PostID:    like.PostId,
		CreatedAt: time.Now(),
	}
	return r.db.WithContext(ctx).Create(model).Error
}

// DeleteLike 删除点赞
func (r *InteractionRepository) DeleteLike(ctx context.Context, userID, postID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Delete(&models.Like{}).Error
}

// IsLiked 检查是否点�?
func (r *InteractionRepository) IsLiked(ctx context.Context, userID, postID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Like{}).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Count(&count).Error
	return count > 0, err
}

// GetLikes 获取点赞列表
func (r *InteractionRepository) GetLikes(ctx context.Context, postID string, page, pageSize int32) ([]*proto.Like, int32, error) {
	var likes []models.Like
	var total int64

	offset := (page - 1) * pageSize

	// 获取总数
	if err := r.db.WithContext(ctx).
		Model(&models.Like{}).
		Where("post_id = ?", postID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	if err := r.db.WithContext(ctx).
		Where("post_id = ?", postID).
		Offset(int(offset)).
		Limit(int(pageSize)).
		Order("created_at DESC").
		Find(&likes).Error; err != nil {
		return nil, 0, err
	}

	// 转换为proto格式
	result := make([]*proto.Like, len(likes))
	for i, like := range likes {
		result[i] = &proto.Like{
			Id:        like.ID,
			UserId:    like.UserID,
			PostId:    like.PostID,
			CreatedAt: like.CreatedAt.Format(time.RFC3339),
		}
	}

	return result, int32(total), nil
}

// CreateRepost 创建转发
func (r *InteractionRepository) CreateRepost(ctx context.Context, repost *proto.Repost) error {
	model := &models.Repost{
		UserID:    repost.UserId,
		PostID:    repost.PostId,
		CreatedAt: time.Now(),
	}
	if repost.Reply != "" {
		model.Reply = &repost.Reply
	}
	return r.db.WithContext(ctx).Create(model).Error
}

// DeleteRepost 删除转发
func (r *InteractionRepository) DeleteRepost(ctx context.Context, userID, postID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Delete(&models.Repost{}).Error
}

// CreateReply 创建回复 - 回复现在作为Post处理，由content服务管理
func (r *InteractionRepository) CreateReply(ctx context.Context, reply *proto.Reply) error {
	// 回复功能已迁移到content服务，这里只需要返回成�?
	// 实际的回复创建通过content服务的CreatePost完成
	return nil
}

// DeleteReply 删除回复 - 回复现在作为Post处理，由content服务管理
func (r *InteractionRepository) DeleteReply(ctx context.Context, replyID, userID string) error {
	// 回复删除功能已迁移到content服务
	return nil
}

// GetReplies 获取回复列表 - 回复现在作为Post处理，由content服务管理
func (r *InteractionRepository) GetReplies(ctx context.Context, postID string, limit int32, cursor string) ([]*proto.Reply, string, bool, error) {
	// 回复功能已迁移到content服务，返回空列表
	return []*proto.Reply{}, "", false, nil
}

// GetPostStats 获取帖子统计
func (r *InteractionRepository) GetPostStats(ctx context.Context, postID string) (*proto.PostStats, error) {
	var stats models.PostStats
	err := r.db.WithContext(ctx).Where("post_id = ?", postID).First(&stats).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果统计记录不存在，创建一个默认的
			stats = models.PostStats{
				PostID:      postID,
				LikeCount:   0,
				ReplyCount:  0,
				RepostCount: 0,
				ViewCount:   0,
				UpdatedAt:   time.Now(),
			}
			r.db.WithContext(ctx).Create(&stats)
		} else {
			return nil, err
		}
	}

	return &proto.PostStats{
		PostId:      stats.PostID,
		LikeCount:   int32(stats.LikeCount),
		ReplyCount:  int32(stats.ReplyCount),
		RepostCount: int32(stats.RepostCount),
		ViewCount:   int32(stats.ViewCount),
		UpdatedAt:   stats.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// UpdatePostStats 更新帖子统计
func (r *InteractionRepository) UpdatePostStats(ctx context.Context, postID, action string, delta int32) (*proto.PostStats, error) {
	var stats models.PostStats

	// 使用事务确保原子�?
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先查询或创建统计记录
		err := tx.Where("post_id = ?", postID).First(&stats).Error
		if err == gorm.ErrRecordNotFound {
			stats = models.PostStats{
				PostID:      postID,
				LikeCount:   0,
				ReplyCount:  0,
				RepostCount: 0,
				ViewCount:   0,
				UpdatedAt:   time.Now(),
			}
			if err := tx.Create(&stats).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		// 根据动作更新对应的计�?
		updates := map[string]interface{}{
			"updated_at": time.Now(),
		}

		switch action {
		case "like", "unlike":
			stats.LikeCount += int(delta)
			updates["like_count"] = stats.LikeCount
		case "reply", "unreply":
			stats.ReplyCount += int(delta)
			updates["reply_count"] = stats.ReplyCount
		case "repost", "unrepost":
			stats.RepostCount += int(delta)
			updates["repost_count"] = stats.RepostCount
		case "view":
			stats.ViewCount += int(delta)
			updates["view_count"] = stats.ViewCount
		default:
			return fmt.Errorf("unknown action: %s", action)
		}

		return tx.Model(&stats).Updates(updates).Error
	})

	if err != nil {
		return nil, err
	}

	return &proto.PostStats{
		PostId:      stats.PostID,
		LikeCount:   int32(stats.LikeCount),
		ReplyCount:  int32(stats.ReplyCount),
		RepostCount: int32(stats.RepostCount),
		ViewCount:   int32(stats.ViewCount),
		UpdatedAt:   stats.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// VotePoll 投票
func (r *InteractionRepository) VotePoll(ctx context.Context, vote *proto.PollVote) error {
	model := &models.PollVote{
		PollID:       vote.PollId,
		PollOptionID: vote.PollOptionId,
		UserID:       vote.UserId,
		CreatedAt:    time.Now(),
	}
	return r.db.WithContext(ctx).Create(model).Error
}

// CreateBookmark 创建收藏
func (r *InteractionRepository) CreateBookmark(ctx context.Context, bookmark *proto.Bookmark) error {
	model := &models.Bookmark{
		UserID:    bookmark.UserId,
		PostID:    bookmark.PostId,
		CreatedAt: time.Now(),
	}
	return r.db.WithContext(ctx).Create(model).Error
}

// DeleteBookmark 删除收藏
func (r *InteractionRepository) DeleteBookmark(ctx context.Context, userID, postID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Delete(&models.Bookmark{}).Error
}

// IsBookmarked 检查是否收�?
func (r *InteractionRepository) IsBookmarked(ctx context.Context, userID, postID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Bookmark{}).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Count(&count).Error
	return count > 0, err
}

// GetBookmarks 获取收藏列表
func (r *InteractionRepository) GetBookmarks(ctx context.Context, userID string, limit int32, cursor string) ([]*proto.Bookmark, string, bool, error) {
	var bookmarks []models.Bookmark
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)

	// 处理游标分页
	if cursor != "" {
		// 解码游标获取时间�?
		decoded, err := base64.StdEncoding.DecodeString(cursor)
		if err == nil {
			if timestamp, err := strconv.ParseInt(string(decoded), 10, 64); err == nil {
				cursorTime := time.Unix(0, timestamp)
				query = query.Where("created_at < ?", cursorTime)
			}
		}
	}

	if err := query.
		Order("created_at DESC").
		Limit(int(limit + 1)). // 多查一条用于判断是否有更多数据
		Find(&bookmarks).Error; err != nil {
		return nil, "", false, err
	}

	hasMore := len(bookmarks) > int(limit)
	if hasMore {
		bookmarks = bookmarks[:limit] // 移除多查的那一�?
	}

	// 转换为proto格式
	result := make([]*proto.Bookmark, len(bookmarks))
	for i, bookmark := range bookmarks {
		result[i] = &proto.Bookmark{
			Id:        bookmark.ID,
			UserId:    bookmark.UserID,
			PostId:    bookmark.PostID,
			CreatedAt: bookmark.CreatedAt.Format(time.RFC3339),
		}
	}

	// 生成下一页游�?
	var nextCursor string
	if hasMore && len(bookmarks) > 0 {
		lastBookmark := bookmarks[len(bookmarks)-1]
		timestamp := lastBookmark.CreatedAt.UnixNano()
		nextCursor = base64.StdEncoding.EncodeToString([]byte(strconv.FormatInt(timestamp, 10)))
	}

	return result, nextCursor, hasMore, nil
}

// CreateReport 创建举报
func (r *InteractionRepository) CreateReport(ctx context.Context, report *proto.Report) error {
	model := &models.Report{
		UserID:     report.UserId,
		TargetID:   report.TargetId,
		TargetType: report.TargetType,
		Reason:     report.Reason,
		CreatedAt:  time.Now(),
	}
	return r.db.WithContext(ctx).Create(model).Error
}

