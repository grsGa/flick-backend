package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiterMap 定义限流器映射类型
type RateLimiterMap struct {
	mu       sync.RWMutex
	limiters map[string]*rate.Limiter
}

// NewRateLimiterMap 创建新的限流器映射
func NewRateLimiterMap() *RateLimiterMap {
	return &RateLimiterMap{
		limiters: make(map[string]*rate.Limiter),
	}
}

// GetLimiter 获取限流器，如果不存在则创建
func (r *RateLimiterMap) GetLimiter(key string, rateLimit rate.Limit, burst int) *rate.Limiter {
	r.mu.RLock()
	limiter, exists := r.limiters[key]
	r.mu.RUnlock()

	if !exists {
		r.mu.Lock()
		// 双重检查，避免并发创建
		limiter, exists = r.limiters[key]
		if !exists {
			limiter = rate.NewLimiter(rateLimit, burst)
			r.limiters[key] = limiter
		}
		r.mu.Unlock()
	}

	return limiter
}

// CleanupExpired 清理过期的限流器（可定期调用）
func (r *RateLimiterMap) CleanupExpired(expiry time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 这里可以根据最后访问时间清理
	// 实际项目中可能需要额外维护访问时间
	// 此示例简单起见省略了具体实现
}

// IPRateLimiter 基于IP的请求限流
// rateLimit: 每秒允许的请求数
// burst: 允许的突发请求数
func IPRateLimiter(rateLimit rate.Limit, burst int) gin.HandlerFunc {
	limiters := NewRateLimiterMap()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := limiters.GetLimiter(ip, rateLimit, burst)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "请求过于频繁",
				"message": "请稍后再试",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// UserRateLimiter 基于用户ID的请求限流
// rateLimit: 每秒允许的请求数
// burst: 允许的突发请求数
// userIDFunc: 获取用户ID的函数
func UserRateLimiter(rateLimit rate.Limit, burst int, userIDFunc func(*gin.Context) string) gin.HandlerFunc {
	limiters := NewRateLimiterMap()

	return func(c *gin.Context) {
		userID := userIDFunc(c)
		if userID == "" {
			// 未认证用户使用IP限流
			userID = "ip:" + c.ClientIP()
		}

		limiter := limiters.GetLimiter(userID, rateLimit, burst)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "请求过于频繁",
				"message": "请稍后再试",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// PathRateLimiter 基于路径的请求限流
func PathRateLimiter(pathRates map[string]rate.Limit, defaultRate rate.Limit, burst int) gin.HandlerFunc {
	limiters := make(map[string]*RateLimiterMap)

	// 为每个路径创建一个限流映射
	for path := range pathRates {
		limiters[path] = NewRateLimiterMap()
	}

	// 默认限流映射
	defaultLimiters := NewRateLimiterMap()

	return func(c *gin.Context) {
		path := c.FullPath()
		clientIP := c.ClientIP()

		var limiter *rate.Limiter

		// 检查是否为特定路径设置了限流
		if pathLimiters, ok := limiters[path]; ok {
			limiter = pathLimiters.GetLimiter(clientIP, pathRates[path], burst)
		} else {
			// 使用默认限流
			limiter = defaultLimiters.GetLimiter(clientIP, defaultRate, burst)
		}

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "请求过于频繁",
				"message": "请稍后再试",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
