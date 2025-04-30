package handlers

import (
	"backend/services/gateway/config"

	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
	url := fmt.Sprintf("%s/users/%s/followers?page=%s&page_size=%s", h.userServiceURL, targetID, page, pageSize)
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
	url := fmt.Sprintf("%s/users/%s/follow", h.userServiceURL, targetID)
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
	url := fmt.Sprintf("%s/users/%s/follow", h.userServiceURL, targetID)
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
	url := fmt.Sprintf("%s/users/me/follow-stats", h.userServiceURL)
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
	// 记录请求信息用于调试
	h.logger.Info().Str("path", c.FullPath()).Str("method", c.Request.Method).Msg("收到头像上传请求")

	// 检查 Authorization 头是否存在
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		h.logger.Warn().Msg("请求中没有Authorization头")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":       "unauthorized",
			"description": "未提供认证令牌",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Info().Msg("尝试从userID获取")
		userID, exists = c.Get("userID")
		if !exists {
			h.logger.Error().Msg("认证中间件未设置user_id或userID")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":       "unauthorized",
				"description": "认证失败，未找到用户ID",
			})
			return
		}
	}

	h.logger.Info().Str("userID", fmt.Sprintf("%v", userID)).Str("auth_header", authHeader).Msg("开始处理头像上传")

	// 检查请求内容类型
	contentType := c.GetHeader("Content-Type")
	h.logger.Info().Str("content_type", contentType).Msg("请求Content-Type")

	// 验证内容类型是否为multipart/form-data
	if !strings.Contains(contentType, "multipart/form-data") {
		h.logger.Error().Str("content_type", contentType).Msg("请求内容类型不是multipart/form-data")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "invalid_content_type",
			"description": "请求必须是multipart/form-data类型",
		})
		return
	}

	// 先解析multipart表单，这一步很重要，修复EOF错误
	err := c.Request.ParseMultipartForm(32 << 20) // 32MB max memory
	if err != nil {
		h.logger.Error().Err(err).Msg("解析multipart表单失败")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "invalid_request",
			"description": "解析multipart表单失败: " + err.Error(),
		})
		return
	}

	// 记录表单字段
	if c.Request.MultipartForm != nil && c.Request.MultipartForm.File != nil {
		for field := range c.Request.MultipartForm.File {
			h.logger.Info().Str("field", field).Msg("发现表单字段")
		}
	}

	// 获取上传的文件
	var file multipart.File
	var header *multipart.FileHeader

	// 依次尝试可能的字段名
	possibleFields := []string{"avatar", "file", "image", "upload", "userAvatar"}
	var lastErr error
	var foundField string

	for _, field := range possibleFields {
		file, header, err = c.Request.FormFile(field)
		if err == nil {
			foundField = field
			h.logger.Info().Str("field", field).Msg("成功从字段获取文件")
			break
		}
		lastErr = err
		h.logger.Debug().Err(err).Str("field", field).Msg("尝试获取字段失败")
	}

	if foundField == "" {
		h.logger.Error().Err(lastErr).Msg("所有可能的字段名尝试均失败")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "invalid_request",
			"description": "获取上传文件失败，请确保表单字段名正确",
		})
		return
	}
	defer file.Close()

	h.logger.Info().Str("filename", header.Filename).Int64("size", header.Size).Str("field", foundField).Msg("成功获取上传文件")

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

	// 先将文件内容读取到内存中
	fileContent, err := io.ReadAll(file)
	if err != nil {
		h.logger.Error().Err(err).Msg("读取文件内容失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "server_error",
			"description": "读取文件内容失败",
		})
		return
	}

	// 从内存中写入到表单部分
	_, err = part.Write(fileContent)
	if err != nil {
		h.logger.Error().Err(err).Msg("写入文件内容到表单失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "server_error",
			"description": "服务器内部错误",
		})
		return
	}

	writer.Close()

	// 转发到用户服务
	url := fmt.Sprintf("%s/users/me/avatar", h.userServiceURL)
	h.logger.Info().Str("url", url).Msg("准备发送请求到用户服务")

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
	req.Header.Set("Authorization", authHeader)

	// 发送请求
	client := &http.Client{}
	h.logger.Info().Str("url", url).Msg("发送请求到用户服务")
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

	h.logger.Info().Int("status_code", resp.StatusCode).Str("response", string(respBody)).Msg("用户服务响应")

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
	// 检查 Authorization 头是否存在
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		h.logger.Warn().Msg("请求中没有Authorization头")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":       "unauthorized",
			"description": "未提供认证令牌",
		})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		h.logger.Error().Msg("认证中间件未设置userID")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":       "unauthorized",
			"description": "认证失败，未找到用户ID",
		})
		return
	}

	h.logger.Info().Str("userID", fmt.Sprintf("%v", userID)).Str("auth_header", authHeader).Msg("开始处理封面图片上传")

	// 检查请求内容类型
	contentType := c.GetHeader("Content-Type")
	h.logger.Info().Str("content_type", contentType).Msg("请求Content-Type")

	// 验证Content-Type是否为multipart/form-data
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		h.logger.Error().Str("content_type", contentType).Msg("请求Content-Type不是multipart/form-data")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "invalid_content_type",
			"description": "请求Content-Type必须为multipart/form-data",
		})
		return
	}

	// 先解析multipart表单，这一步很重要，修复EOF错误
	err := c.Request.ParseMultipartForm(32 << 20) // 32MB max memory
	if err != nil {
		h.logger.Error().Err(err).Msg("解析multipart表单失败")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "invalid_request",
			"description": "解析multipart表单失败: " + err.Error(),
		})
		return
	}

	// 记录表单字段
	if c.Request.MultipartForm != nil && c.Request.MultipartForm.File != nil {
		for field := range c.Request.MultipartForm.File {
			h.logger.Info().Str("field", field).Msg("发现表单字段")
		}
	}

	// 获取上传的文件
	var file multipart.File
	var header *multipart.FileHeader

	// 依次尝试可能的字段名
	possibleFields := []string{"cover_image", "coverImage", "cover", "file", "image", "upload", "userCover"}
	var lastErr error
	var foundField string

	for _, field := range possibleFields {
		file, header, err = c.Request.FormFile(field)
		if err == nil {
			foundField = field
			h.logger.Info().Str("field", field).Msg("成功从字段获取文件")
			break
		}
		lastErr = err
		h.logger.Debug().Err(err).Str("field", field).Msg("尝试获取字段失败")
	}

	if foundField == "" {
		h.logger.Error().Err(lastErr).Msg("所有可能的字段名尝试均失败")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "invalid_request",
			"description": "获取上传文件失败，请确保表单字段名正确",
		})
		return
	}
	defer file.Close()

	h.logger.Info().Str("filename", header.Filename).Int64("size", header.Size).Str("field", foundField).Msg("成功获取上传文件")

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

	// 先将文件内容读取到内存中
	fileContent, err := io.ReadAll(file)
	if err != nil {
		h.logger.Error().Err(err).Msg("读取文件内容失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "server_error",
			"description": "读取文件内容失败",
		})
		return
	}

	// 从内存中写入到表单部分
	_, err = part.Write(fileContent)
	if err != nil {
		h.logger.Error().Err(err).Msg("写入文件内容到表单失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "server_error",
			"description": "服务器内部错误",
		})
		return
	}

	writer.Close()

	// 转发到用户服务
	url := fmt.Sprintf("%s/users/me/cover-image", h.userServiceURL)
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
	req.Header.Set("Authorization", authHeader)

	// 发送请求
	client := &http.Client{}
	h.logger.Info().Str("url", url).Msg("发送请求到用户服务")
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

// GetUserProfileByUsername 通过用户名获取用户资料
func (h *UserHandler) GetUserProfileByUsername(c *gin.Context) {
	// 获取用户名参数
	username := c.Param("username")
	if username == "" {
		h.logger.Error().Msg("缺少用户名参数")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "invalid_request",
			"description": "缺少用户名",
		})
		return
	}

	// 记录当前网关配置和服务URL
	h.logger.Info().
		Str("username", username).
		Str("userServiceURL", h.userServiceURL).
		Msg("通过用户名获取用户资料请求")

	// 使用基本的HandleRequest方法，让基类处理URL前缀问题
	h.logger.Info().
		Str("username", username).
		Msg("使用基本HandleRequest方法转发请求")

	h.HandleRequest(c, "user")
}
