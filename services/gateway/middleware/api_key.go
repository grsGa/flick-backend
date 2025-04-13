package middleware

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"
	
	"github.com/gin-gonic/gin"
)

var (
	ErrMissingAPIKey     = errors.New("缺少API密钥")
	ErrInvalidAPIKey     = errors.New("无效的API密钥")
	ErrAPIKeyExpired     = errors.New("API密钥已过期")
	ErrAPIKeyRateExceeded = errors.New("API密钥请求速率超限")
)

// APIKey 表示API密钥及其配置
type APIKey struct {
	Key         string
	Name        string    // API密钥名称，用于身份标识
	Roles       []string  // 密钥关联的角色
	Permissions []string  // 密钥具有的权限
	RateLimit   int       // 每分钟允许的请求次数
	ExpiresAt   time.Time // 过期时间
	CreatedAt   time.Time
}

// APIKeyStore 定义API密钥存储接口
type APIKeyStore interface {
	GetAPIKey(key string) (*APIKey, error)
	ValidateAPIKey(key string) (*APIKey, error)
}

// InMemoryAPIKeyStore 内存实现的API密钥存储
type InMemoryAPIKeyStore struct {
	keys map[string]*APIKey
	mu   sync.RWMutex
}

// NewInMemoryAPIKeyStore 创建内存API密钥存储
func NewInMemoryAPIKeyStore() *InMemoryAPIKeyStore {
	return &InMemoryAPIKeyStore{
		keys: make(map[string]*APIKey),
	}
}

// AddAPIKey 添加API密钥
func (s *InMemoryAPIKeyStore) AddAPIKey(apiKey *APIKey) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.keys[apiKey.Key] = apiKey
}

// GetAPIKey 获取API密钥
func (s *InMemoryAPIKeyStore) GetAPIKey(key string) (*APIKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	apiKey, exists := s.keys[key]
	if !exists {
		return nil, ErrInvalidAPIKey
	}
	
	return apiKey, nil
}

// ValidateAPIKey 验证API密钥
func (s *InMemoryAPIKeyStore) ValidateAPIKey(key string) (*APIKey, error) {
	apiKey, err := s.GetAPIKey(key)
	if err != nil {
		return nil, err
	}
	
	// 检查是否过期
	if !apiKey.ExpiresAt.IsZero() && time.Now().After(apiKey.ExpiresAt) {
		return nil, ErrAPIKeyExpired
	}
	
	return apiKey, nil
}

// APIKeyMiddleware 创建API密钥验证中间件
func APIKeyMiddleware(store APIKeyStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头或查询参数获取API密钥
		apiKey := extractAPIKey(c)
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "未授权",
				"message": ErrMissingAPIKey.Error(),
			})
			c.Abort()
			return
		}
		
		// 验证API密钥
		key, err := store.ValidateAPIKey(apiKey)
		if err != nil {
			statusCode := http.StatusUnauthorized
			errorMessage := err.Error()
			
			switch err {
			case ErrInvalidAPIKey:
				statusCode = http.StatusUnauthorized
			case ErrAPIKeyExpired:
				statusCode = http.StatusForbidden
			case ErrAPIKeyRateExceeded:
				statusCode = http.StatusTooManyRequests
			}
			
			c.JSON(statusCode, gin.H{
				"error":   "API密钥验证失败",
				"message": errorMessage,
			})
			c.Abort()
			return
		}
		
		// 将API密钥信息添加到上下文
		c.Set("api_key", key.Key)
		c.Set("api_key_name", key.Name)
		c.Set("api_key_roles", key.Roles)
		c.Set("api_key_permissions", key.Permissions)
		
		c.Next()
	}
}

// RequireAPIKey 创建验证API密钥并检查权限的中间件
func RequireAPIKey(store APIKeyStore, requiredPermissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头或查询参数获取API密钥
		apiKey := extractAPIKey(c)
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "未授权",
				"message": ErrMissingAPIKey.Error(),
			})
			c.Abort()
			return
		}
		
		// 验证API密钥
		key, err := store.ValidateAPIKey(apiKey)
		if err != nil {
			statusCode := http.StatusUnauthorized
			errorMessage := err.Error()
			
			switch err {
			case ErrInvalidAPIKey:
				statusCode = http.StatusUnauthorized
			case ErrAPIKeyExpired:
				statusCode = http.StatusForbidden
			case ErrAPIKeyRateExceeded:
				statusCode = http.StatusTooManyRequests
			}
			
			c.JSON(statusCode, gin.H{
				"error":   "API密钥验证失败",
				"message": errorMessage,
			})
			c.Abort()
			return
		}
		
		// 检查权限
		if len(requiredPermissions) > 0 {
			hasPermission := false
			
			// 检查API密钥是否具有所需权限
			for _, required := range requiredPermissions {
				for _, permission := range key.Permissions {
					if permission == required || permission == "*" {
						hasPermission = true
						break
					}
				}
				if hasPermission {
					break
				}
			}
			
			if !hasPermission {
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "权限不足",
					"message": "API密钥没有执行此操作的权限",
				})
				c.Abort()
				return
			}
		}
		
		// 将API密钥信息添加到上下文
		c.Set("api_key", key.Key)
		c.Set("api_key_name", key.Name)
		c.Set("api_key_roles", key.Roles)
		c.Set("api_key_permissions", key.Permissions)
		
		c.Next()
	}
}

// 从请求中提取API密钥
func extractAPIKey(c *gin.Context) string {
	// 首先检查Authorization头
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// 支持Bearer和自定义API密钥前缀
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer ")
		} else if strings.HasPrefix(authHeader, "ApiKey ") {
			return strings.TrimPrefix(authHeader, "ApiKey ")
		}
		
		// 如果没有前缀，直接返回
		return authHeader
	}
	
	// 然后检查X-API-Key头
	apiKey := c.GetHeader("X-API-Key")
	if apiKey != "" {
		return apiKey
	}
	
	// 最后检查URL查询参数
	return c.Query("api_key")
}

// ValidateAPIKeyHeader 验证API密钥的中间件，使用恒定时间比较以防止计时攻击
func ValidateAPIKeyHeader(expectedAPIKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := extractAPIKey(c)
		
		// 使用恒定时间比较以防止计时攻击
		if subtle.ConstantTimeCompare([]byte(apiKey), []byte(expectedAPIKey)) != 1 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "未授权",
				"message": "无效的API密钥",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
} 