package service

import (
	"context"
	"encoding/json"
	"time"

	"backend/pkg/config"
	"backend/pkg/httpclient"
	"backend/services/auth/internal/repository"
	"backend/services/auth/proto"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// authService 认证服务实现
type authService struct {
	authRepo          repository.AuthRepository
	cfg               *config.Config
	logger            *zap.Logger
	githubOauthConfig *oauth2.Config
	googleOauthConfig *oauth2.Config
}

// NewAuthService 创建认证服务实例
func NewAuthService(authRepo repository.AuthRepository, cfg *config.Config, logger *zap.Logger) AuthService {
	// Validate that required OAuth configuration is present.
	if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" || cfg.GoogleRedirectURL == "" {
		logger.Fatal("Google OAuth configuration is incomplete. Please check environment variables: GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, GOOGLE_REDIRECT_URL")
	}
	if cfg.GithubClientID == "" || cfg.GithubClientSecret == "" || cfg.GithubRedirectURL == "" {
		logger.Fatal("GitHub OAuth configuration is incomplete. Please check environment variables: GITHUB_CLIENT_ID, GITHUB_CLIENT_SECRET, GITHUB_REDIRECT_URL")
	}

	githubOauthConfig := &oauth2.Config{
		ClientID:     cfg.GithubClientID,
		ClientSecret: cfg.GithubClientSecret,
		RedirectURL:  cfg.GithubRedirectURL,
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}

	googleOauthConfig := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}

	return &authService{
		authRepo:          authRepo,
		cfg:               cfg,
		logger:            logger,
		githubOauthConfig: githubOauthConfig,
		googleOauthConfig: googleOauthConfig,
	}
}

// Login 用户登录
func (s *authService) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	s.logger.Info("Login attempt", zap.String("identifier", req.Identifier))
	// 获取用户信息
	user, err := s.authRepo.GetUserByIdentifier(ctx, req.Identifier)
	if err != nil {
		s.logger.Warn("Login failed: user not found", zap.String("identifier", req.Identifier), zap.Error(err))
		return &proto.LoginResponse{
			Error: &proto.Error{
				Code:    401,
				Message: "Invalid credentials",
			},
		}, nil
	}

	// 验证密码
	if err := s.authRepo.VerifyPassword(ctx, user.Id, req.Password); err != nil {
		s.logger.Warn("Login failed: invalid password", zap.String("userID", user.Id), zap.String("identifier", req.Identifier))
		return &proto.LoginResponse{
			Error: &proto.Error{
				Code:    401,
				Message: "Invalid credentials",
			},
		}, nil
	}

	// 生成JWT令牌
	accessToken, refreshToken, err := s.generateTokens(user.Id)
	if err != nil {
		s.logger.Error("Failed to generate tokens during login", zap.String("userID", user.Id), zap.Error(err))
		return &proto.LoginResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to generate tokens: " + err.Error(),
			},
		}, err
	}

	// 更新用户最后登录时间
	s.authRepo.UpdateUserLoginInfo(ctx, user.Id, time.Now().Format(time.RFC3339))

	s.logger.Info("User logged in successfully", zap.String("userID", user.Id))
	return &proto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1小时
		User:         user,
	}, nil
}

// Register 用户注册
func (s *authService) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	s.logger.Info("Registration attempt", zap.String("email", req.Email), zap.String("username", req.Username))
	// 检查用户是否已存在
	_, err := s.authRepo.GetUserByIdentifier(ctx, req.Email)
	if err == nil {
		s.logger.Warn("Registration failed: user already exists", zap.String("email", req.Email))
		return &proto.RegisterResponse{
			Error: &proto.Error{
				Code:    409,
				Message: "User with this email already exists",
			},
		}, nil
	}

	// 创建用户对象
	user := &proto.User{
		Username:    req.Username,
		Email:       req.Email,
		Phone:       req.Phone,
		DisplayName: req.DisplayName,
		LoginMethod: req.LoginMethod,
		Status:      "active",
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}

	// 保存用户到数据库
	createdUser, err := s.authRepo.CreateUser(ctx, user, req.Password)
	if err != nil {
		s.logger.Error("Failed to create user during registration", zap.String("email", req.Email), zap.Error(err))
		return &proto.RegisterResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to create user: " + err.Error(),
			},
		}, err
	}
	user.Id = createdUser.Id

	// 生成JWT令牌
	accessToken, refreshToken, err := s.generateTokens(user.Id)
	if err != nil {
		s.logger.Error("Failed to generate tokens after registration", zap.String("userID", user.Id), zap.Error(err))
		return &proto.RegisterResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to generate tokens: " + err.Error(),
			},
		}, err
	}

	s.logger.Info("User registered successfully", zap.String("userID", user.Id))
	return &proto.RegisterResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1小时
		User:         user,
	}, nil
}

// ValidateToken 验证令牌
func (s *authService) ValidateToken(ctx context.Context, req *proto.ValidateTokenRequest) (*proto.ValidateTokenResponse, error) {
	// 解析和验证JWT令牌
	claims, err := s.parseToken(req.Token)
	if err != nil {
		return &proto.ValidateTokenResponse{
			Valid: false,
			Error: &proto.Error{
				Code:    401,
				Message: "Invalid token: " + err.Error(),
			},
		}, nil
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return &proto.ValidateTokenResponse{
			Valid: false,
			Error: &proto.Error{
				Code:    401,
				Message: "Invalid token claims",
			},
		}, nil
	}

	return &proto.ValidateTokenResponse{
		Valid:  true,
		UserId: userID,
	}, nil
}

// RefreshToken 刷新令牌
func (s *authService) RefreshToken(ctx context.Context, req *proto.RefreshTokenRequest) (*proto.RefreshTokenResponse, error) {
	// 获取会话信息
	session, err := s.authRepo.GetSession(ctx, req.RefreshToken)
	if err != nil {
		return &proto.RefreshTokenResponse{
			Error: &proto.Error{
				Code:    401,
				Message: "Invalid refresh token",
			},
		}, nil
	}

	// 生成新的JWT令牌
	accessToken, newRefreshToken, err := s.generateTokens(session.UserID)
	if err != nil {
		return &proto.RefreshTokenResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to generate tokens: " + err.Error(),
			},
		}, err
	}

	// 删除旧会话并创建新会话
	s.authRepo.DeleteSession(ctx, req.RefreshToken)
	s.authRepo.CreateSession(ctx, &repository.Session{
		UserID:       session.UserID,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339), // 7天过期
		CreatedAt:    time.Now().Format(time.RFC3339),
	})

	return &proto.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1小时
	}, nil
}

// Logout 登出
func (s *authService) Logout(ctx context.Context, req *proto.LogoutRequest) (*proto.LogoutResponse, error) {
	// 删除用户所有会话
	err := s.authRepo.DeleteUserSessions(ctx, req.UserId)
	if err != nil {
		return &proto.LogoutResponse{
			Success: false,
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to logout: " + err.Error(),
			},
		}, err
	}

	return &proto.LogoutResponse{
		Success: true,
	}, nil
}

// GithubLogin Github登录
func (s *authService) GithubLogin(ctx context.Context, req *proto.GithubLoginRequest) (*proto.GithubLoginResponse, error) {
	s.logger.Info("Initiating GitHub login flow")
	url := s.githubOauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	return &proto.GithubLoginResponse{
		RedirectUrl: url,
	}, nil
}

// GithubCallback Github回调
func (s *authService) GithubCallback(ctx context.Context, req *proto.GithubCallbackRequest) (*proto.GithubCallbackResponse, error) {
	s.logger.Info("Received GitHub callback")

	// Use the configurable HTTP client
	httpClient := httpclient.NewConfigurableClient(s.cfg.CustomCaCertPath)
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)

	// Exchange the code for a token
	token, err := s.githubOauthConfig.Exchange(ctx, req.Code)
	if err != nil {
		s.logger.Error("GitHub OAuth code exchange failed", zap.Error(err))
		return &proto.GithubCallbackResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to exchange code: " + err.Error(),
			},
		}, err
	}

	s.logger.Info("GitHub token exchanged successfully")
	// Get user info from Github
	client := s.githubOauthConfig.Client(ctx, token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		s.logger.Error("Failed to get user info from GitHub", zap.Error(err))
		return &proto.GithubCallbackResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to get user info: " + err.Error(),
			},
		}, err
	}
	defer resp.Body.Close()

	var githubUser struct {
		ID    int    `json:"id"`
		Login string `json:"login"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		s.logger.Error("Failed to decode user info from GitHub", zap.Error(err))
		return &proto.GithubCallbackResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to decode user info: " + err.Error(),
			},
		}, err
	}
	s.logger.Info("GitHub user info decoded", zap.String("githubUser", githubUser.Login), zap.String("email", githubUser.Email))

	// Check if user exists
	user, err := s.authRepo.GetUserByIdentifier(ctx, githubUser.Email)
	if err != nil {
		s.logger.Info("User not found, creating new user from GitHub login", zap.String("email", githubUser.Email))
		// Create new user
		user = &proto.User{
			Username:    githubUser.Login,
			Email:       githubUser.Email,
			DisplayName: githubUser.Name,
			LoginMethod: "github",
			Status:      "active",
			CreatedAt:   time.Now().Format(time.RFC3339),
			UpdatedAt:   time.Now().Format(time.RFC3339),
		}
		createdUser, err := s.authRepo.CreateUser(ctx, user, "")
		if err != nil {
			s.logger.Error("Failed to create user from GitHub login", zap.String("email", githubUser.Email), zap.Error(err))
			return &proto.GithubCallbackResponse{
				Error: &proto.Error{
					Code:    500,
					Message: "Failed to create user: " + err.Error(),
				},
			}, err
		}
		user = createdUser
		s.logger.Info("New user created from GitHub login", zap.String("userID", user.Id))
	} else {
		s.logger.Info("User found for GitHub login", zap.String("userID", user.Id))
	}

	// Generate JWT tokens
	accessToken, refreshToken, err := s.generateTokens(user.Id)
	if err != nil {
		s.logger.Error("Failed to generate tokens for GitHub user", zap.String("userID", user.Id), zap.Error(err))
		return &proto.GithubCallbackResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to generate tokens: " + err.Error(),
			},
		}, err
	}

	s.logger.Info("Tokens generated successfully for GitHub user", zap.String("userID", user.Id))
	return &proto.GithubCallbackResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1小时
		User:         user,
	}, nil
}

// GoogleLogin Google登录
func (s *authService) GoogleLogin(ctx context.Context, req *proto.GoogleLoginRequest) (*proto.GoogleLoginResponse, error) {
	s.logger.Info("Initiating Google login flow")
	url := s.googleOauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	return &proto.GoogleLoginResponse{
		RedirectUrl: url,
	}, nil
}

// GoogleCallback Google回调
func (s *authService) GoogleCallback(ctx context.Context, req *proto.GoogleCallbackRequest) (*proto.GoogleCallbackResponse, error) {
	s.logger.Info("Received Google callback")

	// Use the configurable HTTP client
	httpClient := httpclient.NewConfigurableClient(s.cfg.CustomCaCertPath)
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)

	// Exchange the code for a token
	token, err := s.googleOauthConfig.Exchange(ctx, req.Code)
	if err != nil {
		s.logger.Error("Google OAuth code exchange failed", zap.Error(err))
		return &proto.GoogleCallbackResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to exchange code: " + err.Error(),
			},
		}, err
	}

	s.logger.Info("Google token exchanged successfully")
	// Get user info from Google
	client := s.googleOauthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		s.logger.Error("Failed to get user info from Google", zap.Error(err))
		return &proto.GoogleCallbackResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to get user info: " + err.Error(),
			},
		}, err
	}
	defer resp.Body.Close()

	var googleUser struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		s.logger.Error("Failed to decode user info from Google", zap.Error(err))
		return &proto.GoogleCallbackResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to decode user info: " + err.Error(),
			},
		}, err
	}
	s.logger.Info("Google user info decoded", zap.String("email", googleUser.Email))

	// Check if user exists
	user, err := s.authRepo.GetUserByIdentifier(ctx, googleUser.Email)
	if err != nil {
		s.logger.Info("User not found, creating new user from Google login", zap.String("email", googleUser.Email))
		// Create new user
		user = &proto.User{
			Username:    googleUser.Email,
			Email:       googleUser.Email,
			DisplayName: googleUser.Name,
			LoginMethod: "google",
			Status:      "active",
			CreatedAt:   time.Now().Format(time.RFC3339),
			UpdatedAt:   time.Now().Format(time.RFC3339),
		}
		createdUser, err := s.authRepo.CreateUser(ctx, user, "")
		if err != nil {
			s.logger.Error("Failed to create user from Google login", zap.String("email", googleUser.Email), zap.Error(err))
			return &proto.GoogleCallbackResponse{
				Error: &proto.Error{
					Code:    500,
					Message: "Failed to create user: " + err.Error(),
				},
			}, err
		}
		user = createdUser
		s.logger.Info("New user created from Google login", zap.String("userID", user.Id))
	} else {
		s.logger.Info("User found for Google login", zap.String("userID", user.Id))
	}

	// Generate JWT tokens
	accessToken, refreshToken, err := s.generateTokens(user.Id)
	if err != nil {
		s.logger.Error("Failed to generate tokens for Google user", zap.String("userID", user.Id), zap.Error(err))
		return &proto.GoogleCallbackResponse{
			Error: &proto.Error{
				Code:    500,
				Message: "Failed to generate tokens: " + err.Error(),
			},
		}, err
	}

	s.logger.Info("Tokens generated successfully for Google user", zap.String("userID", user.Id))
	return &proto.GoogleCallbackResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1小时
		User:         user,
	}, nil
}

// generateTokens 生成访问令牌和刷新令牌
func (s *authService) generateTokens(userID string) (accessToken, refreshToken string, err error) {
	// 生成访问令牌
	accessClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour).Unix(), // 1小时过期
		"iat":     time.Now().Unix(),
	}

	accessJwt := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessJwt.SignedString([]byte(s.cfg.JWTSecret)) // 实际项目中应从配置获取
	if err != nil {
		return "", "", err
	}

	// 生成刷新令牌
	refreshClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(), // 7天过期
		"iat":     time.Now().Unix(),
	}

	refreshJwt := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshJwt.SignedString([]byte(s.cfg.JWTSecret)) // 实际项目中应从配置获取
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// parseToken 解析和验证令牌
func (s *authService) parseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWTSecret), nil // 实际项目中应从配置获取
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, err
	}

	return claims, nil
}
