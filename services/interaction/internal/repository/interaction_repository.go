package repository

import (
	"context"
	"errors"
	"time"

	"backend/pkg/database"
	"backend/pkg/models"
	"backend/services/interaction/proto"

	"gorm.io/gorm"
)

// interactionRepository 互动仓储实现
type interactionRepository struct {
	db *gorm.DB
}

// NewInteractionRepository 创建互动仓储实例
func NewInteractionRepository() InteractionRepository {
	return &interactionRepository{
		db: database.GetDB(),
	}
}

// CreateFollow 创建关注关系
func (r *interactionRepository) CreateFollow(ctx context.Context, follow *proto.Follow) error {
	createdAt, _ := time.Parse(time.RFC3339, follow.CreatedAt)
	f := &models.Follow{
		ID:         follow.Id,
		FollowerID: follow.FollowerId,
		FolloweeID: follow.FolloweeId,
		CreatedAt:  createdAt,
	}

	return r.db.Create(f).Error
}

// DeleteFollow 删除关注关系
func (r *interactionRepository) DeleteFollow(ctx context.Context, followerID, followeeID string) error {
	return r.db.Where("follower_id = ? AND followee_id = ?", followerID, followeeID).Delete(&models.Follow{}).Error
}

// IsFollowing 检查是否关注
func (r *interactionRepository) IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Follow{}).Where("follower_id = ? AND followee_id = ? AND deleted_at IS NULL", followerID, followeeID).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetFollowers 获取粉丝列表
func (r *interactionRepository) GetFollowers(ctx context.Context, userID string, page, pageSize int32) ([]*proto.Follow, int32, error) {
	var follows []models.Follow
	var total int64

	// 查询总数
	if err := r.db.Model(&models.Follow{}).
		Where("followee_id = ? AND deleted_at IS NULL", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询列表
	offset := (page - 1) * pageSize
	if err := r.db.Where("followee_id = ? AND deleted_at IS NULL", userID).
		Offset(int(offset)).
		Limit(int(pageSize)).
		Find(&follows).Error; err != nil {
		return nil, 0, err
	}

	protoFollows := make([]*proto.Follow, len(follows))
	for i, follow := range follows {
		protoFollows[i] = &proto.Follow{
			Id:         follow.ID,
			FollowerId: follow.FollowerID,
			FolloweeId: follow.FolloweeID,
			CreatedAt:  follow.CreatedAt.Format(time.RFC3339),
		}
	}

	return protoFollows, int32(total), nil
}

// GetFollowing 获取关注列表
func (r *interactionRepository) GetFollowing(ctx context.Context, userID string, page, pageSize int32) ([]*proto.Follow, int32, error) {
	var follows []models.Follow
	var total int64

	// 查询总数
	if err := r.db.Model(&models.Follow{}).
		Where("follower_id = ? AND deleted_at IS NULL", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询列表
	offset := (page - 1) * pageSize
	if err := r.db.Where("follower_id = ? AND deleted_at IS NULL", userID).
		Offset(int(offset)).
		Limit(int(pageSize)).
		Find(&follows).Error; err != nil {
		return nil, 0, err
	}

	protoFollows := make([]*proto.Follow, len(follows))
	for i, follow := range follows {
		protoFollows[i] = &proto.Follow{
			Id:         follow.ID,
			FollowerId: follow.FollowerID,
			FolloweeId: follow.FolloweeID,
			CreatedAt:  follow.CreatedAt.Format(time.RFC3339),
		}
	}

	return protoFollows, int32(total), nil
}

// CreateLike 创建点赞
func (r *interactionRepository) CreateLike(ctx context.Context, like *proto.Like) error {
	createdAt, _ := time.Parse(time.RFC3339, like.CreatedAt)
	l := &models.Like{
		ID:        like.Id,
		UserID:    like.UserId,
		PostID:    like.PostId,
		CreatedAt: createdAt,
	}

	return r.db.Create(l).Error
}

// DeleteLike 删除点赞
func (r *interactionRepository) DeleteLike(ctx context.Context, userID, postID string) error {
	return r.db.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&models.Like{}).Error
}

// IsLiked 检查是否点赞
func (r *interactionRepository) IsLiked(ctx context.Context, userID, postID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Like{}).Where("user_id = ? AND post_id = ? AND deleted_at IS NULL", userID, postID).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetLikes 获取点赞列表
func (r *interactionRepository) GetLikes(ctx context.Context, postID string, page, pageSize int32) ([]*proto.Like, int32, error) {
	var likes []models.Like
	var total int64

	// 查询总数
	if err := r.db.Model(&models.Like{}).
		Where("post_id = ? AND deleted_at IS NULL", postID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询列表
	offset := (page - 1) * pageSize
	if err := r.db.Where("post_id = ? AND deleted_at IS NULL", postID).
		Offset(int(offset)).
		Limit(int(pageSize)).
		Find(&likes).Error; err != nil {
		return nil, 0, err
	}

	protoLikes := make([]*proto.Like, len(likes))
	for i, like := range likes {
		protoLikes[i] = &proto.Like{
			Id:        like.ID,
			UserId:    like.UserID,
			PostId:    like.PostID,
			CreatedAt: like.CreatedAt.Format(time.RFC3339),
		}
	}

	return protoLikes, int32(total), nil
}

// CreateRepost 创建转发
func (r *interactionRepository) CreateRepost(ctx context.Context, repost *proto.Repost) error {
	createdAt, _ := time.Parse(time.RFC3339, repost.CreatedAt)
	rp := &models.Repost{
		ID:        repost.Id,
		UserID:    repost.UserId,
		PostID:    repost.PostId,
		Comment:   &repost.Comment,
		CreatedAt: createdAt,
	}

	return r.db.Create(rp).Error
}

// DeleteRepost 删除转发
func (r *interactionRepository) DeleteRepost(ctx context.Context, userID, postID string) error {
	return r.db.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&models.Repost{}).Error
}

// CreateReport 创建举报
func (r *interactionRepository) CreateReport(ctx context.Context, report *proto.Report) error {
	createdAt, _ := time.Parse(time.RFC3339, report.CreatedAt)
	newReport := &models.Report{
		ID:         report.Id,
		UserID:     report.UserId,
		TargetID:   report.TargetId,
		TargetType: report.TargetType,
		Reason:     report.Reason,
		CreatedAt:  createdAt,
	}

	return r.db.Create(newReport).Error
}

// GetReport 获取举报
func (r *interactionRepository) GetReport(ctx context.Context, id string) (*proto.Report, error) {
	var report models.Report
	if err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&report).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("report not found")
		}
		return nil, err
	}

	return &proto.Report{
		Id:         report.ID,
		UserId:     report.UserID,
		TargetId:   report.TargetID,
		TargetType: report.TargetType,
		Reason:     report.Reason,
		CreatedAt:  report.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
