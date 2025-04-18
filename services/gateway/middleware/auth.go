package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims 表示JWT的声明
type JWTClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// JWTAuth 中间件检查JWT令牌认证
func JWTAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 允许OPTIONS请求通过
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}
		
		token, err := getTokenFromRequest(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权：" + err.Error()})
			c.Abort()
			return
		}

		claims, err := validateToken(token, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的令牌：" + err.Error()})
			c.Abort()
			return
		}

		// 将用户信息保存到上下文中
		c.Set("userID", claims.UserID)
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// getTokenFromRequest 从请求中获取JWT令牌
func getTokenFromRequest(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")

	if authHeader == "" {
		return "", errors.New("授权头部不存在")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("授权格式无效")
	}

	return parts[1], nil
}

// validateToken 验证JWT令牌
func validateToken(tokenString, jwtSecret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		// 检查令牌是否过期
		expirationTime, err := claims.GetExpirationTime()
		if err != nil {
			return nil, errors.New("无法获取过期时间")
		}
		
		if expirationTime.Time.Before(time.Now()) {
			return nil, errors.New("令牌已过期")
		}
		
		return claims, nil
	}

	return nil, errors.New("无效的令牌声明")
}

// GenerateJWT 生成JWT令牌
func GenerateJWT(userID, role, jwtSecret string, expirationHours int) (string, error) {
	expirationTime := time.Now().Add(time.Hour * time.Duration(expirationHours))
	
	claims := JWTClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "flick-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(jwtSecret))
}

// RoleAuth 中间件检查用户角色
func RoleAuth(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 允许OPTIONS请求通过
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}
		
		// 从上下文中获取角色（需要先经过JWTAuth中间件）
		role, exists := c.Get("role")

		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权：没有角色信息"})
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误：角色格式无效"})
			c.Abort()
			return
		}

		// 检查用户角色是否在所需角色列表中
		for _, r := range requiredRoles {
			if r == roleStr {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "禁止访问：无足够权限"})
		c.Abort()
	}
}
