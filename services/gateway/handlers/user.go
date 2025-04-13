package handlers

import (
	"backend/services/gateway/config"
	
	"github.com/gin-gonic/gin"
)

// UserHandler 处理用户相关的请求
type UserHandler struct {
	BaseHandler
}

// NewUserHandler 创建一个新的用户处理程序
func NewUserHandler(cfg *config.Config) *UserHandler {
	return &UserHandler{
		BaseHandler: BaseHandler{Config: cfg},
	}
}

// Register 处理用户注册请求
func (h *UserHandler) Register(c *gin.Context) {
	h.HandleRequest(c, "user")
}

// Login 处理用户登录请求
func (h *UserHandler) Login(c *gin.Context) {
	h.HandleRequest(c, "user")
}

// RefreshToken 处理令牌刷新请求
func (h *UserHandler) RefreshToken(c *gin.Context) {
	h.HandleRequest(c, "user")
}

// ForgotPassword 处理忘记密码请求
func (h *UserHandler) ForgotPassword(c *gin.Context) {
	h.HandleRequest(c, "user")
}

// ResetPassword 处理重置密码请求
func (h *UserHandler) ResetPassword(c *gin.Context) {
	h.HandleRequest(c, "user")
}

// VerifyEmail 处理邮箱验证请求
func (h *UserHandler) VerifyEmail(c *gin.Context) {
	h.HandleRequest(c, "user")
}

// GetCurrentUser 获取当前用户信息
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	h.HandleRequest(c, "user")
}

// UpdateCurrentUser 更新当前用户信息
func (h *UserHandler) UpdateCurrentUser(c *gin.Context) {
	h.HandleRequest(c, "user")
}

// GetUserByID 获取指定用户信息
func (h *UserHandler) GetUserByID(c *gin.Context) {
	h.HandleRequest(c, "user")
}

// GetUserFollowers 获取用户的关注者
func (h *UserHandler) GetUserFollowers(c *gin.Context) {
	h.HandleRequest(c, "user")
}

// GetUserFollowing 获取用户关注的用户
func (h *UserHandler) GetUserFollowing(c *gin.Context) {
	h.HandleRequest(c, "user")
}

// FollowUser 关注用户
func (h *UserHandler) FollowUser(c *gin.Context) {
	h.HandleRequest(c, "user")
}

// UnfollowUser 取消关注用户
func (h *UserHandler) UnfollowUser(c *gin.Context) {
	h.HandleRequest(c, "user")
}

// BlockUser 屏蔽用户
func (h *UserHandler) BlockUser(c *gin.Context) {
	h.HandleRequest(c, "user")
}

// UnblockUser 取消屏蔽用户
func (h *UserHandler) UnblockUser(c *gin.Context) {
	h.HandleRequest(c, "user")
}

// SearchUsers 搜索用户
func (h *UserHandler) SearchUsers(c *gin.Context) {
	h.HandleRequest(c, "user")
}

// ListUsers 列出所有用户（管理员）
func (h *UserHandler) ListUsers(c *gin.Context) {
	h.HandleRequest(c, "user")
}

// UpdateUserRoles 更新用户角色（管理员）
func (h *UserHandler) UpdateUserRoles(c *gin.Context) {
	h.HandleRequest(c, "user")
}

// DeleteUser 删除用户（管理员）
func (h *UserHandler) DeleteUser(c *gin.Context) {
	h.HandleRequest(c, "user")
} 