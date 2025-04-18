package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenMissing     = errors.New("未提供令牌")
	ErrTokenInvalid     = errors.New("无效的令牌")
	ErrTokenExpired     = errors.New("令牌已过期")
	ErrTokenValidation  = errors.New("令牌验证失败")
	ErrUnexpectedMethod = errors.New("令牌签名方法错误")
)

// JWTConfig JWT中间件配置
type JWTConfig struct {
	// 密钥
	SigningKey []byte
	// 令牌有效期
	TokenExpiration time.Duration
	// 刷新令牌有效期
	RefreshExpiration time.Duration
	// 匿名路径，不需要认证
	AnonymousPaths []string
	// 令牌解析后的回调函数，用于设置用户信息
	ParseTokenCallback func(*gin.Context, map[string]interface{}) error
}

// DefaultJWTConfig 默认JWT配置
func DefaultJWTConfig() *JWTConfig {
	return &JWTConfig{
		SigningKey:        []byte("default_signing_key"),
		TokenExpiration:   time.Hour * 24,
		RefreshExpiration: time.Hour * 24 * 7,
		AnonymousPaths:    []string{},
		ParseTokenCallback: func(c *gin.Context, claims map[string]interface{}) error {
			// 默认不做任何操作
			return nil
		},
	}
}

// TokenClaims 令牌声明
type TokenClaims struct {
	jwt.RegisteredClaims
	UserID   uint     `json:"user_id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
}

// JWTMiddleware 创建JWT认证中间件
func JWTMiddleware(config *JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 允许OPTIONS请求通过
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}
		
		// 检查是否是匿名路径
		path := c.FullPath()
		for _, anonPath := range config.AnonymousPaths {
			if strings.HasPrefix(path, anonPath) {
				c.Next()
				return
			}
		}

		// 获取令牌
		tokenString := extractToken(c)
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "未授权",
				"message": "未提供有效的认证令牌",
			})
			c.Abort()
			return
		}

		// 验证令牌
		token, err := parseToken(tokenString, config.SigningKey)
		if err != nil {
			statusCode := http.StatusUnauthorized
			errorMessage := "认证令牌无效"

			if errors.Is(err, ErrTokenExpired) {
				errorMessage = "认证令牌已过期"
			} else if errors.Is(err, ErrTokenInvalid) {
				errorMessage = "无效的认证令牌"
			}

			c.JSON(statusCode, gin.H{
				"error":   "未授权",
				"message": errorMessage,
			})
			c.Abort()
			return
		}

		// 获取声明
		claims, ok := token.Claims.(*TokenClaims)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "未授权",
				"message": "令牌声明无效",
			})
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("user_roles", claims.Roles)

		// 如果有自定义回调，则执行
		if config.ParseTokenCallback != nil {
			claimsMap := make(map[string]interface{})
			claimsMap["user_id"] = claims.UserID
			claimsMap["username"] = claims.Username
			claimsMap["email"] = claims.Email
			claimsMap["roles"] = claims.Roles

			if err := config.ParseTokenCallback(c, claimsMap); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "内部错误",
					"message": "处理令牌时出错",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// 从请求中提取令牌
func extractToken(c *gin.Context) string {
	// 从授权头获取
	bearerToken := c.Request.Header.Get("Authorization")
	if len(bearerToken) > 7 && strings.ToUpper(bearerToken[0:7]) == "BEARER " {
		return bearerToken[7:]
	}

	// 从查询参数获取
	token := c.Query("token")
	if token != "" {
		return token
	}

	// 从Cookie获取
	tokenCookie, err := c.Cookie("token")
	if err == nil {
		return tokenCookie
	}

	return ""
}

// 解析令牌
func parseToken(tokenString string, signingKey []byte) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %v", ErrUnexpectedMethod, token.Header["alg"])
		}
		return signingKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %v", ErrTokenValidation, err)
	}

	if !token.Valid {
		return nil, ErrTokenInvalid
	}

	return token, nil
}

// GenerateToken 生成JWT令牌
func GenerateToken(userID uint, username, email string, roles []string, config *JWTConfig) (string, error) {
	// 创建令牌声明
	now := time.Now()
	claims := &TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(config.TokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "flick_api",
			Subject:   fmt.Sprintf("%d", userID),
		},
		UserID:   userID,
		Username: username,
		Email:    email,
		Roles:    roles,
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名令牌
	tokenString, err := token.SignedString(config.SigningKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GenerateRefreshToken 生成刷新令牌
func GenerateRefreshToken(userID uint, config *JWTConfig) (string, error) {
	// 创建令牌声明
	now := time.Now()
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(config.RefreshExpiration)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Issuer:    "flick_api",
		Subject:   fmt.Sprintf("%d", userID),
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名令牌
	tokenString, err := token.SignedString(config.SigningKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// RequireAuth 简单验证用户是否已认证
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists || userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "未授权",
				"message": "请先登录",
			})
			c.Abort()
			return
		}
		c.Next()
	}
} 