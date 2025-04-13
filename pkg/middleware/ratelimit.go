package middleware

import (
	"backend/pkg/config"
	"go.uber.org/zap"
	"net"
	"net/http"
	"sync"
	"time"
)

// RateLimiter 实现请求限流功能
type RateLimiter struct {
	// 每个IP或用户的请求记录
	requests map[string][]time.Time
	// 互斥锁保护map
	mu sync.Mutex
	// 配置
	maxRequests int           // 最大请求数
	interval    time.Duration // 时间窗口
	// 日志
	logger *zap.Logger
}

// NewRateLimiter 创建新的限流器
func NewRateLimiter(maxRequests int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		requests:    make(map[string][]time.Time),
		maxRequests: maxRequests,
		interval:    interval,
		logger:      config.GetLogger(),
	}
}

// 清理过期的请求记录
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, times := range rl.requests {
		var validTimes []time.Time
		for _, t := range times {
			if now.Sub(t) < rl.interval {
				validTimes = append(validTimes, t)
			}
		}
		if len(validTimes) > 0 {
			rl.requests[key] = validTimes
		} else {
			delete(rl.requests, key)
		}
	}
}

// isAllowed 检查请求是否允许
func (rl *RateLimiter) isAllowed(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	times := rl.requests[key]

	// 移除过期记录
	var validTimes []time.Time
	for _, t := range times {
		if now.Sub(t) < rl.interval {
			validTimes = append(validTimes, t)
		}
	}

	// 检查是否超出限制
	if len(validTimes) >= rl.maxRequests {
		rl.requests[key] = validTimes
		return false
	}

	// 添加新请求
	rl.requests[key] = append(validTimes, now)
	return true
}

// IPRateLimit 基于IP地址的限流中间件
func (rl *RateLimiter) IPRateLimit(next http.Handler) http.Handler {
	// 定期清理过期记录
	go func() {
		ticker := time.NewTicker(rl.interval / 2)
		defer ticker.Stop()
		for range ticker.C {
			rl.cleanup()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 获取IP地址
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		// 检查是否允许请求
		if !rl.isAllowed(ip) {
			rl.logger.Debug("Rate limit exceeded", zap.String("ip", ip))
			w.Header().Set("Retry-After", (rl.interval / time.Second).String())
			http.Error(w, "请求过于频繁，请稍后再试", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// UserRateLimit 基于用户ID的限流中间件
func (rl *RateLimiter) UserRateLimit(next http.Handler) http.Handler {
	// 定期清理过期记录
	go func() {
		ticker := time.NewTicker(rl.interval / 2)
		defer ticker.Stop()
		for range ticker.C {
			rl.cleanup()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 获取用户ID
		var userID string
		if claims := GetCurrentUser(r); claims != nil {
			userID = claims.UserID
		} else {
			// 如果用户未登录，使用IP作为标识
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}
			userID = "ip:" + ip
		}

		// 检查是否允许请求
		if !rl.isAllowed(userID) {
			rl.logger.Debug("Rate limit exceeded", zap.String("user_id", userID))
			w.Header().Set("Retry-After", (rl.interval / time.Second).String())
			http.Error(w, "请求过于频繁，请稍后再试", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// PathRateLimit 针对特定路径的限流中间件
func (rl *RateLimiter) PathRateLimit(path string) func(http.Handler) http.Handler {
	// 每个路径创建单独的限流器
	pathLimiter := NewRateLimiter(rl.maxRequests, rl.interval)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 如果路径匹配，使用路径限流器
			if r.URL.Path == path {
				pathLimiter.IPRateLimit(next).ServeHTTP(w, r)
				return
			}

			// 否则直接处理请求
			next.ServeHTTP(w, r)
		})
	}
}
