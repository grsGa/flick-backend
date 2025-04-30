package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend/pkg/models"

	"gorm.io/gorm"
)

// PostgresRepository 实现了UserRepository接口
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository 创建一个新的PostgresSQL仓库实例
func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

// CreateUser 创建新用户
func (r *PostgresRepository) CreateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetUserByID 根据ID获取用户
func (r *PostgresRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Preload("Roles").First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Preload("Roles").Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByUsername 根据用户名获取用户
func (r *PostgresRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Preload("Roles").Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByUsernameIgnoreCase 根据用户名查询用户，不区分大小写
func (r *PostgresRepository) GetUserByUsernameIgnoreCase(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	// 使用LOWER函数进行大小写不敏感的匹配
	if err := r.db.WithContext(ctx).Preload("Roles").
		Where("LOWER(username) = LOWER(?)", username).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser 更新用户信息
func (r *PostgresRepository) UpdateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// DeleteUser 软删除用户
func (r *PostgresRepository) DeleteUser(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, id).Error
}

// ListUsers 获取用户列表
func (r *PostgresRepository) ListUsers(ctx context.Context, offset, limit int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	if err := r.db.WithContext(ctx).Preload("Roles").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// SearchUsers 搜索用户
func (r *PostgresRepository) SearchUsers(ctx context.Context, query string, offset, limit int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	// 构建查询条件
	searchQuery := "%" + query + "%"
	queryDB := r.db.WithContext(ctx).Model(&models.User{}).
		Where("username LIKE ? OR display_name LIKE ? OR email LIKE ?", searchQuery, searchQuery, searchQuery)

	// 获取总数
	if err := queryDB.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	if err := queryDB.Preload("Roles").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetUserRoles 获取用户角色
func (r *PostgresRepository) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	var roles []models.Role
	if err := r.db.WithContext(ctx).Model(&models.User{ID: userID}).Association("Roles").Find(&roles); err != nil {
		return nil, err
	}

	// 提取角色名称
	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}

	return roleNames, nil
}

// AssignRoleToUser 为用户分配角色
func (r *PostgresRepository) AssignRoleToUser(ctx context.Context, userID, roleName string) error {
	var role models.Role
	if err := r.db.WithContext(ctx).Where("name = ?", roleName).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("role not found: %w", err)
		}
		return err
	}

	var user models.User
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user not found: %w", err)
		}
		return err
	}

	return r.db.WithContext(ctx).Model(&user).Association("Roles").Append(&role)
}

// RemoveRoleFromUser 从用户移除角色
func (r *PostgresRepository) RemoveRoleFromUser(ctx context.Context, userID, roleName string) error {
	var role models.Role
	if err := r.db.WithContext(ctx).Where("name = ?", roleName).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("role not found: %w", err)
		}
		return err
	}

	var user models.User
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user not found: %w", err)
		}
		return err
	}

	return r.db.WithContext(ctx).Model(&user).Association("Roles").Delete(&role)
}

// FollowUser 关注用户
func (r *PostgresRepository) FollowUser(ctx context.Context, followerID, followingID string) error {
	// 检查是否已经关注
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.UserFollow{}).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("already following")
	}

	// 创建关注关系
	follow := models.UserFollow{
		FollowerID:  followerID,
		FollowingID: followingID,
	}

	return r.db.WithContext(ctx).Create(&follow).Error
}

// UnfollowUser 取消关注用户
func (r *PostgresRepository) UnfollowUser(ctx context.Context, followerID, followingID string) error {
	return r.db.WithContext(ctx).Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Delete(&models.UserFollow{}).Error
}

// GetFollowers 获取关注者列表
func (r *PostgresRepository) GetFollowers(ctx context.Context, userID string, offset, limit int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&models.UserFollow{}).
		Where("following_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	if err := r.db.WithContext(ctx).
		Joins("JOIN flick_user_follows ON flick_user_follows.follower_id = flick_users.id").
		Where("flick_user_follows.following_id = ?", userID).
		Preload("Roles").
		Offset(offset).Limit(limit).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetFollowing 获取正在关注的用户列表
func (r *PostgresRepository) GetFollowing(ctx context.Context, userID string, offset, limit int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&models.UserFollow{}).
		Where("follower_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	if err := r.db.WithContext(ctx).
		Joins("JOIN flick_user_follows ON flick_user_follows.following_id = flick_users.id").
		Where("flick_user_follows.follower_id = ?", userID).
		Preload("Roles").
		Offset(offset).Limit(limit).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// IsFollowing 检查是否关注
func (r *PostgresRepository) IsFollowing(ctx context.Context, followerID, followingID string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.UserFollow{}).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetFollowStats 获取关注统计
func (r *PostgresRepository) GetFollowStats(ctx context.Context, userID string) (int64, int64, error) {
	// 获取粉丝数
	var followers int64
	if err := r.db.WithContext(ctx).Model(&models.UserFollow{}).
		Where("following_id = ?", userID).Count(&followers).Error; err != nil {
		return 0, 0, err
	}

	// 获取关注数
	var following int64
	if err := r.db.WithContext(ctx).Model(&models.UserFollow{}).
		Where("follower_id = ?", userID).Count(&following).Error; err != nil {
		return 0, 0, err
	}

	return followers, following, nil
}

// BlockUser 屏蔽用户
func (r *PostgresRepository) BlockUser(ctx context.Context, blockerID, blockedID string, reason string) error {
	// 检查是否已经屏蔽
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.UserBlock{}).
		Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).
		Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("already blocked")
	}

	// 创建屏蔽关系
	block := models.UserBlock{
		BlockerID: blockerID,
		BlockedID: blockedID,
		Reason:    reason,
	}

	return r.db.WithContext(ctx).Create(&block).Error
}

// UnblockUser 解除屏蔽
func (r *PostgresRepository) UnblockUser(ctx context.Context, blockerID, blockedID string) error {
	return r.db.WithContext(ctx).Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).
		Delete(&models.UserBlock{}).Error
}

// GetBlockedUsers 获取已屏蔽用户列表
func (r *PostgresRepository) GetBlockedUsers(ctx context.Context, userID string, offset, limit int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&models.UserBlock{}).
		Where("blocker_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	if err := r.db.WithContext(ctx).
		Joins("JOIN flick_user_blocks ON flick_user_blocks.blocked_id = flick_users.id").
		Where("flick_user_blocks.blocker_id = ?", userID).
		Preload("Roles").
		Offset(offset).Limit(limit).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// IsBlocked 检查是否屏蔽
func (r *PostgresRepository) IsBlocked(ctx context.Context, blockerID, blockedID string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.UserBlock{}).
		Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// CreateVerification 创建验证记录
func (r *PostgresRepository) CreateVerification(ctx context.Context, verification *models.UserVerification) error {
	// 首先删除同类型的旧验证记录
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND type = ? AND used = ?", verification.UserID, verification.Type, false).
		Delete(&models.UserVerification{}).Error; err != nil {
		return err
	}

	return r.db.WithContext(ctx).Create(verification).Error
}

// GetVerification 获取验证记录
func (r *PostgresRepository) GetVerification(ctx context.Context, userID, verificationType string) (*models.UserVerification, error) {
	var verification models.UserVerification
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND type = ? AND used = ? AND expires_at > ?", userID, verificationType, false, time.Now()).
		First(&verification).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("verification not found: %w", err)
		}
		return nil, err
	}
	return &verification, nil
}

// VerifyToken 验证令牌
func (r *PostgresRepository) VerifyToken(ctx context.Context, userID, verificationType, token string) (bool, error) {
	var verification models.UserVerification
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND type = ? AND token = ? AND used = ? AND expires_at > ?",
			userID, verificationType, token, false, time.Now()).
		First(&verification).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	// 标记为已使用
	verification.Used = true
	if err := r.db.WithContext(ctx).Save(&verification).Error; err != nil {
		return true, err
	}

	return true, nil
}

// MarkVerificationUsed 标记验证为已使用
func (r *PostgresRepository) MarkVerificationUsed(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&models.UserVerification{}).
		Where("id = ?", id).
		Update("used", true).Error
}

// CreateSession 创建会话
func (r *PostgresRepository) CreateSession(ctx context.Context, session *models.UserSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// GetSession 获取会话
func (r *PostgresRepository) GetSession(ctx context.Context, id string) (*models.UserSession, error) {
	var session models.UserSession
	if err := r.db.WithContext(ctx).First(&session, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("session not found: %w", err)
		}
		return nil, err
	}
	return &session, nil
}

// DeleteSession 删除会话
func (r *PostgresRepository) DeleteSession(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.UserSession{}, id).Error
}

// DeleteUserSessions 删除用户所有会话
func (r *PostgresRepository) DeleteUserSessions(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.UserSession{}).Error
}

// LogUserActivity 记录用户活动
func (r *PostgresRepository) LogUserActivity(ctx context.Context, activity *models.UserActivity) error {
	return r.db.WithContext(ctx).Create(activity).Error
}

// GetUserActivities 获取用户活动日志
func (r *PostgresRepository) GetUserActivities(ctx context.Context, userID string, offset, limit int) ([]*models.UserActivity, int64, error) {
	var activities []*models.UserActivity
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&models.UserActivity{}).
		Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&activities).Error; err != nil {
		return nil, 0, err
	}

	return activities, total, nil
}

// GetActiveSessionsForUser 获取用户的活跃会话
func (r *PostgresRepository) GetActiveSessionsForUser(ctx context.Context, userID string) ([]*models.UserSession, error) {
	var sessions []*models.UserSession

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Find(&sessions).Error

	if err != nil {
		return nil, err
	}

	return sessions, nil
}

// UpdateSession 更新会话信息
func (r *PostgresRepository) UpdateSession(ctx context.Context, session *models.UserSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}
