package middleware

import (
	"context"
	"net/http"
	"strings"

	"backend/pkg/auth"
	"backend/pkg/config"

	"go.uber.org/zap"
)

// 上下文键类型，避免与其他包的键冲突
type contextKey string

// UserContextKey 上下文键常量
const (
	UserContextKey = contextKey("user")
)

// AuthMiddleware 基本认证中间件，验证JWT令牌并将用户信息添加到上下文
// 与JWTAuthMiddleware的区别：
// 1. 此版本在没有认证头时不会拒绝请求，而是作为游客继续处理
// 2. 使用config.GetLogger()而不是依赖注入日志器
// 3. 由auth包的ValidateToken函数验证令牌，而不是authService
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := config.GetLogger()

		// 从请求头获取令牌
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// 无认证信息，作为游客处理
			next.ServeHTTP(w, r)
			return
		}

		// 解析令牌
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
			// 认证头格式不正确
			logger.Debug("Invalid authorization header format")
			http.Error(w, "无效的认证格式", http.StatusUnauthorized)
			return
		}

		// 验证令牌
		claims, err := auth.ValidateToken(tokenParts[1])
		if err != nil {
			if err == auth.ErrExpiredToken {
				logger.Debug("Token expired")
				http.Error(w, "令牌已过期", http.StatusUnauthorized)
			} else {
				logger.Debug("Invalid token", zap.Error(err))
				http.Error(w, "无效的令牌", http.StatusUnauthorized)
			}
			return
		}

		// 将用户信息存储在请求上下文中
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth 要求请求必须经过认证
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := config.GetLogger()

		// 获取用户信息
		claims, ok := r.Context().Value(UserContextKey).(*auth.Claims)
		if !ok || claims == nil {
			logger.Debug("Authentication required")
			http.Error(w, "请先登录", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireRoles 要求请求具有特定角色
func RequireRoles(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := config.GetLogger()

			// 获取用户信息
			claims, ok := r.Context().Value(UserContextKey).(*auth.Claims)
			if !ok || claims == nil {
				logger.Debug("Authentication required for role check")
				http.Error(w, "请先登录", http.StatusUnauthorized)
				return
			}

			// 检查用户是否具有所需角色
			hasRole := false
			for _, role := range roles {
				for _, userRole := range claims.Roles {
					if role == userRole {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				logger.Debug("Insufficient role",
					zap.Strings("required", roles),
					zap.Strings("user_roles", claims.Roles),
				)
				http.Error(w, "权限不足", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermission 要求请求具有特定权限
func RequirePermission(rbac *auth.RBAC, permission auth.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := config.GetLogger()

			// 获取用户信息
			claims, ok := r.Context().Value(UserContextKey).(*auth.Claims)
			if !ok || claims == nil {
				logger.Debug("Authentication required for permission check")
				http.Error(w, "请先登录", http.StatusUnauthorized)
				return
			}

			// 检查用户是否具有所需权限
			if !rbac.HasPermission(claims.Roles, permission) {
				logger.Debug("Permission denied",
					zap.String("permission", string(permission)),
					zap.Strings("user_roles", claims.Roles),
				)
				http.Error(w, "权限不足", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetCurrentUser 从请求上下文获取当前用户
func GetCurrentUser(r *http.Request) *auth.Claims {
	if claims, ok := r.Context().Value(UserContextKey).(*auth.Claims); ok {
		return claims
	}
	return nil
}
