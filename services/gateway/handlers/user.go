package handlers

import (
	"backend/services/gateway/config"

	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"mime/multipart"
	"net/http"
)

// UserHandler 处理用户相关的请求
type UserHandler struct {
	BaseHandler
	userServiceURL string
	logger         zerolog.Logger
}

// NewUserHandler 创建一个新的用户处理程序
func NewUserHandler(cfg *config.Config) *UserHandler {
	return &UserHandler{
		BaseHandler:    BaseHandler{Config: cfg},
		userServiceURL: cfg.Services["user"].URL,
		logger:         log.With().Str("handler", "user").Logger(),
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
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":       "unauthorized",
			"description": "未提供认证令牌",
		})
		return
	}

	// 获取目标用户ID
	targetID := c.Param("id")
	if targetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "invalid_request",
			"description": "缺少用户ID",
		})
		return
	}

	// 获取分页参数
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "20")

	// 转发到用户服务
	url := fmt.Sprintf("%s/api/v1/users/%s/followers?page=%s&page_size=%s", h.userServiceURL, targetID, page, pageSize)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		h.logger.Error().Err(err).Msg("创建请求失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "server_error",
			"description": "服务器内部错误",
		})
		return
	}

	// 设置用户ID头
	req.Header.Set("X-User-ID", fmt.Sprintf("%v", userID))

	// 转发认证头
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error().Err(err).Msg("发送请求失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "server_error",
			"description": "服务器内部错误",
		})
		return
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error().Err(err).Msg("读取响应失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "server_error",
			"description": "服务器内部错误",
		})
		return
	}

	// 设置响应头
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// 返回响应
	c.Data(resp.StatusCode, "application/json", body)
}

// GetUserFollowing 获取用户关注的用户
func (h *UserHandler) GetUserFollowing(c *gin.Context) {
	h.HandleRequest(c, "user")
}

// FollowUser 关注用户
func (h *UserHandler) FollowUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":       "unauthorized",
			"description": "未提供认证令牌",
		})
		return
	}

	// 获取目标用户ID
	targetID := c.Param("id")
	if targetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "invalid_request",
			"description": "缺少用户ID",
		})
		return
	}

	// 转发到用户服务
	url := fmt.Sprintf("%s/api/v1/users/%s/follow", h.userServiceURL, targetID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		h.logger.Error().Err(err).Msg("创建请求失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "server_error",
			"description": "服务器内部错误",
		})
		return
	}

	// 设置用户ID头
	req.Header.Set("X-User-ID", fmt.Sprintf("%v", userID))

	// 转发认证头
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error().Err(err).Msg("发送请求失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "server_error",
			"description": "服务器内部错误",
		})
		return
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error().Err(err).Msg("读取响应失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "server_error",
			"description": "服务器内部错误",
		})
		return
	}

	// 设置响应头
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// 返回响应
	c.Data(resp.StatusCode, "application/json", body)
}

// UnfollowUser 取消关注用户
func (h *UserHandler) UnfollowUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":       "unauthorized",
			"description": "未提供认证令牌",
		})
		return
	}

	// 获取目标用户ID
	targetID := c.Param("id")
	if targetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "invalid_request",
			"description": "缺少用户ID",
		})
		return
	}

	// 转发到用户服务
	url := fmt.Sprintf("%s/api/v1/users/%s/follow", h.userServiceURL, targetID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		h.logger.Error().Err(err).Msg("创建请求失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "server_error",
			"description": "服务器内部错误",
		})
		return
	}

	// 设置用户ID头
	req.Header.Set("X-User-ID", fmt.Sprintf("%v", userID))

	// 转发认证头
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error().Err(err).Msg("发送请求失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "server_error",
			"description": "服务器内部错误",
		})
		return
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error().Err(err).Msg("读取响应失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "server_error",
			"description": "服务器内部错误",
		})
		return
	}

	// 设置响应头
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// 返回响应
	c.Data(resp.StatusCode, "application/json", body)
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

// GetUserFollowStats 获取用户关注统计信息
func (h *UserHandler) GetUserFollowStats(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":       "unauthorized",
			"description": "未提供认证令牌",
		})
		return
	}

	// 转发到用户服务
	url := fmt.Sprintf("%s/api/v1/users/me/follow-stats", h.userServiceURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		h.logger.Error().Err(err).Msg("创建请求失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "server_error",
			"description": "服务器内部错误",
		})
		return
	}

	// 设置用户ID头
	req.Header.Set("X-User-ID", fmt.Sprintf("%v", userID))

	// 转发认证头
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error().Err(err).Msg("发送请求失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "server_error",
			"description": "服务器内部错误",
		})
		return
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error().Err(err).Msg("读取响应失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "server_error",
			"description": "服务器内部错误",
		})
		return
	}

	// 设置响应头
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// 返回响应
	c.Data(resp.StatusCode, "application/json", body)
}

// UploadAvatar 上传用户头像
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":       "unauthorized",
			"description": "未提供认证令牌",
		})
		return
	}

	// 获取上传的文件
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		h.logger.Error().Err(err).Msg("获取上传文件失败")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "invalid_request",
			"description": "获取上传文件失败",
		})
		return
	}
	defer file.Close()

	// 创建表单
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("avatar", header.Filename)
	if err != nil {
		h.logger.Error().Err(err).Msg("创建表单失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "server_error",
			"description": "服务器内部错误",
		})
		return
	}

	// 复制文件内容
	_, err = io.Copy(part, file)
	if err != nil {
		h.logger.Error().Err(err).Msg("复制文件内容失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "server_error",
			"description": "服务器内部错误",
		})
		return
	}
	writer.Close()

	// 转发到用户服务
	url := fmt.Sprintf("%s/api/v1/users/me/avatar", h.userServiceURL)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		h.logger.Error().Err(err).Msg("创建请求失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "server_error",
			"description": "服务器内部错误",
		})
		return
	}

	// 设置用户ID头
	req.Header.Set("X-User-ID", fmt.Sprintf("%v", userID))

	// 设置内容类型
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 转发认证头
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error().Err(err).Msg("发送请求失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "server_error",
			"description": "服务器内部错误",
		})
		return
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error().Err(err).Msg("读取响应失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "server_error",
			"description": "服务器内部错误",
		})
		return
	}

	// 设置响应头
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// 返回响应
	c.Data(resp.StatusCode, "application/json", respBody)
}

// UploadCoverImage 上传用户封面图片
func (h *UserHandler) UploadCoverImage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":       "unauthorized",
			"description": "未提供认证令牌",
		})
		return
	}

	// 获取上传的文件
	file, header, err := c.Request.FormFile("cover_image")
	if err != nil {
		h.logger.Error().Err(err).Msg("获取上传文件失败")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "invalid_request",
			"description": "获取上传文件失败",
		})
		return
	}
	defer file.Close()

	// 创建表单
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("cover_image", header.Filename)
	if err != nil {
		h.logger.Error().Err(err).Msg("创建表单失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "server_error",
			"description": "服务器内部错误",
		})
		return
	}

	// 复制文件内容
	_, err = io.Copy(part, file)
	if err != nil {
		h.logger.Error().Err(err).Msg("复制文件内容失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "server_error",
			"description": "服务器内部错误",
		})
		return
	}
	writer.Close()

	// 转发到用户服务
	url := fmt.Sprintf("%s/api/v1/users/me/cover-image", h.userServiceURL)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		h.logger.Error().Err(err).Msg("创建请求失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "server_error",
			"description": "服务器内部错误",
		})
		return
	}

	// 设置用户ID头
	req.Header.Set("X-User-ID", fmt.Sprintf("%v", userID))

	// 设置内容类型
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 转发认证头
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error().Err(err).Msg("发送请求失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "server_error",
			"description": "服务器内部错误",
		})
		return
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error().Err(err).Msg("读取响应失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "server_error",
			"description": "服务器内部错误",
		})
		return
	}

	// 设置响应头
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// 返回响应
	c.Data(resp.StatusCode, "application/json", respBody)
}
