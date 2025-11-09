package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/flick/backend/pkg/auth"
	"github.com/flick/backend/pkg/config"
	"github.com/flick/backend/pkg/messagebus"
	"github.com/flick/backend/services/user/internal/repository"
	"github.com/flick/backend/services/user/proto"

	"go.uber.org/zap"
)

// List of blocked email domains to prevent temporary email registrations.
// In a real-world application, this should be managed via a config file or database.
var blockedEmailDomains = map[string]struct{}{
	"mailinator.com":    {},
	"temp-mail.org":     {},
	"10minutemail.com":  {},
	"guerrillamail.com": {},
	"yopmail.com":       {},
}

// userService 用户服务实现
type UserService struct {
	userRepo *repository.UserRepository
	cfg         *config.Config
	logger      *zap.Logger
	messageBus  messagebus.MessageBus
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo *repository.UserRepository, cfg *config.Config, logger *zap.Logger, messageBus messagebus.MessageBus) *UserService {
	return &UserService{
		userRepo:   userRepo,
		cfg:        cfg,
		logger:     logger,
		messageBus: messageBus,
	}
}

// GetUser 获取用户信息
func (s *UserService) GetUser(ctx context.Context, req *proto.GetUserRequest) (*proto.GetUserResponse, error) {
	fmt.Printf("[User Service] GetUser request for userID: %s\n", req.UserId)
	
	user, err := s.userRepo.GetUserByID(ctx, req.UserId)
	if err != nil {
		fmt.Printf("[User Service] GetUser failed for userID %s: %v\n", req.UserId, err)
		return &proto.GetUserResponse{
			Error: &proto.Error{
				Code:    404,
				Message: "Failed to get user: " + err.Error(),
			},
		}, err
	}

	fmt.Printf("[User Service] GetUser success for userID %s: %s (avatar: %s)\n", req.UserId, user.Username, user.AvatarUrl)
	return &proto.GetUserResponse{
		User: user,
	}, nil
}

// GetUserByUsername implements the gRPC method.
func (s *UserService) GetUserByUsername(ctx context.Context, req *proto.GetUserByUsernameRequest) (*proto.GetUserResponse, error) {
	s.logger.Info("Fetching user by username", zap.String("username", req.Username))
	user, err := s.userRepo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		s.logger.Warn("Failed to get user by username", zap.String("username", req.Username), zap.Error(err))
		return &proto.GetUserResponse{
			Error: &proto.Error{
				Code:    404,
				Message: "User not found: " + err.Error(),
			},
		}, nil // Return nil error to client, error details are in the response message
	}

	return &proto.GetUserResponse{
		User: user,
	}, nil
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(ctx context.Context, req *proto.UpdateUserRequest) (*proto.UpdateUserResponse, error) {
	// 首先获取现有用户信息
	existingUser, err := s.userRepo.GetUserByID(ctx, req.UserId)
	if err != nil {
		return &proto.UpdateUserResponse{
			Error: &proto.Error{
				Code:    404,
				Message: "User not found: " + err.Error(),
			},
		}, err
	}

	// 更新用户字段
	if req.Username != "" {
		existingUser.Username = req.Username
	}

	if req.DisplayName != "" {
		existingUser.DisplayName = req.DisplayName
	}

	if req.Email != "" {
		existingUser.Email = req.Email
	}

	if req.Phone != "" {
		existingUser.Phone = req.Phone
	}

	if req.AvatarUrl != "" {
		existingUser.AvatarUrl = req.AvatarUrl
	}

	if req.BannerUrl != "" {
		existingUser.BannerUrl = req.BannerUrl
	}

	if req.Bio != "" {
		existingUser.Bio = req.Bio
	}

	if req.Location != "" {
		existingUser.Location = req.Location
	}

	if req.WebsiteUrl != "" {
		existingUser.WebsiteUrl = req.WebsiteUrl
	}

	if req.Status != "" {
		existingUser.Status = req.Status
	}

	// 保存更新
	err = s.userRepo.UpdateUser(ctx, existingUser)
	if err != nil {
		return &proto.UpdateUserResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to update user: " + err.Error(),
			},
		}, err
	}

	return &proto.UpdateUserResponse{
		User: existingUser,
	}, nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(ctx context.Context, req *proto.DeleteUserRequest) (*proto.DeleteUserResponse, error) {
	err := s.userRepo.DeleteUser(ctx, req.UserId)
	if err != nil {
		return &proto.DeleteUserResponse{
			Success: false,
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to delete user: " + err.Error(),
			},
		}, err
	}

	return &proto.DeleteUserResponse{
		Success: true,
	}, nil
}

// Register 用户注册
func (s *UserService) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	// Validate email domain against the blocklist
	if err := s.validateEmailDomain(req.Email); err != nil {
		s.logger.Warn("Registration blocked for disposable email", zap.String("email", req.Email), zap.Error(err))
		return &proto.RegisterResponse{
			Error: &proto.Error{
				Code:    400, // Bad Request
				Message: err.Error(),
			},
		}, nil
	}

	// 检查用户是否已存在
	_, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err == nil {
		s.logger.Warn("Registration failed: user already exists", zap.String("email", req.Email))
		return &proto.RegisterResponse{
			Error: &proto.Error{
				Code:    409,
				Message: "User with this email already exists",
			},
		}, nil
	}

	// 哈希密码
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("Failed to hash password during registration", zap.Error(err))
		return &proto.RegisterResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to hash password: " + err.Error(),
			},
		}, err
	}

	// 创建新用�?
	newUser := &proto.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		DisplayName:  req.DisplayName,
		LoginMethod:  "password",
		Status:       "active",
	}

	// 保存用户
	err = s.userRepo.CreateUser(ctx, newUser)
	if err != nil {
		s.logger.Error("Failed to create user in database", zap.String("username", req.Username), zap.Error(err))
		return &proto.RegisterResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to create user: " + err.Error(),
			},
		}, err
	}

	// 生成JWT令牌
	token, err := auth.GenerateJWT(newUser, s.cfg.JWTSecret)
	if err != nil {
		s.logger.Error("Failed to generate token after registration", zap.String("userID", newUser.Id), zap.Error(err))
		return &proto.RegisterResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to generate token: " + err.Error(),
			},
		}, err
	}

	s.logger.Info("User registered successfully", zap.String("username", newUser.Username), zap.String("userID", newUser.Id))
	return &proto.RegisterResponse{
		Token: token,
		User:  newUser,
	}, nil
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	// 用户认证
	user, err := s.userRepo.Authenticate(ctx, req.Identifier, req.Password)
	if err != nil {
		s.logger.Warn("Authentication failed for identifier", zap.String("identifier", req.Identifier), zap.Error(err))
		return &proto.LoginResponse{
			Error: &proto.Error{
				Code:    401,
				Message: "Authentication failed: " + err.Error(),
			},
		}, nil
	}

	// 生成JWT令牌
	token, err := auth.GenerateJWT(user, s.cfg.JWTSecret)
	if err != nil {
		s.logger.Error("Failed to generate token after login", zap.String("userID", user.Id), zap.Error(err))
		return &proto.LoginResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to generate token: " + err.Error(),
			},
		}, err
	}

	s.logger.Info("User logged in successfully", zap.String("username", user.Username), zap.String("userID", user.Id))
	return &proto.LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

// GetFollowers 获取关注�?
func (s *UserService) GetFollowers(ctx context.Context, req *proto.GetFollowersRequest) (*proto.GetFollowersResponse, error) {
	users, pageInfo, err := s.userRepo.GetFollowers(ctx, req.UserId, int(req.First), req.After)
	if err != nil {
		return &proto.GetFollowersResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to get followers: " + err.Error(),
			},
		}, err
	}

	return &proto.GetFollowersResponse{
		Users:    users,
		PageInfo: pageInfo,
	}, nil
}

// GetFollowing 获取正在关注
func (s *UserService) GetFollowing(ctx context.Context, req *proto.GetFollowingRequest) (*proto.GetFollowingResponse, error) {
	users, pageInfo, err := s.userRepo.GetFollowing(ctx, req.UserId, int(req.First), req.After)
	if err != nil {
		return &proto.GetFollowingResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to get following: " + err.Error(),
			},
		}, err
	}

	return &proto.GetFollowingResponse{
		Users:    users,
		PageInfo: pageInfo,
	}, nil
}

// UpdateProfile 更新个人资料
func (s *UserService) UpdateProfile(ctx context.Context, req *proto.UpdateProfileRequest) (*proto.UpdateProfileResponse, error) {
	// Debug: Log request values
	fmt.Printf("[User Service] UpdateProfile request:\n")
	fmt.Printf("  UserId: %s\n", req.UserId)
	if req.DisplayName != nil {
		fmt.Printf("  DisplayName: %s\n", *req.DisplayName)
	}
	if req.Bio != nil {
		fmt.Printf("  Bio: %s\n", *req.Bio)
	}
	if req.Location != nil {
		fmt.Printf("  Location: %s\n", *req.Location)
	}
	if req.Website != nil {
		fmt.Printf("  Website: %s\n", *req.Website)
	}
	if req.AvatarUrl != nil {
		fmt.Printf("  AvatarUrl: %s\n", *req.AvatarUrl)
	}
	if req.BannerUrl != nil {
		fmt.Printf("  BannerUrl: %s\n", *req.BannerUrl)
	}

	// 首先获取现有用户信息
	existingUser, err := s.userRepo.GetUserByID(ctx, req.UserId)
	if err != nil {
		return &proto.UpdateProfileResponse{
			Error: &proto.Error{
				Code:    404,
				Message: "User not found: " + err.Error(),
			},
		}, err
	}

	// 更新用户字段
	if req.DisplayName != nil {
		existingUser.DisplayName = *req.DisplayName
	}

	if req.Bio != nil {
		existingUser.Bio = *req.Bio
	}

	if req.Location != nil {
		existingUser.Location = *req.Location
	}

	if req.Website != nil {
		existingUser.WebsiteUrl = *req.Website
	}

	if req.AvatarUrl != nil {
		existingUser.AvatarUrl = *req.AvatarUrl
	}

	if req.BannerUrl != nil {
		existingUser.BannerUrl = *req.BannerUrl
	}

	// 保存更新
	fmt.Printf("[User Service] About to save user profile updates to database\n")
	err = s.userRepo.UpdateUser(ctx, existingUser)
	if err != nil {
		fmt.Printf("[User Service] Failed to save profile updates: %v\n", err)
		return &proto.UpdateProfileResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to update profile: " + err.Error(),
			},
		}, err
	}

	fmt.Printf("[User Service] Profile updated successfully for user %s, new avatar: %s\n", existingUser.Username, existingUser.AvatarUrl)
	return &proto.UpdateProfileResponse{
		User: existingUser,
	}, nil
}

// FollowUser 关注用户
func (s *UserService) FollowUser(ctx context.Context, req *proto.FollowUserRequest) (*proto.FollowUserResponse, error) {
	err := s.userRepo.FollowUser(ctx, req.FollowerId, req.FollowingId)
	if err != nil {
		return &proto.FollowUserResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to follow user: " + err.Error(),
			},
		}, err
	}

	user, err := s.userRepo.GetUserByID(ctx, req.FollowingId)
	if err != nil {
		return &proto.FollowUserResponse{
			Error: &proto.Error{
				Code:    404,
				Message: "Failed to get user: " + err.Error(),
			},
		}, err
	}

	return &proto.FollowUserResponse{
		User: user,
	}, nil
}

// UnfollowUser 取消关注用户
func (s *UserService) UnfollowUser(ctx context.Context, req *proto.UnfollowUserRequest) (*proto.UnfollowUserResponse, error) {
	err := s.userRepo.UnfollowUser(ctx, req.FollowerId, req.FollowingId)
	if err != nil {
		return &proto.UnfollowUserResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to unfollow user: " + err.Error(),
			},
		}, err
	}

	user, err := s.userRepo.GetUserByID(ctx, req.FollowingId)
	if err != nil {
		return &proto.UnfollowUserResponse{
			Error: &proto.Error{
				Code:    404,
				Message: "Failed to get user: " + err.Error(),
			},
		}, err
	}

	return &proto.UnfollowUserResponse{
		User: user,
	}, nil
}

// UpdateUserAvatar 更新用户头像版本和URL
func (s *UserService) UpdateUserAvatar(ctx context.Context, req *proto.UpdateUserAvatarRequest) (*proto.UpdateUserAvatarResponse, error) {
	fmt.Printf("[User Service] UpdateUserAvatar request for userID: %s, version: %d, URL: %s\n", 
		req.UserId, req.AvatarVersion, req.AvatarUrl)
	
	// 更新用户头像信息
	err := s.userRepo.UpdateUserAvatar(ctx, req.UserId, req.AvatarUrl, req.AvatarVersion)
	if err != nil {
		fmt.Printf("[User Service] UpdateUserAvatar failed for userID %s: %v\n", req.UserId, err)
		return &proto.UpdateUserAvatarResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to update user avatar: " + err.Error(),
			},
		}, err
	}
	
	// 获取更新后的用户信息
	user, err := s.userRepo.GetUserByID(ctx, req.UserId)
	if err != nil {
		return &proto.UpdateUserAvatarResponse{
			Error: &proto.Error{
				Code:    404,
				Message: "Failed to get updated user: " + err.Error(),
			},
		}, err
	}
	
	// 发布头像更新事件到消息总线
	if s.messageBus != nil {
		err = messagebus.PublishUserAvatarUpdated(ctx, s.messageBus, user.Id, user.Username, user.AvatarUrl, user.AvatarVersion)
		if err != nil {
			fmt.Printf("[User Service] Warning: Failed to publish avatar update event: %v\n", err)
			// 不返回错误，因为主要操作已经成功
		} else {
			fmt.Printf("[User Service] Published avatar update event for user: %s\n", user.Username)
		}
	}
	
	fmt.Printf("[User Service] UpdateUserAvatar successful for userID: %s\n", req.UserId)
	return &proto.UpdateUserAvatarResponse{
		User: user,
	}, nil
}

// UpdateUserBanner 更新用户横幅版本和URL
func (s *UserService) UpdateUserBanner(ctx context.Context, req *proto.UpdateUserBannerRequest) (*proto.UpdateUserBannerResponse, error) {
	fmt.Printf("[User Service] UpdateUserBanner request for userID: %s, version: %d, URL: %s\n", 
		req.UserId, req.BannerVersion, req.BannerUrl)
	
	// 更新用户横幅信息
	err := s.userRepo.UpdateUserBanner(ctx, req.UserId, req.BannerUrl, req.BannerVersion)
	if err != nil {
		fmt.Printf("[User Service] UpdateUserBanner failed for userID %s: %v\n", req.UserId, err)
		return &proto.UpdateUserBannerResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to update user banner: " + err.Error(),
			},
		}, err
	}
	
	// 获取更新后的用户信息
	user, err := s.userRepo.GetUserByID(ctx, req.UserId)
	if err != nil {
		return &proto.UpdateUserBannerResponse{
			Error: &proto.Error{
				Code:    404,
				Message: "Failed to get updated user: " + err.Error(),
			},
		}, err
	}
	
	// 发布横幅更新事件到消息总线
	if s.messageBus != nil {
		err = messagebus.PublishUserBannerUpdated(ctx, s.messageBus, user.Id, user.Username, user.BannerUrl, user.BannerVersion)
		if err != nil {
			fmt.Printf("[User Service] Warning: Failed to publish banner update event: %v\n", err)
			// 不返回错误，因为主要操作已经成功
		} else {
			fmt.Printf("[User Service] Published banner update event for user: %s\n", user.Username)
		}
	}
	
	fmt.Printf("[User Service] UpdateUserBanner successful for userID: %s\n", req.UserId)
	return &proto.UpdateUserBannerResponse{
		User: user,
	}, nil
}

// validateEmailDomain checks if the email domain is in the blocklist.
func (s *UserService) validateEmailDomain(email string) error {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return errors.New("invalid email format")
	}
	domain := parts[1]
	if _, blocked := blockedEmailDomains[strings.ToLower(domain)]; blocked {
		return errors.New("registration with this email provider is not allowed")
	}
	return nil
}

