package auth

import (
	"errors"
	"sync"

	"go.uber.org/zap"
)

// 定义权限相关错误
var (
	ErrRoleNotFound       = errors.New("角色不存在")
	ErrPermissionNotFound = errors.New("权限不存在")
	ErrPermissionDenied   = errors.New("权限被拒绝")
)

// Permission 表示系统中的权限
type Permission string

// 定义系统中的权限常量
const (
	// 用户相关权限
	PermissionUserRead   Permission = "user:read"
	PermissionUserCreate Permission = "user:create"
	PermissionUserUpdate Permission = "user:update"
	PermissionUserDelete Permission = "user:delete"
	PermissionUserAdmin  Permission = "user:admin"

	// 内容相关权限
	PermissionPostRead   Permission = "post:read"
	PermissionPostCreate Permission = "post:create"
	PermissionPostUpdate Permission = "post:update"
	PermissionPostDelete Permission = "post:delete"
	PermissionPostAdmin  Permission = "post:admin"

	// 互动相关权限
	PermissionInteractionRead   Permission = "interaction:read"
	PermissionInteractionCreate Permission = "interaction:create"
	PermissionInteractionDelete Permission = "interaction:delete"
	PermissionInteractionAdmin  Permission = "interaction:admin"

	// 系统管理权限
	PermissionSystemAdmin Permission = "system:admin"
)

// RBAC 实现基于角色的访问控制
type RBAC struct {
	rolePermissions map[string]map[Permission]bool
	mutex           sync.RWMutex
	logger          *zap.Logger
}

// NewRBAC 创建RBAC实例
func NewRBAC(logger *zap.Logger) *RBAC {
	rbac := &RBAC{
		rolePermissions: make(map[string]map[Permission]bool),
		logger:          logger,
	}

	// 初始化默认角色和权限
	rbac.initDefaultRoles()

	return rbac
}

// initDefaultRoles 初始化默认角色和权限
func (r *RBAC) initDefaultRoles() {
	// 游客角色（未登录用户）
	r.AddRole("guest", []Permission{
		PermissionPostRead,
		PermissionInteractionRead,
	})

	// 普通用户角色
	r.AddRole("user", []Permission{
		PermissionUserRead,
		PermissionUserUpdate, // 只能更新自己的信息
		PermissionPostRead,
		PermissionPostCreate,
		PermissionPostUpdate, // 只能更新自己的帖子
		PermissionPostDelete, // 只能删除自己的帖子
		PermissionInteractionRead,
		PermissionInteractionCreate,
		PermissionInteractionDelete, // 只能删除自己的互动
	})

	// 内容管理员
	r.AddRole("content_moderator", []Permission{
		PermissionUserRead,
		PermissionPostRead,
		PermissionPostUpdate,
		PermissionPostDelete,
		PermissionPostAdmin,
		PermissionInteractionRead,
		PermissionInteractionDelete,
		PermissionInteractionAdmin,
	})

	// 系统管理员
	r.AddRole("admin", []Permission{
		PermissionUserRead,
		PermissionUserCreate,
		PermissionUserUpdate,
		PermissionUserDelete,
		PermissionUserAdmin,
		PermissionPostRead,
		PermissionPostCreate,
		PermissionPostUpdate,
		PermissionPostDelete,
		PermissionPostAdmin,
		PermissionInteractionRead,
		PermissionInteractionCreate,
		PermissionInteractionDelete,
		PermissionInteractionAdmin,
		PermissionSystemAdmin,
	})
}

// AddRole 添加角色及其权限
func (r *RBAC) AddRole(role string, permissions []Permission) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 如果角色不存在，创建它
	if _, exists := r.rolePermissions[role]; !exists {
		r.rolePermissions[role] = make(map[Permission]bool)
	}

	// 分配权限
	for _, perm := range permissions {
		r.rolePermissions[role][perm] = true
	}

	r.logger.Debug("Role added with permissions",
		zap.String("role", role),
		zap.Any("permissions", permissions),
	)
}

// RemoveRole 删除角色
func (r *RBAC) RemoveRole(role string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.rolePermissions[role]; !exists {
		return ErrRoleNotFound
	}

	delete(r.rolePermissions, role)
	r.logger.Debug("Role removed", zap.String("role", role))
	return nil
}

// AddPermissionToRole 为角色添加权限
func (r *RBAC) AddPermissionToRole(role string, permission Permission) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.rolePermissions[role]; !exists {
		return ErrRoleNotFound
	}

	r.rolePermissions[role][permission] = true
	r.logger.Debug("Permission added to role",
		zap.String("role", role),
		zap.String("permission", string(permission)),
	)
	return nil
}

// RemovePermissionFromRole 从角色中移除权限
func (r *RBAC) RemovePermissionFromRole(role string, permission Permission) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.rolePermissions[role]; !exists {
		return ErrRoleNotFound
	}

	if _, exists := r.rolePermissions[role][permission]; !exists {
		return ErrPermissionNotFound
	}

	delete(r.rolePermissions[role], permission)
	r.logger.Debug("Permission removed from role",
		zap.String("role", role),
		zap.String("permission", string(permission)),
	)
	return nil
}

// HasPermission 检查用户是否具有指定权限
func (r *RBAC) HasPermission(roles []string, permission Permission) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, role := range roles {
		if permissions, exists := r.rolePermissions[role]; exists {
			if hasPermission, exists := permissions[permission]; exists && hasPermission {
				return true
			}
		}
	}

	return false
}

// GetRolePermissions 获取角色的所有权限
func (r *RBAC) GetRolePermissions(role string) ([]Permission, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if _, exists := r.rolePermissions[role]; !exists {
		return nil, ErrRoleNotFound
	}

	var permissions []Permission
	for perm, allowed := range r.rolePermissions[role] {
		if allowed {
			permissions = append(permissions, perm)
		}
	}

	return permissions, nil
}

// GetAllRoles 获取所有角色
func (r *RBAC) GetAllRoles() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var roles []string
	for role := range r.rolePermissions {
		roles = append(roles, role)
	}

	return roles
}
