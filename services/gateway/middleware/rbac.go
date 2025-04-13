package middleware

import (
	"errors"
	"net/http"
	"strings"
	
	"github.com/gin-gonic/gin"
)

var (
	ErrUnauthorized      = errors.New("未授权")
	ErrForbidden         = errors.New("权限不足")
	ErrRoleMissing       = errors.New("角色不存在")
	ErrPermissionMissing = errors.New("权限不足")
)

// 定义权限和角色名称类型
type (
	Permission string
	Role       string
)

// RBACConfig RBAC权限配置
type RBACConfig struct {
	// 角色到权限的映射
	RolePermissions map[Role][]Permission
	// 路径到所需权限的映射
	PathPermissions map[string][]Permission
}

// NewRBACConfig 创建新的RBAC配置
func NewRBACConfig() *RBACConfig {
	return &RBACConfig{
		RolePermissions: make(map[Role][]Permission),
		PathPermissions: make(map[string][]Permission),
	}
}

// AddRolePermission 添加角色权限
func (rc *RBACConfig) AddRolePermission(role Role, permissions ...Permission) {
	if _, exists := rc.RolePermissions[role]; !exists {
		rc.RolePermissions[role] = make([]Permission, 0)
	}
	
	// 添加新权限
	for _, permission := range permissions {
		// 检查权限是否已存在
		exists := false
		for _, existingPerm := range rc.RolePermissions[role] {
			if existingPerm == permission {
				exists = true
				break
			}
		}
		
		if !exists {
			rc.RolePermissions[role] = append(rc.RolePermissions[role], permission)
		}
	}
}

// AddPathPermission 添加路径所需权限
func (rc *RBACConfig) AddPathPermission(path string, permissions ...Permission) {
	if _, exists := rc.PathPermissions[path]; !exists {
		rc.PathPermissions[path] = make([]Permission, 0)
	}
	
	// 添加新权限
	for _, permission := range permissions {
		// 检查权限是否已存在
		exists := false
		for _, existingPerm := range rc.PathPermissions[path] {
			if existingPerm == permission {
				exists = true
				break
			}
		}
		
		if !exists {
			rc.PathPermissions[path] = append(rc.PathPermissions[path], permission)
		}
	}
}

// HasPermission 检查角色是否具有权限
func (rc *RBACConfig) HasPermission(role Role, requiredPermission Permission) bool {
	// 检查角色是否存在
	permissions, exists := rc.RolePermissions[role]
	if !exists {
		return false
	}
	
	// 检查是否有*通配符权限
	for _, permission := range permissions {
		if permission == "*" {
			return true
		}
	}
	
	// 检查具体权限
	for _, permission := range permissions {
		if permission == requiredPermission {
			return true
		}
		
		// 支持通配符匹配，如 "user:*" 匹配 "user:read"
		if strings.HasSuffix(string(permission), ":*") {
			prefix := strings.TrimSuffix(string(permission), ":*")
			if strings.HasPrefix(string(requiredPermission), prefix+":") {
				return true
			}
		}
	}
	
	return false
}

// HasAnyPermission 检查角色是否具有任意一个权限
func (rc *RBACConfig) HasAnyPermission(role Role, requiredPermissions []Permission) bool {
	for _, permission := range requiredPermissions {
		if rc.HasPermission(role, permission) {
			return true
		}
	}
	return false
}

// HasAllPermissions 检查角色是否具有所有权限
func (rc *RBACConfig) HasAllPermissions(role Role, requiredPermissions []Permission) bool {
	for _, permission := range requiredPermissions {
		if !rc.HasPermission(role, permission) {
			return false
		}
	}
	return true
}

// RBACMiddleware 创建RBAC权限控制中间件
func RBACMiddleware(config *RBACConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户角色
		// 假设JWT中间件已经验证令牌并将用户角色存储在上下文中
		userRoles, exists := c.Get("user_roles")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "未授权",
				"message": "用户未认证或无角色信息",
			})
			c.Abort()
			return
		}
		
		roles, ok := userRoles.([]string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "内部错误",
				"message": "角色格式错误",
			})
			c.Abort()
			return
		}
		
		// 获取当前路径需要的权限
		path := c.FullPath()
		requiredPermissions, pathExists := config.PathPermissions[path]
		
		// 如果路径不需要特定权限，则允许访问
		if !pathExists || len(requiredPermissions) == 0 {
			c.Next()
			return
		}
		
		// 检查用户是否有所需权限
		hasPermission := false
		for _, role := range roles {
			if config.HasAnyPermission(Role(role), requiredPermissions) {
				hasPermission = true
				break
			}
		}
		
		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "访问被拒绝",
				"message": "您没有访问此资源的权限",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// RequirePermission 创建要求特定权限的中间件
func RequirePermission(permission Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户权限
		userPermissions, exists := c.Get("user_permissions")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "未授权",
				"message": "用户未认证或无权限信息",
			})
			c.Abort()
			return
		}
		
		permissions, ok := userPermissions.([]string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "内部错误",
				"message": "权限格式错误",
			})
			c.Abort()
			return
		}
		
		// 检查用户是否有所需权限
		hasPermission := false
		for _, p := range permissions {
			if p == "*" || p == string(permission) {
				hasPermission = true
				break
			}
			
			// 支持通配符匹配，如 "user:*" 匹配 "user:read"
			if strings.HasSuffix(p, ":*") {
				prefix := strings.TrimSuffix(p, ":*")
				if strings.HasPrefix(string(permission), prefix+":") {
					hasPermission = true
					break
				}
			}
		}
		
		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "访问被拒绝",
				"message": "您没有访问此资源的权限",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// RequireRole 创建要求特定角色的中间件
func RequireRole(role Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户角色
		userRoles, exists := c.Get("user_roles")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "未授权",
				"message": "用户未认证或无角色信息",
			})
			c.Abort()
			return
		}
		
		roles, ok := userRoles.([]string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "内部错误",
				"message": "角色格式错误",
			})
			c.Abort()
			return
		}
		
		// 检查用户是否有所需角色
		hasRole := false
		for _, r := range roles {
			if r == string(role) {
				hasRole = true
				break
			}
		}
		
		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "访问被拒绝",
				"message": "您没有访问此资源的角色权限",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// RequireAnyRole 创建要求任意一个角色的中间件
func RequireAnyRole(roles ...Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户角色
		userRoles, exists := c.Get("user_roles")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "未授权",
				"message": "用户未认证或无角色信息",
			})
			c.Abort()
			return
		}
		
		userRolesSlice, ok := userRoles.([]string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "内部错误",
				"message": "角色格式错误",
			})
			c.Abort()
			return
		}
		
		// 检查用户是否有所需角色中的任意一个
		hasRole := false
		for _, requiredRole := range roles {
			for _, userRole := range userRolesSlice {
				if userRole == string(requiredRole) {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}
		
		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "访问被拒绝",
				"message": "您没有访问此资源的角色权限",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
} 