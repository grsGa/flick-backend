package repository

import (
	"context"
	"errors"
	"time"

	"github.com/flick/backend/pkg/auth"
	"github.com/flick/backend/pkg/database"
	"github.com/flick/backend/pkg/models"
	"github.com/flick/backend/services/user/proto"

	"gorm.io/gorm"
)

// userRepository 用户仓储实现
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储实例
func NewUserRepository() UserRepository {
	return &userRepository{
		db: database.GetDB(),
	}
}

// CreateUser 创建用户
func (r *userRepository) CreateUser(ctx context.Context, user *proto.User) error {
	modelUser := r.protoToModel(user)
	return r.db.Create(modelUser).Error
}

// GetUserByID 根据ID获取用户
func (r *userRepository) GetUserByID(ctx context.Context, id string) (*proto.User, error) {
	var user models.User
	if err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return r.modelToProto(&user), nil
}

// UpdateUser 更新用户
func (r *userRepository) UpdateUser(ctx context.Context, user *proto.User) error {
	// 只更新提供的字段
	updates := map[string]interface{}{}

	if user.Username != "" {
		updates["username"] = user.Username
	}

	if user.DisplayName != "" {
		updates["display_name"] = user.DisplayName
	}

	if user.Email != "" {
		updates["email"] = user.Email
	}

	if user.Phone != "" {
		updates["phone"] = user.Phone
	}

	if user.AvatarUrl != "" {
		updates["avatar_url"] = user.AvatarUrl
	}

	if user.BannerUrl != "" {
		updates["banner_url"] = user.BannerUrl
	}

	if user.Bio != "" {
		updates["bio"] = user.Bio
	}

	if user.Location != "" {
		updates["location"] = user.Location
	}

	if user.WebsiteUrl != "" {
		updates["website_url"] = user.WebsiteUrl
	}

	if user.Status != "" {
		updates["status"] = user.Status
	}

	updates["updated_at"] = time.Now()

	return r.db.Model(&models.User{}).Where("id = ? AND deleted_at IS NULL", user.Id).Updates(updates).Error
}

// DeleteUser 删除用户（软删除）
func (r *userRepository) DeleteUser(ctx context.Context, id string) error {
	// 软删除
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}

// GetUserByUsername 根据用户名获取用户
func (r *userRepository) GetUserByUsername(ctx context.Context, username string) (*proto.User, error) {
	var user models.User
	if err := r.db.Where("username = ? AND deleted_at IS NULL", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return r.modelToProto(&user), nil
}

// GetUserByEmail 根据邮箱获取用户
func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*proto.User, error) {
	var user models.User
	// Log the email parameter
	println("GetUserByEmail: email =", email)
	// Log the generated SQL query
	sql := r.db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Where("email = ? AND deleted_at IS NULL", email).First(&user)
	})
	println("GetUserByEmail: sql =", sql)
	if err := r.db.Where("email = ? AND deleted_at IS NULL", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return r.modelToProto(&user), nil
}

// Authenticate 用户认证
func (r *userRepository) Authenticate(ctx context.Context, identifier, password string) (*proto.User, error) {
	var user models.User

	// 根据用户名或邮箱查找用户
	if err := r.db.Where("username = ? OR email = ?", identifier, identifier).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 为了安全起见，不区分是用户不存在还是密码错误
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	// 验证密码哈希
	if !auth.CheckPasswordHash(password, user.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	// 更新最后登录时间
	r.db.Model(&user).Update("last_login_at", time.Now())

	return r.modelToProto(&user), nil
}

// modelToProto 将模型转换为protobuf消息
func (r *userRepository) modelToProto(user *models.User) *proto.User {
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

// protoToModel 将protobuf消息转换为模型
func (r *userRepository) protoToModel(user *proto.User) *models.User {
	modelUser := &models.User{
		ID:              user.Id,
		Username:        user.Username,
		DisplayName:     user.DisplayName,
		Email:           user.Email,
		PasswordHash:    user.PasswordHash,
		AvatarURL:       user.AvatarUrl,
		FollowersCount:  int(user.FollowersCount),
		FollowingCount:  int(user.FollowingCount),
		IsFollowing:     user.IsFollowing,
		IsVerified:      user.IsVerified,
		IsEmailVerified: user.IsEmailVerified,
		IsPhoneVerified: user.IsPhoneVerified,
		LoginMethod:     user.LoginMethod,
		Status:          user.Status,
	}

	if user.Phone != "" {
		modelUser.Phone = &user.Phone
	}

	if user.PasswordHash != "" {
		modelUser.PasswordHash = user.PasswordHash
	}

	if user.BannerUrl != "" {
		modelUser.BannerURL = &user.BannerUrl
	}

	if user.Bio != "" {
		modelUser.Bio = &user.Bio
	}

	if user.Location != "" {
		modelUser.Location = &user.Location
	}

	if user.WebsiteUrl != "" {
		modelUser.WebsiteURL = &user.WebsiteUrl
	}

	if user.LastLoginAt != "" {
		if t, err := time.Parse(time.RFC3339, user.LastLoginAt); err == nil {
			modelUser.LastLoginAt = &t
		}
	}

	if user.CreatedAt != "" {
		if t, err := time.Parse(time.RFC3339, user.CreatedAt); err == nil {
			modelUser.CreatedAt = t
		}
	}

	if user.UpdatedAt != "" {
		if t, err := time.Parse(time.RFC3339, user.UpdatedAt); err == nil {
			modelUser.UpdatedAt = t
		}
	}

	return modelUser
}

// toString 将*string转换为string
func toString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// GetFollowers 获取关注者
func (r *userRepository) GetFollowers(ctx context.Context, userID string, first int, after string) ([]*proto.User, *proto.PageInfo, error) {
	var users []*models.User
	db := r.db.Joins("JOIN follows ON follows.follower_id = users.id").
		Where("follows.followee_id = ?", userID).
		Limit(first)

	if after != "" {
		db = db.Where("users.id > ?", after)
	}

	if err := db.Find(&users).Error; err != nil {
		return nil, nil, err
	}

	pbUsers := make([]*proto.User, len(users))
	for i, u := range users {
		pbUsers[i] = r.modelToProto(u)
	}

	var endCursor string
	if len(users) > 0 {
		endCursor = users[len(users)-1].ID
	}

	var count int64
	r.db.Model(&models.User{}).Joins("JOIN follows ON follows.follower_id = users.id").
		Where("follows.followee_id = ? AND users.id > ?", userID, endCursor).
		Count(&count)
	hasNextPage := count > 0

	pageInfo := &proto.PageInfo{
		HasNextPage: hasNextPage,
		EndCursor:   endCursor,
	}

	return pbUsers, pageInfo, nil
}

// GetFollowing 获取正在关注
func (r *userRepository) GetFollowing(ctx context.Context, userID string, first int, after string) ([]*proto.User, *proto.PageInfo, error) {
	var users []*models.User
	db := r.db.Joins("JOIN follows ON follows.followee_id = users.id").
		Where("follows.follower_id = ?", userID).
		Limit(first)

	if after != "" {
		db = db.Where("users.id > ?", after)
	}

	if err := db.Find(&users).Error; err != nil {
		return nil, nil, err
	}

	pbUsers := make([]*proto.User, len(users))
	for i, u := range users {
		pbUsers[i] = r.modelToProto(u)
	}

	var endCursor string
	if len(users) > 0 {
		endCursor = users[len(users)-1].ID
	}

	var count int64
	r.db.Model(&models.User{}).Joins("JOIN follows ON follows.followee_id = users.id").
		Where("follows.follower_id = ? AND users.id > ?", userID, endCursor).
		Count(&count)
	hasNextPage := count > 0

	pageInfo := &proto.PageInfo{
		HasNextPage: hasNextPage,
		EndCursor:   endCursor,
	}

	return pbUsers, pageInfo, nil
}

// FollowUser 关注用户
func (r *userRepository) FollowUser(ctx context.Context, followerID, followingID string) error {
	follow := models.Follow{
		FollowerID: followerID,
		FolloweeID: followingID,
	}
	return r.db.Create(&follow).Error
}

// UnfollowUser 取消关注用户
func (r *userRepository) UnfollowUser(ctx context.Context, followerID, followingID string) error {
	return r.db.Where("follower_id = ? AND followee_id = ?", followerID, followingID).Delete(&models.Follow{}).Error
}
