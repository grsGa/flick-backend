package auth

import (
	"errors"
	"fmt"
	"time"

	"backend/pkg/config"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// 定义JWT错误
var (
	ErrInvalidToken = errors.New("无效的令牌")
	ErrExpiredToken = errors.New("令牌已过期")
)

// Claims 定义JWT令牌的声明
type Claims struct {
	UserID    string   `json:"user_id"`
	Email     string   `json:"email"`
	Username  string   `json:"username"`
	Roles     []string `json:"roles"`
	jwt.RegisteredClaims
}

// UserKey 上下文键，用于从请求上下文中获取用户信息
type UserKey struct{}

// Service 认证服务，处理JWT相关操作
type Service struct {
	secretKey []byte
	issuer    string
	expiry    time.Duration
}

// NewAuthService 创建新的认证服务
func NewAuthService(secretKey string, issuer string, expiry time.Duration) *Service {
	return &Service{
		secretKey: []byte(secretKey),
		issuer:    issuer,
		expiry:    expiry,
	}
}

// GenerateToken 生成JWT令牌
func (s *Service) GenerateToken(userID string, username string, roles []string) (string, error) {
	// 创建令牌
	token := jwt.New(jwt.SigningMethodHS256)

	// 设置令牌声明
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = userID
	claims["username"] = username
	claims["roles"] = roles
	claims["iss"] = s.issuer
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(s.expiry).Unix()

	// 签名令牌
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken 验证JWT令牌
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非预期的签名方法: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("无效的令牌")
	}

	// 获取声明
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("无效的声明")
	}

	return claims, nil
}

// GenerateRefreshToken 生成刷新令牌
func GenerateRefreshToken(userID string) (string, error) {
	cfg := config.GetConfig()
	log := config.GetLogger()

	// 设置过期时间（刷新令牌有更长的过期时间）
	expirationTime := time.Now().Add(cfg.JWTExpiration * 2)

	// 创建声明
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    "flick-api",
		Subject:   userID,
		ID:        fmt.Sprintf("refresh-%d", time.Now().Unix()),
		Audience:  []string{"flick-users"},
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名令牌
	tokenString, err := token.SignedString([]byte(cfg.JWTSecret + "-refresh"))
	if err != nil {
		log.Error("Failed to generate refresh token", zap.Error(err))
		return "", err
	}

	return tokenString, nil
}

// ValidateRefreshToken 验证刷新令牌
func ValidateRefreshToken(tokenString string) (string, error) {
	cfg := config.GetConfig()
	log := config.GetLogger()

	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.JWTSecret + "-refresh"), nil
	})

	// 处理解析错误
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			log.Debug("Refresh token expired")
			return "", ErrExpiredToken
		}
		log.Debug("Refresh token validation failed", zap.Error(err))
		return "", ErrInvalidToken
	}

	// 获取声明
	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		return claims.Subject, nil
	}

	log.Debug("Invalid refresh token")
	return "", ErrInvalidToken
}

// ValidateToken 包级函数，用于验证JWT令牌
// 这个函数是为了方便中间件调用，它使用配置中的JWT密钥
func ValidateToken(tokenString string) (*Claims, error) {
	cfg := config.GetConfig()
	log := config.GetLogger()
	
	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非预期的签名方法: %v", token.Header["alg"])
		}
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			log.Debug("Token expired")
			return nil, ErrExpiredToken
		}
		log.Debug("Token validation failed", zap.Error(err))
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// 获取声明
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
} 