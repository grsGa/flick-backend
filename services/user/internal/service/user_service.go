package service

import (
	"context"

	"backend/pkg/auth"
	"backend/pkg/config"
	"backend/services/user/internal/repository"
	"backend/services/user/proto"

	"go.uber.org/zap"
)

// userService 用户服务实现
type userService struct {
	userRepo repository.UserRepository
	cfg      *config.Config
	logger   *zap.Logger
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo repository.UserRepository, cfg *config.Config, logger *zap.Logger) UserService {
	return &userService{
		userRepo: userRepo,
		cfg:      cfg,
		logger:   logger,
	}
}

// GetUser 获取用户信息
func (s *userService) GetUser(ctx context.Context, req *proto.GetUserRequest) (*proto.GetUserResponse, error) {
	user, err := s.userRepo.GetUserByID(ctx, req.UserId)
	if err != nil {
		return &proto.GetUserResponse{
			Error: &proto.Error{
				Code:    404,
				Message: "Failed to get user: " + err.Error(),
			},
		}, err
	}

	return &proto.GetUserResponse{
		User: user,
	}, nil
}

// UpdateUser 更新用户
func (s *userService) UpdateUser(ctx context.Context, req *proto.UpdateUserRequest) (*proto.UpdateUserResponse, error) {
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

	if req.CoverUrl != "" {
		existingUser.CoverUrl = req.CoverUrl
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
func (s *userService) DeleteUser(ctx context.Context, req *proto.DeleteUserRequest) (*proto.DeleteUserResponse, error) {
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
func (s *userService) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
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

	// 创建新用户
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
func (s *userService) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
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
