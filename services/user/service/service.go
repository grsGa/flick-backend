package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"backend/pkg/auth"
	"backend/pkg/config"
	"backend/pkg/models"
	"backend/pkg/storage"
	"backend/services/user/repository"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

// UserService 定义用户服务接口
type UserService interface {
	// 用户注册与认证
	Register(ctx context.Context, email, username, password string) (*models.User, error)
	Login(ctx context.Context, usernameOrEmail, password string) (*models.User, string, string, error) // 返回用户, access token, refresh token, error
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)                     // 返回新的access token, refresh token, error
	Logout(ctx context.Context, userID, sessionID string) error

	// 验证相关
	RequestEmailVerification(ctx context.Context, userID string) error
	VerifyEmail(ctx context.Context, userID, token string) error
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, userID, token, newPassword string) error
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error

	// 用户信息管理
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	UpdateUserProfile(ctx context.Context, userID string, profile *UserProfileUpdate) error
	UpdateAvatar(ctx context.Context, userID string, avatarURL string) error
	UpdateCoverImage(ctx context.Context, userID string, coverImageURL string) error
	DeleteUser(ctx context.Context, userID string) error
	SearchUsers(ctx context.Context, query string, page, pageSize int) ([]*models.User, int64, error)

	// 角色管理
	GetUserRoles(ctx context.Context, userID string) ([]string, error)
	AssignRoleToUser(ctx context.Context, userID, role string) error
	RemoveRoleFromUser(ctx context.Context, userID, role string) error

	// 关系管理
	FollowUser(ctx context.Context, followerID, followingID string) error
	UnfollowUser(ctx context.Context, followerID, followingID string) error
	GetFollowers(ctx context.Context, userID string, page, pageSize int) ([]*models.User, int64, error)
	GetFollowing(ctx context.Context, userID string, page, pageSize int) ([]*models.User, int64, error)
	IsFollowing(ctx context.Context, followerID, followingID string) (bool, error)
	GetFollowStats(ctx context.Context, userID string) (int64, int64, error)

	// 屏蔽管理
	BlockUser(ctx context.Context, blockerID, blockedID string, reason string) error
	UnblockUser(ctx context.Context, blockerID, blockedID string) error
	GetBlockedUsers(ctx context.Context, userID string, page, pageSize int) ([]*models.User, int64, error)
	IsBlocked(ctx context.Context, blockerID, blockedID string) (bool, error)

	// 会话管理
	GetActiveSessionsForUser(ctx context.Context, userID string) ([]*models.UserSession, error)
	TerminateSession(ctx context.Context, userID, sessionID string) error
	TerminateAllSessions(ctx context.Context, userID string) error

	// 活动日志
	GetUserActivityLogs(ctx context.Context, userID string, page, pageSize int) ([]*models.UserActivity, int64, error)
}

// UserProfileUpdate 用户资料更新请求
type UserProfileUpdate struct {
	Username    *string `json:"username,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
	Bio         *string `json:"bio,omitempty"`
	PhoneNumber *string `json:"phone_number,omitempty"`
}

// NewUserRequest 新用户请求
type NewUserRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Username    string `json:"username" binding:"required,min=3,max=30"`
	Password    string `json:"password" binding:"required,min=8"`
	PhoneNumber string `json:"phone_number" binding:"required"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	UsernameOrEmail string `json:"username_or_email" binding:"required"`
	Password        string `json:"password" binding:"required"`
}

// VerificationRequest 验证请求
type VerificationRequest struct {
	Token string `json:"token" binding:"required"`
}

// PasswordResetRequest 密码重置请求
type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// PasswordChangeRequest 密码变更请求
type PasswordChangeRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// PasswordResetConfirmRequest 密码重置确认请求
type PasswordResetConfirmRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// SessionResponse 会话响应
type SessionResponse struct {
	ID           string    `json:"id"`
	Device       string    `json:"device"`
	LastActivity time.Time `json:"last_activity"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// AuthResponse 认证响应
type AuthResponse struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int          `json:"expires_in"` // 过期时间（秒）
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error       string `json:"error"`
	Description string `json:"description,omitempty"`
}

// FollowResponse 关注响应
type FollowResponse struct {
	Success   bool  `json:"success"`
	Followers int64 `json:"followers"`
	Following int64 `json:"following"`
}

// UserServiceImpl 用户服务实现
type UserServiceImpl struct {
	repo        *repository.Repository
	redisClient *redis.Client
	logger      zerolog.Logger
}

// NewUserService 创建用户服务实例
func NewUserService(repo *repository.Repository, redisClient *redis.Client, logger zerolog.Logger) *UserServiceImpl {
	return &UserServiceImpl{
		repo:        repo,
		redisClient: redisClient,
		logger:      logger,
	}
}

// ListUsersHandler 处理列出用户请求
func (s *UserServiceImpl) ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not implemented"))
}

// CreateUserHandler 处理创建用户请求
func (s *UserServiceImpl) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not implemented"))
}

// GetUserHandler 处理获取用户请求
func (s *UserServiceImpl) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not implemented"))
}

// UpdateUserHandler 处理更新用户请求
func (s *UserServiceImpl) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not implemented"))
}

// DeleteUserHandler 处理删除用户请求
func (s *UserServiceImpl) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not implemented"))
}

// GetCurrentUserHandler 处理获取当前用户请求
func (s *UserServiceImpl) GetCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	// 设置内容类型
	w.Header().Set("Content-Type", "application/json")

	// 从请求中获取用户ID
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		// 尝试从Authorization头部获取
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "unauthorized",
				Description: "未提供认证令牌",
			})
			return
		}

		tokenStr := authHeader[7:] // 去掉"Bearer "前缀

		// 解析令牌
		cfg := config.GetConfig()
		token, err := jwt.ParseWithClaims(tokenStr, &auth.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "invalid_token",
				Description: "认证令牌无效",
			})
			return
		}

		claims, ok := token.Claims.(*auth.Claims)
		if !ok || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "invalid_token",
				Description: "认证令牌无效",
			})
			return
		}

		userID = claims.UserID
	}

	// 确保有用户ID
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "unauthorized",
			Description: "无法识别用户",
		})
		return
	}

	// 获取当前用户
	user, err := s.repo.GetUserByID(r.Context(), userID)
	if err != nil {
		s.logger.Error().Err(err).Str("userID", userID).Msg("获取用户失败")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "server_error",
			Description: "服务器内部错误",
		})
		return
	}

	// 获取用户角色
	roles, err := s.repo.GetUserRoles(r.Context(), user.ID)
	if err != nil {
		s.logger.Warn().Err(err).Str("userID", user.ID).Msg("获取用户角色失败")
	} else {
		// 设置用户角色
		for _, roleName := range roles {
			role := models.Role{Name: roleName}
			user.Roles = append(user.Roles, role)
		}
	}

	// 返回用户信息
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// UpdateCurrentUserHandler 处理更新当前用户请求
func (s *UserServiceImpl) UpdateCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	// 设置内容类型
	w.Header().Set("Content-Type", "application/json")

	// 从请求中获取用户ID
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		// 尝试从Authorization头部获取
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "unauthorized",
				Description: "未提供认证令牌",
			})
			return
		}

		tokenStr := authHeader[7:] // 去掉"Bearer "前缀

		// 解析令牌
		cfg := config.GetConfig()
		token, err := jwt.ParseWithClaims(tokenStr, &auth.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "invalid_token",
				Description: "认证令牌无效",
			})
			return
		}

		claims, ok := token.Claims.(*auth.Claims)
		if !ok || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "invalid_token",
				Description: "认证令牌无效",
			})
			return
		}

		userID = claims.UserID
	}

	// 确保有用户ID
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "unauthorized",
			Description: "无法识别用户",
		})
		return
	}

	// 获取当前用户
	user, err := s.repo.GetUserByID(r.Context(), userID)
	if err != nil {
		s.logger.Error().Err(err).Str("userID", userID).Msg("获取用户失败")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "server_error",
			Description: "服务器内部错误",
		})
		return
	}

	// 解析请求体
	var updateReq struct {
		Username        *string `json:"username,omitempty"`
		DisplayName     *string `json:"display_name,omitempty"`
		Bio             *string `json:"bio,omitempty"`
		ProfileComplete *bool   `json:"profile_complete,omitempty"`
	}

	err = json.NewDecoder(r.Body).Decode(&updateReq)
	if err != nil {
		s.logger.Error().Err(err).Msg("无法解析更新请求")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "invalid_request",
			Description: "无法解析请求",
		})
		return
	}

	// 更新字段
	updated := false

	if updateReq.Username != nil && *updateReq.Username != user.Username {
		// 检查用户名是否已存在
		existingUser, err := s.repo.GetUserByUsername(r.Context(), *updateReq.Username)
		if err == nil && existingUser != nil && existingUser.ID != user.ID {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "username_exists",
				Description: "该用户名已被使用",
			})
			return
		}

		user.Username = *updateReq.Username
		updated = true
	}

	if updateReq.DisplayName != nil {
		user.DisplayName = *updateReq.DisplayName
		updated = true
	}

	if updateReq.Bio != nil {
		user.Bio = *updateReq.Bio
		updated = true
	}

	if updateReq.ProfileComplete != nil {
		user.ProfileComplete = *updateReq.ProfileComplete
		updated = true
	}

	// 如果有更新，保存到数据库
	if updated {
		err = s.repo.UpdateUser(r.Context(), user)
		if err != nil {
			s.logger.Error().Err(err).Str("userID", userID).Msg("更新用户资料失败")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "server_error",
				Description: "更新用户资料失败",
			})
			return
		}

		// 记录活动
		activity := &models.UserActivity{
			UserID:      user.ID,
			ActionType:  "profile_update",
			IPAddress:   r.RemoteAddr,
			UserAgent:   r.UserAgent(),
			Description: "更新用户资料",
		}

		if err := s.repo.LogUserActivity(r.Context(), activity); err != nil {
			s.logger.Warn().Err(err).Str("userID", user.ID).Msg("记录活动失败")
		}
	}

	// 返回更新后的用户信息
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// RegisterHandler 处理用户注册请求
func (s *UserServiceImpl) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// 设置内容类型
	w.Header().Set("Content-Type", "application/json")

	// 解析请求体
	var req NewUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		s.logger.Error().Err(err).Msg("无法解析注册请求")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "invalid_request",
			Description: "无法解析请求",
		})
		return
	}

	// 验证请求
	if req.Email == "" || req.Username == "" || req.Password == "" || req.PhoneNumber == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "invalid_request",
			Description: "邮箱、用户名、密码和手机号码为必填项",
		})
		return
	}

	// 验证手机号码格式（中国大陆手机号）
	phonePattern := `^1[3-9]\d{9}$`
	matched, _ := regexp.MatchString(phonePattern, req.PhoneNumber)
	if !matched {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "invalid_phone",
			Description: "请输入有效的中国大陆手机号码",
		})
		return
	}

	if len(req.Password) < 8 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "invalid_password",
			Description: "密码长度必须至少为8个字符",
		})
		return
	}

	// 检查邮箱是否已存在
	existingUser, err := s.repo.GetUserByEmail(r.Context(), req.Email)
	if err == nil && existingUser != nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "email_exists",
			Description: "该邮箱已被注册",
		})
		return
	}

	// 检查用户名是否已存在
	existingUser, err = s.repo.GetUserByUsername(r.Context(), req.Username)
	if err == nil && existingUser != nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "username_exists",
			Description: "该用户名已被使用",
		})
		return
	}

	// 创建用户
	user := &models.User{
		Email:         req.Email,
		Username:      req.Username,
		PhoneNumber:   req.PhoneNumber,
		DisplayName:   req.Username, // 默认使用用户名作为显示名
		AccountStatus: "active",
	}

	// 设置密码（Hash处理）
	err = user.SetPassword(req.Password)
	if err != nil {
		s.logger.Error().Err(err).Msg("密码加密失败")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "server_error",
			Description: "服务器内部错误",
		})
		return
	}

	// 保存用户
	err = s.repo.CreateUser(r.Context(), user)
	if err != nil {
		s.logger.Error().Err(err).Msg("创建用户失败")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "server_error",
			Description: "服务器内部错误",
		})
		return
	}

	// 分配普通用户角色
	err = s.repo.AssignRoleToUser(r.Context(), user.ID, "user")
	if err != nil {
		s.logger.Warn().Err(err).Str("userID", user.ID).Msg("分配角色失败")
	}

	// 记录用户注册活动
	activity := &models.UserActivity{
		UserID:      user.ID,
		ActionType:  "user_register",
		IPAddress:   r.RemoteAddr,
		UserAgent:   r.UserAgent(),
		Description: "用户注册",
	}

	if err := s.repo.LogUserActivity(r.Context(), activity); err != nil {
		s.logger.Warn().Err(err).Str("userID", user.ID).Msg("记录活动失败")
	}

	// 生成JWT令牌
	tokenExpiry := 24 * time.Hour       // 24小时
	refreshExpiry := 7 * 24 * time.Hour // 7天

	// 准备JWT声明
	claims := &auth.Claims{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		Roles:    []string{"user"},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "flick-api",
		},
	}

	// 获取配置
	cfg := config.GetConfig()

	// 签名JWT
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		s.logger.Error().Err(err).Msg("生成访问令牌失败")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "server_error",
			Description: "服务器内部错误",
		})
		return
	}

	// 准备刷新令牌声明
	refreshClaims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshExpiry)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Subject:   user.ID,
		Issuer:    "flick-api",
	}

	// 签名刷新令牌
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(cfg.JWTSecret + "-refresh"))
	if err != nil {
		s.logger.Error().Err(err).Msg("生成刷新令牌失败")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "server_error",
			Description: "服务器内部错误",
		})
		return
	}

	// 创建会话
	session := &models.UserSession{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		IPAddress:    r.RemoteAddr,
		UserAgent:    r.UserAgent(),
		Device:       extractDeviceInfo(r.UserAgent()),
		LastActivity: time.Now(),
		ExpiresAt:    time.Now().Add(refreshExpiry),
	}

	err = s.repo.CreateSession(r.Context(), session)
	if err != nil {
		s.logger.Warn().Err(err).Str("userID", user.ID).Msg("创建会话失败")
	}

	// 返回响应
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(tokenExpiry.Seconds()),
	})
}

// LoginHandler 处理用户登录请求
func (s *UserServiceImpl) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// 设置内容类型
	w.Header().Set("Content-Type", "application/json")

	// 解析请求体
	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		s.logger.Error().Err(err).Msg("无法解析登录请求")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "invalid_request",
			Description: "无法解析请求",
		})
		return
	}

	// 验证请求
	if req.UsernameOrEmail == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "invalid_request",
			Description: "用户名/邮箱和密码为必填项",
		})
		return
	}

	// 根据用户名或邮箱查找用户
	var user *models.User

	// 尝试通过邮箱查找
	if strings.Contains(req.UsernameOrEmail, "@") {
		user, err = s.repo.GetUserByEmail(r.Context(), req.UsernameOrEmail)
	} else {
		// 尝试通过用户名查找
		user, err = s.repo.GetUserByUsername(r.Context(), req.UsernameOrEmail)
	}

	if err != nil || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "invalid_credentials",
			Description: "用户名/邮箱或密码错误",
		})
		return
	}

	// 验证密码
	if !user.CheckPassword(req.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "invalid_credentials",
			Description: "用户名/邮箱或密码错误",
		})
		return
	}

	// 检查账户状态
	if user.AccountStatus != "active" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "account_" + user.AccountStatus,
			Description: "账户" + user.AccountStatus,
		})
		return
	}

	// 获取用户角色
	roles, err := s.repo.GetUserRoles(r.Context(), user.ID)
	if err != nil {
		s.logger.Warn().Err(err).Str("userID", user.ID).Msg("获取用户角色失败")
		roles = []string{"user"} // 默认角色
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLogin = &now
	user.LastIPAddress = r.RemoteAddr
	user.UserAgent = r.UserAgent()

	err = s.repo.UpdateUser(r.Context(), user)
	if err != nil {
		s.logger.Warn().Err(err).Str("userID", user.ID).Msg("更新用户登录信息失败")
	}

	// 记录登录活动
	activity := &models.UserActivity{
		UserID:      user.ID,
		ActionType:  "user_login",
		IPAddress:   r.RemoteAddr,
		UserAgent:   r.UserAgent(),
		Description: "用户登录",
	}

	if err := s.repo.LogUserActivity(r.Context(), activity); err != nil {
		s.logger.Warn().Err(err).Str("userID", user.ID).Msg("记录活动失败")
	}

	// 生成JWT令牌
	tokenExpiry := 24 * time.Hour       // 24小时
	refreshExpiry := 7 * 24 * time.Hour // 7天

	// 准备JWT声明
	claims := &auth.Claims{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "flick-api",
		},
	}

	// 获取配置
	cfg := config.GetConfig()

	// 签名JWT
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		s.logger.Error().Err(err).Msg("生成访问令牌失败")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "server_error",
			Description: "服务器内部错误",
		})
		return
	}

	// 准备刷新令牌声明
	refreshClaims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshExpiry)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Subject:   user.ID,
		Issuer:    "flick-api",
	}

	// 签名刷新令牌
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(cfg.JWTSecret + "-refresh"))
	if err != nil {
		s.logger.Error().Err(err).Msg("生成刷新令牌失败")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "server_error",
			Description: "服务器内部错误",
		})
		return
	}

	// 创建会话
	session := &models.UserSession{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		IPAddress:    r.RemoteAddr,
		UserAgent:    r.UserAgent(),
		Device:       extractDeviceInfo(r.UserAgent()),
		LastActivity: time.Now(),
		ExpiresAt:    time.Now().Add(refreshExpiry),
	}

	err = s.repo.CreateSession(r.Context(), session)
	if err != nil {
		s.logger.Warn().Err(err).Str("userID", user.ID).Msg("创建会话失败")
	}

	// 返回响应
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(tokenExpiry.Seconds()),
	})
}

// LogoutHandler 处理用户登出请求
func (s *UserServiceImpl) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// 设置内容类型
	w.Header().Set("Content-Type", "application/json")

	// 从Authorization头获取令牌
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "unauthorized",
			Description: "未提供认证令牌",
		})
		return
	}

	tokenStr := authHeader[7:] // 去掉"Bearer "前缀

	// 解析令牌
	cfg := config.GetConfig()
	token, err := jwt.ParseWithClaims(tokenStr, &auth.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "invalid_token",
			Description: "认证令牌无效",
		})
		return
	}

	claims, ok := token.Claims.(*auth.Claims)
	if !ok || !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "invalid_token",
			Description: "认证令牌无效",
		})
		return
	}

	// 获取会话ID
	sessionID := r.URL.Query().Get("session_id")

	// 如果提供了会话ID，则删除特定会话
	if sessionID != "" {
		session, err := s.repo.GetSession(r.Context(), sessionID)
		if err != nil || session == nil || session.UserID != claims.UserID {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "session_not_found",
				Description: "未找到会话",
			})
			return
		}

		err = s.repo.DeleteSession(r.Context(), sessionID)
		if err != nil {
			s.logger.Error().Err(err).Str("sessionID", sessionID).Msg("删除会话失败")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "server_error",
				Description: "服务器内部错误",
			})
			return
		}
	} else {
		// 否则，删除所有会话（全部登出）
		err = s.repo.DeleteUserSessions(r.Context(), claims.UserID)
		if err != nil {
			s.logger.Error().Err(err).Str("userID", claims.UserID).Msg("删除用户会话失败")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "server_error",
				Description: "服务器内部错误",
			})
			return
		}
	}

	// 记录登出活动
	activity := &models.UserActivity{
		UserID:      claims.UserID,
		ActionType:  "user_logout",
		IPAddress:   r.RemoteAddr,
		UserAgent:   r.UserAgent(),
		Description: "用户登出",
	}

	if err := s.repo.LogUserActivity(r.Context(), activity); err != nil {
		s.logger.Warn().Err(err).Str("userID", claims.UserID).Msg("记录活动失败")
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "已成功登出",
	})
}

// RefreshTokenHandler 处理刷新令牌请求
func (s *UserServiceImpl) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	// 设置内容类型
	w.Header().Set("Content-Type", "application/json")

	// 解析请求体
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		s.logger.Error().Err(err).Msg("无法解析刷新令牌请求")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "invalid_request",
			Description: "无法解析请求",
		})
		return
	}

	if req.RefreshToken == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "invalid_request",
			Description: "刷新令牌是必需的",
		})
		return
	}

	// 解析刷新令牌
	cfg := config.GetConfig()
	token, err := jwt.ParseWithClaims(req.RefreshToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret + "-refresh"), nil
	})

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "invalid_token",
			Description: "刷新令牌无效",
		})
		return
	}

	refreshClaims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "invalid_token",
			Description: "刷新令牌无效",
		})
		return
	}

	// 获取用户ID
	userID := refreshClaims.Subject

	// 获取用户
	user, err := s.repo.GetUserByID(r.Context(), userID)
	if err != nil || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "invalid_token",
			Description: "刷新令牌无效",
		})
		return
	}

	// 检查账户状态
	if user.AccountStatus != "active" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "account_" + user.AccountStatus,
			Description: "账户" + user.AccountStatus,
		})
		return
	}

	// 获取用户角色
	roles, err := s.repo.GetUserRoles(r.Context(), user.ID)
	if err != nil {
		s.logger.Warn().Err(err).Str("userID", user.ID).Msg("获取用户角色失败")
		roles = []string{"user"} // 默认角色
	}

	// 生成新令牌
	tokenExpiry := 24 * time.Hour       // 24小时
	refreshExpiry := 7 * 24 * time.Hour // 7天

	// 准备JWT声明
	claims := &auth.Claims{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "flick-api",
		},
	}

	// 签名JWT
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		s.logger.Error().Err(err).Msg("生成访问令牌失败")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "server_error",
			Description: "服务器内部错误",
		})
		return
	}

	// 准备新的刷新令牌声明
	newRefreshClaims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshExpiry)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Subject:   user.ID,
		Issuer:    "flick-api",
	}

	// 签名新的刷新令牌
	newRefreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, newRefreshClaims).SignedString([]byte(cfg.JWTSecret + "-refresh"))
	if err != nil {
		s.logger.Error().Err(err).Msg("生成刷新令牌失败")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "server_error",
			Description: "服务器内部错误",
		})
		return
	}

	// 更新会话（查找与旧刷新令牌相关的会话）
	sessions, err := s.repo.GetActiveSessionsForUser(r.Context(), user.ID)
	if err == nil {
		for _, session := range sessions {
			if session.RefreshToken == req.RefreshToken {
				// 更新会话
				session.RefreshToken = newRefreshToken
				session.LastActivity = time.Now()
				session.ExpiresAt = time.Now().Add(refreshExpiry)

				err = s.repo.UpdateSession(r.Context(), session)
				if err != nil {
					s.logger.Warn().Err(err).Str("sessionID", session.ID).Msg("更新会话失败")
				}
				break
			}
		}
	}

	// 记录活动
	activity := &models.UserActivity{
		UserID:      user.ID,
		ActionType:  "token_refresh",
		IPAddress:   r.RemoteAddr,
		UserAgent:   r.UserAgent(),
		Description: "令牌刷新",
	}

	if err := s.repo.LogUserActivity(r.Context(), activity); err != nil {
		s.logger.Warn().Err(err).Str("userID", user.ID).Msg("记录活动失败")
	}

	// 返回新令牌
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
		"expires_in":    int(tokenExpiry.Seconds()),
		"token_type":    "Bearer",
	})
}

// 辅助函数：从User-Agent提取设备信息
func extractDeviceInfo(userAgent string) string {
	ua := strings.ToLower(userAgent)

	// 检测设备类型
	device := "unknown"

	if strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") || strings.Contains(ua, "ipod") {
		device = "iOS"
	} else if strings.Contains(ua, "android") {
		device = "Android"
	} else if strings.Contains(ua, "windows") {
		device = "Windows"
	} else if strings.Contains(ua, "macintosh") || strings.Contains(ua, "mac os") {
		device = "macOS"
	} else if strings.Contains(ua, "linux") {
		device = "Linux"
	}

	// 检测浏览器
	browser := "unknown"

	if strings.Contains(ua, "chrome") && !strings.Contains(ua, "chromium") {
		browser = "Chrome"
	} else if strings.Contains(ua, "firefox") {
		browser = "Firefox"
	} else if strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome") {
		browser = "Safari"
	} else if strings.Contains(ua, "edge") {
		browser = "Edge"
	} else if strings.Contains(ua, "opera") || strings.Contains(ua, "opr") {
		browser = "Opera"
	}

	return fmt.Sprintf("%s / %s", device, browser)
}

// GetFollowStatsHandler 处理获取用户关注统计信息请求
func (s *UserServiceImpl) GetFollowStatsHandler(w http.ResponseWriter, r *http.Request) {
	// 设置内容类型
	w.Header().Set("Content-Type", "application/json")

	// 从请求中获取用户ID
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		// 尝试从Authorization头部获取
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "unauthorized",
				Description: "未提供认证令牌",
			})
			return
		}

		tokenStr := authHeader[7:] // 去掉"Bearer "前缀

		// 解析令牌
		cfg := config.GetConfig()
		token, err := jwt.ParseWithClaims(tokenStr, &auth.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "invalid_token",
				Description: "认证令牌无效",
			})
			return
		}

		claims, ok := token.Claims.(*auth.Claims)
		if !ok || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "invalid_token",
				Description: "认证令牌无效",
			})
			return
		}

		userID = claims.UserID
	}

	// 确保有用户ID
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "unauthorized",
			Description: "无法识别用户",
		})
		return
	}

	// 获取关注统计信息
	followers, following, err := s.repo.GetFollowStats(r.Context(), userID)
	if err != nil {
		s.logger.Error().Err(err).Str("userID", userID).Msg("获取关注统计信息失败")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "server_error",
			Description: "服务器内部错误",
		})
		return
	}

	// 返回关注统计信息
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		Followers int64 `json:"followers"`
		Following int64 `json:"following"`
	}{
		Followers: followers,
		Following: following,
	})
}

// UploadAvatarHandler 处理上传用户头像
func (s *UserServiceImpl) UploadAvatarHandler(w http.ResponseWriter, r *http.Request) {
	// 设置内容类型
	w.Header().Set("Content-Type", "application/json")

	// 从请求中获取用户ID
	userID := r.Header.Get("X-User-ID")
	s.logger.Info().Str("debug", "开始处理头像上传").Str("userID", userID).Str("auth_header", r.Header.Get("Authorization")).Msg("收到头像上传请求")

	if userID == "" {
		// 尝试从Authorization头部获取
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			s.logger.Error().Msg("未提供认证令牌")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "unauthorized",
				Description: "未提供认证令牌",
			})
			return
		}

		tokenStr := authHeader[7:] // 去掉"Bearer "前缀
		s.logger.Info().Str("token_start", tokenStr[:10]+"...").Msg("从Authorization头部获取到令牌")

		// 解析令牌
		cfg := config.GetConfig()
		token, err := jwt.ParseWithClaims(tokenStr, &auth.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil {
			s.logger.Error().Err(err).Msg("解析令牌错误")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "invalid_token",
				Description: "认证令牌无效",
			})
			return
		}

		claims, ok := token.Claims.(*auth.Claims)
		if !ok || !token.Valid {
			s.logger.Error().Msg("无效的令牌声明")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "invalid_token",
				Description: "认证令牌无效",
			})
			return
		}

		userID = claims.UserID
		s.logger.Info().Str("userID", userID).Msg("从令牌中获取到用户ID")
	}

	// 确保有用户ID
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "unauthorized",
			Description: "无法识别用户",
		})
		return
	}

	// 获取当前用户
	user, err := s.repo.GetUserByID(r.Context(), userID)
	if err != nil {
		s.logger.Error().Err(err).Str("userID", userID).Msg("获取用户失败")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "server_error",
			Description: "服务器内部错误",
		})
		return
	}

	// 最大文件大小 (5MB)
	maxSize := int64(5 * 1024 * 1024)
	err = r.ParseMultipartForm(maxSize)
	if err != nil {
		s.logger.Error().Err(err).Msg("解析文件失败")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "invalid_request",
			Description: "无法解析上传文件",
		})
		return
	}

	// 获取上传的文件
	file, handler, err := r.FormFile("avatar")
	if err != nil {
		s.logger.Error().Err(err).Msg("获取上传文件失败")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "invalid_request",
			Description: "获取上传文件失败",
		})
		return
	}
	defer file.Close()

	// 检查文件类型
	contentType := handler.Header.Get("Content-Type")
	s.logger.Info().Str("content_type", contentType).Str("filename", handler.Filename).Int64("size", handler.Size).Msg("上传文件信息")

	// 1. 检查Content-Type
	isImage := false
	if strings.HasPrefix(contentType, "image/") {
		isImage = true
	}

	// 2. 检查文件扩展名作为备用验证
	fileName := handler.Filename
	fileExt := strings.ToLower(filepath.Ext(fileName))

	// 允许的图片扩展名列表
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp"}
	extAllowed := false

	for _, ext := range allowedExts {
		if fileExt == ext {
			extAllowed = true
			break
		}
	}

	// 只有当Content-Type存在且不是图片类型，并且文件扩展名也不是允许的类型时才拒绝
	if contentType != "" && !isImage && !extAllowed {
		s.logger.Warn().Str("content_type", contentType).Str("file_ext", fileExt).Msg("文件类型不是图片格式")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "invalid_request",
			Description: "只允许上传图片文件",
		})
		return
	}

	// 创建MinIO客户端
	cfg := config.GetConfig()
	minioConfig := storage.MinioConfig{
		Endpoint:       cfg.MinioEndpoint,
		AccessKey:      cfg.MinioAccessKey,
		SecretKey:      cfg.MinioSecretKey,
		BucketName:     cfg.MinioBucketName,
		UseSSL:         cfg.MinioUseSSL,
		PublicEndpoint: cfg.MinioPublicEndpoint,
	}

	minioClient, err := storage.NewMinioClient(minioConfig)
	if err != nil {
		s.logger.Error().Err(err).Msg("创建MinIO客户端失败")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "server_error",
			Description: "存储服务连接失败",
		})
		return
	}

	// 将文件上传到MinIO
	fileSize := handler.Size
	avatarURL, err := minioClient.UploadFile(r.Context(), userID, "avatar", fileSize, file, handler.Filename)
	if err != nil {
		s.logger.Error().Err(err).Msg("上传文件到MinIO失败")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "server_error",
			Description: "文件上传失败",
		})
		return
	}

	s.logger.Info().Str("userID", userID).Str("url", avatarURL).Msg("头像上传成功")

	// 如果用户之前有头像，尝试删除旧文件
	if user.AvatarURL != "" && !strings.HasPrefix(user.AvatarURL, "/uploads/") {
		oldObjectName := storage.ExtractObjectNameFromURL(user.AvatarURL, "user-media")
		if oldObjectName != "" {
			if err := minioClient.DeleteFile(r.Context(), oldObjectName); err != nil {
				s.logger.Warn().Err(err).Str("objectName", oldObjectName).Msg("删除旧头像失败")
			}
		}
	}

	// 更新用户头像URL
	user.AvatarURL = avatarURL
	err = s.repo.UpdateUser(r.Context(), user)
	if err != nil {
		s.logger.Error().Err(err).Str("userID", userID).Msg("更新用户头像失败")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "server_error",
			Description: "更新用户头像失败",
		})
		return
	}

	// 记录活动
	activity := &models.UserActivity{
		UserID:      user.ID,
		ActionType:  "avatar_update",
		IPAddress:   r.RemoteAddr,
		UserAgent:   r.UserAgent(),
		Description: "更新用户头像",
	}

	if err := s.repo.LogUserActivity(r.Context(), activity); err != nil {
		s.logger.Warn().Err(err).Str("userID", user.ID).Msg("记录活动失败")
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		Success   bool   `json:"success"`
		AvatarURL string `json:"avatar_url"`
	}{
		Success:   true,
		AvatarURL: avatarURL,
	})
}

// UploadCoverImageHandler 处理上传用户封面图片
func (s *UserServiceImpl) UploadCoverImageHandler(w http.ResponseWriter, r *http.Request) {
	// 设置内容类型
	w.Header().Set("Content-Type", "application/json")

	// 从请求中获取用户ID
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		// 尝试从Authorization头部获取
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "unauthorized",
				Description: "未提供认证令牌",
			})
			return
		}

		tokenStr := authHeader[7:] // 去掉"Bearer "前缀

		// 解析令牌
		cfg := config.GetConfig()
		token, err := jwt.ParseWithClaims(tokenStr, &auth.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "invalid_token",
				Description: "认证令牌无效",
			})
			return
		}

		claims, ok := token.Claims.(*auth.Claims)
		if !ok || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "invalid_token",
				Description: "认证令牌无效",
			})
			return
		}

		userID = claims.UserID
	}

	// 确保有用户ID
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "unauthorized",
			Description: "无法识别用户",
		})
		return
	}

	// 获取当前用户
	user, err := s.repo.GetUserByID(r.Context(), userID)
	if err != nil {
		s.logger.Error().Err(err).Str("userID", userID).Msg("获取用户失败")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "server_error",
			Description: "服务器内部错误",
		})
		return
	}

	// 最大文件大小 (5MB)
	maxSize := int64(5 * 1024 * 1024)
	err = r.ParseMultipartForm(maxSize)
	if err != nil {
		s.logger.Error().Err(err).Msg("解析文件失败")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "invalid_request",
			Description: "无法解析上传文件",
		})
		return
	}

	// 获取上传的文件
	file, handler, err := r.FormFile("cover_image")
	if err != nil {
		s.logger.Error().Err(err).Msg("获取上传文件失败")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "invalid_request",
			Description: "获取上传文件失败",
		})
		return
	}
	defer file.Close()

	// 检查文件类型
	contentType := handler.Header.Get("Content-Type")
	s.logger.Info().Str("content_type", contentType).Str("filename", handler.Filename).Int64("size", handler.Size).Msg("上传文件信息")

	// 1. 检查Content-Type
	isImage := false
	if strings.HasPrefix(contentType, "image/") {
		isImage = true
	}

	// 2. 检查文件扩展名作为备用验证
	fileName := handler.Filename
	fileExt := strings.ToLower(filepath.Ext(fileName))

	// 允许的图片扩展名列表
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp"}
	extAllowed := false

	for _, ext := range allowedExts {
		if fileExt == ext {
			extAllowed = true
			break
		}
	}

	// 只有当Content-Type存在且不是图片类型，并且文件扩展名也不是允许的类型时才拒绝
	if contentType != "" && !isImage && !extAllowed {
		s.logger.Warn().Str("content_type", contentType).Str("file_ext", fileExt).Msg("文件类型不是图片格式")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "invalid_request",
			Description: "只允许上传图片文件",
		})
		return
	}

	// 创建MinIO客户端
	cfg := config.GetConfig()
	minioConfig := storage.MinioConfig{
		Endpoint:       cfg.MinioEndpoint,
		AccessKey:      cfg.MinioAccessKey,
		SecretKey:      cfg.MinioSecretKey,
		BucketName:     cfg.MinioBucketName,
		UseSSL:         cfg.MinioUseSSL,
		PublicEndpoint: cfg.MinioPublicEndpoint,
	}

	minioClient, err := storage.NewMinioClient(minioConfig)
	if err != nil {
		s.logger.Error().Err(err).Msg("创建MinIO客户端失败")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "server_error",
			Description: "存储服务连接失败",
		})
		return
	}

	// 将文件上传到MinIO
	fileSize := handler.Size
	coverImageURL, err := minioClient.UploadFile(r.Context(), userID, "cover-image", fileSize, file, handler.Filename)
	if err != nil {
		s.logger.Error().Err(err).Msg("上传文件到MinIO失败")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "server_error",
			Description: "文件上传失败",
		})
		return
	}

	s.logger.Info().Str("userID", userID).Str("url", coverImageURL).Msg("封面图片上传成功")

	// 如果用户之前有封面图片，尝试删除旧文件
	if user.CoverImageURL != "" && !strings.HasPrefix(user.CoverImageURL, "/uploads/") {
		oldObjectName := storage.ExtractObjectNameFromURL(user.CoverImageURL, "user-media")
		if oldObjectName != "" {
			if err := minioClient.DeleteFile(r.Context(), oldObjectName); err != nil {
				s.logger.Warn().Err(err).Str("objectName", oldObjectName).Msg("删除旧封面图片失败")
			}
		}
	}

	// 更新用户封面图片URL
	user.CoverImageURL = coverImageURL
	err = s.repo.UpdateUser(r.Context(), user)
	if err != nil {
		s.logger.Error().Err(err).Str("userID", userID).Msg("更新用户封面图片失败")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "server_error",
			Description: "更新用户封面图片失败",
		})
		return
	}

	// 记录活动
	activity := &models.UserActivity{
		UserID:      user.ID,
		ActionType:  "cover_image_update",
		IPAddress:   r.RemoteAddr,
		UserAgent:   r.UserAgent(),
		Description: "更新用户封面图片",
	}

	if err := s.repo.LogUserActivity(r.Context(), activity); err != nil {
		s.logger.Warn().Err(err).Str("userID", user.ID).Msg("记录活动失败")
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		Success       bool   `json:"success"`
		CoverImageURL string `json:"cover_image_url"`
	}{
		Success:       true,
		CoverImageURL: coverImageURL,
	})
}

// FollowUserHandler 处理关注用户请求
func (s *UserServiceImpl) FollowUserHandler(w http.ResponseWriter, r *http.Request) {
	// 设置内容类型
	w.Header().Set("Content-Type", "application/json")

	// 从请求中获取当前用户ID
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		// 尝试从Authorization头部获取
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "unauthorized",
				Description: "未提供认证令牌",
			})
			return
		}

		tokenStr := authHeader[7:] // 去掉"Bearer "前缀

		// 解析令牌
		cfg := config.GetConfig()
		token, err := jwt.ParseWithClaims(tokenStr, &auth.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "invalid_token",
				Description: "认证令牌无效",
			})
			return
		}

		claims, ok := token.Claims.(*auth.Claims)
		if !ok || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "invalid_token",
				Description: "认证令牌无效",
			})
			return
		}

		userID = claims.UserID
	}

	// 确保有用户ID
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "unauthorized",
			Description: "无法识别用户",
		})
		return
	}

	// 获取目标用户ID（被关注的用户）
	vars := mux.Vars(r)
	targetID := vars["id"]
	if targetID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "invalid_request",
			Description: "未提供目标用户ID",
		})
		return
	}

	// 检查是否自己关注自己
	if userID == targetID {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "invalid_request",
			Description: "不能关注自己",
		})
		return
	}

	// 检查目标用户是否存在
	targetUser, err := s.repo.GetUserByID(r.Context(), targetID)
	if err != nil || targetUser == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "not_found",
			Description: "目标用户不存在",
		})
		return
	}

	// 检查是否已经关注
	isFollowing, err := s.repo.IsFollowing(r.Context(), userID, targetID)
	if err != nil {
		s.logger.Error().Err(err).Str("userID", userID).Str("targetID", targetID).Msg("检查关注状态失败")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "server_error",
			Description: "服务器内部错误",
		})
		return
	}

	// 如果已经关注，返回成功但不执行关注操作
	if isFollowing {
		// 获取关注统计
		followers, following, err := s.repo.GetFollowStats(r.Context(), targetID)
		if err != nil {
			s.logger.Error().Err(err).Str("targetID", targetID).Msg("获取关注统计失败")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "server_error",
				Description: "服务器内部错误",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(FollowResponse{
			Success:   true,
			Followers: followers,
			Following: following,
		})
		return
	}

	// 执行关注操作
	err = s.repo.FollowUser(r.Context(), userID, targetID)
	if err != nil {
		s.logger.Error().Err(err).Str("userID", userID).Str("targetID", targetID).Msg("关注用户失败")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "server_error",
			Description: "服务器内部错误",
		})
		return
	}

	// 记录活动
	activity := &models.UserActivity{
		UserID:      userID,
		ActionType:  "follow_user",
		IPAddress:   r.RemoteAddr,
		UserAgent:   r.UserAgent(),
		Description: fmt.Sprintf("关注用户 %s", targetID),
	}

	if err := s.repo.LogUserActivity(r.Context(), activity); err != nil {
		s.logger.Warn().Err(err).Str("userID", userID).Msg("记录活动失败")
	}

	// 获取关注统计
	followers, following, err := s.repo.GetFollowStats(r.Context(), targetID)
	if err != nil {
		s.logger.Error().Err(err).Str("targetID", targetID).Msg("获取关注统计失败")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "server_error",
			Description: "服务器内部错误",
		})
		return
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(FollowResponse{
		Success:   true,
		Followers: followers,
		Following: following,
	})
}

// UnfollowUserHandler 处理取消关注用户请求
func (s *UserServiceImpl) UnfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	// 设置内容类型
	w.Header().Set("Content-Type", "application/json")

	// 从请求中获取当前用户ID
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		// 尝试从Authorization头部获取
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "unauthorized",
				Description: "未提供认证令牌",
			})
			return
		}

		tokenStr := authHeader[7:] // 去掉"Bearer "前缀

		// 解析令牌
		cfg := config.GetConfig()
		token, err := jwt.ParseWithClaims(tokenStr, &auth.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "invalid_token",
				Description: "认证令牌无效",
			})
			return
		}

		claims, ok := token.Claims.(*auth.Claims)
		if !ok || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "invalid_token",
				Description: "认证令牌无效",
			})
			return
		}

		userID = claims.UserID
	}

	// 确保有用户ID
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "unauthorized",
			Description: "无法识别用户",
		})
		return
	}

	// 获取目标用户ID（被取消关注的用户）
	vars := mux.Vars(r)
	targetID := vars["id"]
	if targetID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "invalid_request",
			Description: "未提供目标用户ID",
		})
		return
	}

	// 检查目标用户是否存在
	targetUser, err := s.repo.GetUserByID(r.Context(), targetID)
	if err != nil || targetUser == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "not_found",
			Description: "目标用户不存在",
		})
		return
	}

	// 检查是否已经关注
	isFollowing, err := s.repo.IsFollowing(r.Context(), userID, targetID)
	if err != nil {
		s.logger.Error().Err(err).Str("userID", userID).Str("targetID", targetID).Msg("检查关注状态失败")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "server_error",
			Description: "服务器内部错误",
		})
		return
	}

	// 如果没有关注，返回成功但不执行取消关注操作
	if !isFollowing {
		// 获取关注统计
		followers, following, err := s.repo.GetFollowStats(r.Context(), targetID)
		if err != nil {
			s.logger.Error().Err(err).Str("targetID", targetID).Msg("获取关注统计失败")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "server_error",
				Description: "服务器内部错误",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(FollowResponse{
			Success:   true,
			Followers: followers,
			Following: following,
		})
		return
	}

	// 执行取消关注操作
	err = s.repo.UnfollowUser(r.Context(), userID, targetID)
	if err != nil {
		s.logger.Error().Err(err).Str("userID", userID).Str("targetID", targetID).Msg("取消关注用户失败")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "server_error",
			Description: "服务器内部错误",
		})
		return
	}

	// 记录活动
	activity := &models.UserActivity{
		UserID:      userID,
		ActionType:  "unfollow_user",
		IPAddress:   r.RemoteAddr,
		UserAgent:   r.UserAgent(),
		Description: fmt.Sprintf("取消关注用户 %s", targetID),
	}

	if err := s.repo.LogUserActivity(r.Context(), activity); err != nil {
		s.logger.Warn().Err(err).Str("userID", userID).Msg("记录活动失败")
	}

	// 获取关注统计
	followers, following, err := s.repo.GetFollowStats(r.Context(), targetID)
	if err != nil {
		s.logger.Error().Err(err).Str("targetID", targetID).Msg("获取关注统计失败")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "server_error",
			Description: "服务器内部错误",
		})
		return
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(FollowResponse{
		Success:   true,
		Followers: followers,
		Following: following,
	})
}

// GetUserFollowersHandler 处理获取用户粉丝列表请求
func (s *UserServiceImpl) GetUserFollowersHandler(w http.ResponseWriter, r *http.Request) {
	// 设置内容类型
	w.Header().Set("Content-Type", "application/json")

	// 从请求中获取当前用户ID
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		// 尝试从Authorization头部获取
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "unauthorized",
				Description: "未提供认证令牌",
			})
			return
		}

		tokenStr := authHeader[7:] // 去掉"Bearer "前缀

		// 解析令牌
		cfg := config.GetConfig()
		token, err := jwt.ParseWithClaims(tokenStr, &auth.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "invalid_token",
				Description: "认证令牌无效",
			})
			return
		}

		claims, ok := token.Claims.(*auth.Claims)
		if !ok || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "invalid_token",
				Description: "认证令牌无效",
			})
			return
		}

		userID = claims.UserID
	}

	// 确保有用户ID
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "unauthorized",
			Description: "无法识别用户",
		})
		return
	}

	// 获取目标用户ID
	vars := mux.Vars(r)
	targetID := vars["id"]
	if targetID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "invalid_request",
			Description: "未提供目标用户ID",
		})
		return
	}

	// 检查目标用户是否存在
	targetUser, err := s.repo.GetUserByID(r.Context(), targetID)
	if err != nil || targetUser == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "not_found",
			Description: "目标用户不存在",
		})
		return
	}

	// 获取分页参数
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page := 1
	pageSize := 20

	if pageStr != "" {
		pageVal, err := strconv.Atoi(pageStr)
		if err == nil && pageVal > 0 {
			page = pageVal
		}
	}

	if pageSizeStr != "" {
		pageSizeVal, err := strconv.Atoi(pageSizeStr)
		if err == nil && pageSizeVal > 0 && pageSizeVal <= 100 {
			pageSize = pageSizeVal
		}
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 获取粉丝列表
	followers, total, err := s.repo.GetFollowers(r.Context(), targetID, offset, pageSize)
	if err != nil {
		s.logger.Error().Err(err).Str("targetID", targetID).Msg("获取粉丝列表失败")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "server_error",
			Description: "服务器内部错误",
		})
		return
	}

	// 返回粉丝列表
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		Data    []*models.User `json:"data"`
		Total   int64          `json:"total"`
		Page    int            `json:"page"`
		PerPage int            `json:"per_page"`
	}{
		Data:    followers,
		Total:   total,
		Page:    page,
		PerPage: pageSize,
	})
}

// GetUserFollowingHandler 处理获取用户关注列表请求
func (s *UserServiceImpl) GetUserFollowingHandler(w http.ResponseWriter, r *http.Request) {
	// 设置内容类型
	w.Header().Set("Content-Type", "application/json")

	// 从请求中获取当前用户ID
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		// 尝试从Authorization头部获取
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "unauthorized",
				Description: "未提供认证令牌",
			})
			return
		}

		tokenStr := authHeader[7:] // 去掉"Bearer "前缀

		// 解析令牌
		cfg := config.GetConfig()
		token, err := jwt.ParseWithClaims(tokenStr, &auth.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "invalid_token",
				Description: "认证令牌无效",
			})
			return
		}

		claims, ok := token.Claims.(*auth.Claims)
		if !ok || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "invalid_token",
				Description: "认证令牌无效",
			})
			return
		}

		userID = claims.UserID
	}

	// 确保有用户ID
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "unauthorized",
			Description: "无法识别用户",
		})
		return
	}

	// 获取目标用户ID
	vars := mux.Vars(r)
	targetID := vars["id"]
	if targetID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "invalid_request",
			Description: "未提供目标用户ID",
		})
		return
	}

	// 检查目标用户是否存在
	targetUser, err := s.repo.GetUserByID(r.Context(), targetID)
	if err != nil || targetUser == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "not_found",
			Description: "目标用户不存在",
		})
		return
	}

	// 获取分页参数
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page := 1
	pageSize := 20

	if pageStr != "" {
		pageVal, err := strconv.Atoi(pageStr)
		if err == nil && pageVal > 0 {
			page = pageVal
		}
	}

	if pageSizeStr != "" {
		pageSizeVal, err := strconv.Atoi(pageSizeStr)
		if err == nil && pageSizeVal > 0 && pageSizeVal <= 100 {
			pageSize = pageSizeVal
		}
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 获取关注列表
	following, total, err := s.repo.GetFollowing(r.Context(), targetID, offset, pageSize)
	if err != nil {
		s.logger.Error().Err(err).Str("targetID", targetID).Msg("获取关注列表失败")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:       "server_error",
			Description: "服务器内部错误",
		})
		return
	}

	// 返回关注列表
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		Data    []*models.User `json:"data"`
		Total   int64          `json:"total"`
		Page    int            `json:"page"`
		PerPage int            `json:"per_page"`
	}{
		Data:    following,
		Total:   total,
		Page:    page,
		PerPage: pageSize,
	})
}
