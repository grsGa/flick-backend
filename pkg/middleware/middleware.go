package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"

	"backend/pkg/auth"
	"backend/pkg/errors"
)

var (
	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_duration_seconds",
		Help: "HTTP请求持续时间",
	}, []string{"path", "method", "status"})

	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "HTTP请求总数",
	}, []string{"path", "method", "status"})
)

// RequestIDKey 是上下文中存储请求ID的键
type RequestIDKey struct{}

// LoggerMiddleware 日志中间件
func LoggerMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// 创建请求ID
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}
			w.Header().Set("X-Request-ID", requestID)

			// 将请求ID添加到上下文
			ctx := context.WithValue(r.Context(), RequestIDKey{}, requestID)
			r = r.WithContext(ctx)

			// 包装响应写入器以捕获状态代码
			wrapper := &responseWriterWrapper{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// 记录请求开始
			logger.Info("开始处理请求",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("request_id", requestID),
			)

			// 处理请求
			next.ServeHTTP(wrapper, r)

			// 记录请求结束
			duration := time.Since(start)
			logger.Info("请求处理完成",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", wrapper.statusCode),
				zap.Duration("duration", duration),
				zap.String("request_id", requestID),
			)
		})
	}
}

// 响应写入器包装器，用于捕获状态代码
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader 实现http.ResponseWriter接口
func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// JWTAuthMiddleware 认证中间件
func JWTAuthMiddleware(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 从Authorization头获取令牌
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "未提供授权令牌", http.StatusUnauthorized)
				return
			}

			// 解析令牌
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "授权格式无效", http.StatusUnauthorized)
				return
			}
			token := parts[1]

			// 验证令牌
			claims, err := authService.ValidateToken(token)
			if err != nil {
				http.Error(w, "无效的令牌", http.StatusUnauthorized)
				return
			}

			// 将用户信息添加到上下文
			ctx := context.WithValue(r.Context(), auth.UserKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// MetricsMiddleware 指标中间件
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 包装响应写入器以捕获状态代码
		wrapper := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// 处理请求
		next.ServeHTTP(wrapper, r)

		// 记录指标
		duration := time.Since(start).Seconds()
		status := fmt.Sprintf("%d", wrapper.statusCode)

		httpDuration.WithLabelValues(r.URL.Path, r.Method, status).Observe(duration)
		httpRequestsTotal.WithLabelValues(r.URL.Path, r.Method, status).Inc()
	})
}

// CORSMiddleware CORS中间件
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 设置CORS头
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// 处理预检请求
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RecoveryMiddleware 恢复中间件
func RecoveryMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// 记录panic
					logger.Error("服务器panic",
						zap.Any("error", err),
						zap.String("path", r.URL.Path),
						zap.String("method", r.Method),
					)

					// 返回500错误
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]string{
						"error": "内部服务器错误",
					})
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// ErrorHandlerMiddleware 错误处理中间件
func ErrorHandlerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 包装响应写入器以捕获错误
		wrapper := &errorResponseWriter{
			ResponseWriter: w,
		}

		next.ServeHTTP(wrapper, r)
	})
}

// 错误响应写入器
type errorResponseWriter struct {
	http.ResponseWriter
}

// RespondWithError 响应错误
func RespondWithError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	var statusCode int
	var responseBody map[string]string

	switch errors.GetErrorCode(err) {
	case errors.ErrCodeNotFound:
		statusCode = http.StatusNotFound
	case errors.ErrCodeInvalidInput:
		statusCode = http.StatusBadRequest
	case errors.ErrCodeUnauthorized:
		statusCode = http.StatusUnauthorized
	case errors.ErrCodeForbidden:
		statusCode = http.StatusForbidden
	case errors.ErrCodeConflict:
		statusCode = http.StatusConflict
	case errors.ErrCodeUnavailable:
		statusCode = http.StatusServiceUnavailable
	case errors.ErrCodeTimeout:
		statusCode = http.StatusGatewayTimeout
	case errors.ErrCodeRateLimited:
		statusCode = http.StatusTooManyRequests
	default:
		statusCode = http.StatusInternalServerError
	}

	responseBody = map[string]string{
		"error":       errors.GetErrorCode(err),
		"description": errors.GetErrorMessage(err),
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(responseBody)
}

// RespondWithJSON 响应JSON
func RespondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
