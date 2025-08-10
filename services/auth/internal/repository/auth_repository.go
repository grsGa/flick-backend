package repository

import (
	"context"
	"errors"
	"time"

	"backend/pkg/auth"
	"backend/pkg/database"
	"backend/pkg/models"
	"backend/services/auth/proto"

	"gorm.io/gorm"
)

// authRepository 认证仓储实现
type authRepository struct {
	db *gorm.DB
}

// NewAuthRepository 创建认证仓储实例
func NewAuthRepository() AuthRepository {
	return &authRepository{
		db: database.GetDB(),
	}
}

// GetUserByIdentifier 根据标识符（用户名、邮箱或手机号）获取用户
func (r *authRepository) GetUserByIdentifier(ctx context.Context, identifier string) (*proto.User, error) {
	var user models.User
	if err := r.db.Where("(username = ? OR email = ? OR phone = ?) AND deleted_at IS NULL", identifier, identifier, identifier).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return r.modelToProto(&user), nil
}

// VerifyPassword 验证密码
func (r *authRepository) VerifyPassword(ctx context.Context, userID string, password string) error {
	var user models.User
	if err := r.db.Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
		return err
	}

	if !auth.CheckPasswordHash(password, user.PasswordHash) {
		return errors.New("invalid password")
	}

	return nil
}

// GetUserByID 根据ID获取用户
func (r *authRepository) GetUserByID(ctx context.Context, id string) (*proto.User, error) {
	var user models.User
	if err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return r.modelToProto(&user), nil
}

// CreateUser 创建用户
func (r *authRepository) CreateUser(ctx context.Context, user *proto.User, password string) (*proto.User, error) {
	// 对密码进行哈希处理
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	u := &models.User{
		ID:              user.Id,
		Username:        user.Username,
		DisplayName:     user.DisplayName,
		Email:           user.Email,
		PasswordHash:    passwordHash,
		AvatarURL:       user.AvatarUrl,
		IsEmailVerified: user.IsEmailVerified,
		IsPhoneVerified: user.IsPhoneVerified,
		LoginMethod:     user.LoginMethod,
		Status:          user.Status,
	}

	if user.Phone != "" {
		u.Phone = &user.Phone
	}

	if user.BannerUrl != "" {
		u.BannerURL = &user.BannerUrl
	}

	if user.Bio != "" {
		u.Bio = &user.Bio
	}

	if user.Location != "" {
		u.Location = &user.Location
	}

	if user.WebsiteUrl != "" {
		u.WebsiteURL = &user.WebsiteUrl
	}

	if user.CreatedAt != "" {
		if t, err := time.Parse(time.RFC3339, user.CreatedAt); err == nil {
			u.CreatedAt = t
		}
	} else {
		u.CreatedAt = time.Now()
	}

	u.UpdatedAt = time.Now()

	if err := r.db.Create(u).Error; err != nil {
		return nil, err
	}
	return r.modelToProto(u), nil
}

// UpdateUserLoginInfo 更新用户登录信息
func (r *authRepository) UpdateUserLoginInfo(ctx context.Context, userID string, lastLoginAt string) error {
	updates := map[string]interface{}{
		"last_login_at": time.Now(),
		"updated_at":    time.Now(),
	}

	return r.db.Model(&models.User{}).Where("id = ? AND deleted_at IS NULL", userID).Updates(updates).Error
}

// CreateSession 创建会话
func (r *authRepository) CreateSession(ctx context.Context, session *Session) error {
	// 在实际实现中，这里应该创建会话记录
	// 由于目前没有会话表，暂时返回nil
	return nil
}

// GetSession 获取会话
func (r *authRepository) GetSession(ctx context.Context, refreshToken string) (*Session, error) {
	// 在实际实现中，这里应该查询会话记录
	// 由于目前没有会话表，暂时返回错误
	return nil, errors.New("session not found")
}

// DeleteSession 删除会话
func (r *authRepository) DeleteSession(ctx context.Context, refreshToken string) error {
	// 在实际实现中，这里应该删除会话记录
	// 由于目前没有会话表，暂时返回nil
	return nil
}

// DeleteUserSessions 删除用户所有会话
func (r *authRepository) DeleteUserSessions(ctx context.Context, userID string) error {
	// 在实际实现中，这里应该删除用户所有会话记录
	// 由于目前没有会话表，暂时返回nil
	return nil
}

// modelToProto 将模型转换为protobuf消息
func (r *authRepository) modelToProto(user *models.User) *proto.User {
	pbUser := &proto.User{
		Id:              user.ID,
		Username:        user.Username,
		DisplayName:     user.DisplayName,
		Email:           user.Email,
		AvatarUrl:       user.AvatarURL,
		Bio:             toString(user.Bio),
		Location:        toString(user.Location),
		WebsiteUrl:      toString(user.WebsiteURL),
		FollowersCount:  int32(user.FollowersCount),
		FollowingCount:  int32(user.FollowingCount),
		IsFollowing:     user.IsFollowing,
		IsVerified:      user.IsVerified,
		IsEmailVerified: user.IsEmailVerified,
		IsPhoneVerified: user.IsPhoneVerified,
		LoginMethod:     user.LoginMethod,
		Status:          user.Status,
		CreatedAt:       user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       user.UpdatedAt.Format(time.RFC3339),
	}

	if user.Phone != nil {
		pbUser.Phone = *user.Phone
	}

	if user.BannerURL != nil {
		pbUser.BannerUrl = *user.BannerURL
	}

	if user.LastLoginAt != nil {
		pbUser.LastLoginAt = user.LastLoginAt.Format(time.RFC3339)
	}

	return pbUser
}

// toString 将*string转换为string
func toString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
