package handlers

import (
	"backend/services/gateway/config"

	"net/http"

	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// ContentHandler 处理内容相关的请求
type ContentHandler struct {
	BaseHandler
}

// NewContentHandler 创建一个新的内容处理程序
func NewContentHandler(cfg *config.Config) *ContentHandler {
	return &ContentHandler{
		BaseHandler: BaseHandler{Config: cfg},
	}
}

// GetPublicPosts 获取公开内容
func (h *ContentHandler) GetPublicPosts(c *gin.Context) {
	h.HandleRequest(c, "content")
}

// GetPostByID 获取内容详情
func (h *ContentHandler) GetPostByID(c *gin.Context) {
	h.HandleRequest(c, "content")
}

// CreatePost 创建帖子
func (h *ContentHandler) CreatePost(c *gin.Context) {
	h.HandleRequest(c, "content")
}

// GetPost 获取帖子
func (h *ContentHandler) GetPost(c *gin.Context) {
	h.HandleRequest(c, "content")
}

// UpdatePost 更新帖子
func (h *ContentHandler) UpdatePost(c *gin.Context) {
	h.HandleRequest(c, "content")
}

// DeletePost 删除帖子
func (h *ContentHandler) DeletePost(c *gin.Context) {
	h.HandleRequest(c, "content")
}

// GetUserPosts 获取用户的内容
func (h *ContentHandler) GetUserPosts(c *gin.Context) {
	// 获取用户名参数
	username := c.Param("username")
	if username == "" {
		h.BaseHandler.RespondWithError(c, http.StatusBadRequest, "缺少用户名参数")
		return
	}

	// 记录请求详情
	log.Info().
		Str("method", c.Request.Method).
		Str("path", c.Request.URL.Path).
		Str("username", username).
		Str("service", "content").
		Msg("获取用户帖子请求")

	// 使用基本的转发方法
	h.HandleRequest(c, "content")
}

// GetCurrentUserPosts 获取当前登录用户的内容
func (h *ContentHandler) GetCurrentUserPosts(c *gin.Context) {
	// 从上下文获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		h.BaseHandler.RespondWithError(c, http.StatusUnauthorized, "未经授权的请求")
		return
	}

	// 记录请求详情
	log.Info().
		Str("method", c.Request.Method).
		Str("path", c.Request.URL.Path).
		Str("user_id", fmt.Sprintf("%v", userID)).
		Str("service", "content").
		Msg("获取当前用户帖子请求")

	// 使用基本的转发方法
	h.HandleRequest(c, "content")
}

// GetUserFeed 获取用户的内容流
func (h *ContentHandler) GetUserFeed(c *gin.Context) {
	h.HandleRequest(c, "content")
}

// GetAllPosts 获取所有内容（管理员）
func (h *ContentHandler) GetAllPosts(c *gin.Context) {
	h.HandleRequest(c, "content")
}

// UpdatePostStatus 更新内容状态（管理员）
func (h *ContentHandler) UpdatePostStatus(c *gin.Context) {
	h.HandleRequest(c, "content")
}

// ListPosts 列出帖子
func (h *ContentHandler) ListPosts(c *gin.Context) {
	h.HandleRequest(c, "content")
}

// CreateComment 创建评论
func (h *ContentHandler) CreateComment(c *gin.Context) {
	h.HandleRequest(c, "content")
}

// GetComment 获取评论
func (h *ContentHandler) GetComment(c *gin.Context) {
	h.HandleRequest(c, "content")
}

// UpdateComment 更新评论
func (h *ContentHandler) UpdateComment(c *gin.Context) {
	h.HandleRequest(c, "content")
}

// DeleteComment 删除评论
func (h *ContentHandler) DeleteComment(c *gin.Context) {
	h.HandleRequest(c, "content")
}

// ListComments 列出评论
func (h *ContentHandler) ListComments(c *gin.Context) {
	h.HandleRequest(c, "content")
}

// UploadMedia 上传媒体文件
func (h *ContentHandler) UploadMedia(c *gin.Context) {
	h.HandleRequest(c, "content")
}

// GetMedia 获取媒体文件
func (h *ContentHandler) GetMedia(c *gin.Context) {
	h.HandleRequest(c, "content")
}

// DeleteMedia 删除媒体文件
func (h *ContentHandler) DeleteMedia(c *gin.Context) {
	h.HandleRequest(c, "content")
}

// SearchContent 搜索内容
func (h *ContentHandler) SearchContent(c *gin.Context) {
	h.HandleRequest(c, "content")
}

// VotePoll 对帖子进行投票
func (h *ContentHandler) VotePoll(c *gin.Context) {
	h.HandleRequest(c, "content")
}

// GetPollResults 获取投票结果
func (h *ContentHandler) GetPollResults(c *gin.Context) {
	h.HandleRequest(c, "content")
}

// SavePost 保存帖子
func (h *ContentHandler) SavePost(c *gin.Context) {
	h.HandleRequest(c, "content")
}

// UnsavePost 取消保存帖子
func (h *ContentHandler) UnsavePost(c *gin.Context) {
	h.HandleRequest(c, "content")
}

// GetSavedPosts 获取已保存的帖子
func (h *ContentHandler) GetSavedPosts(c *gin.Context) {
	h.HandleRequest(c, "content")
}

// GetTopPosts 获取热门帖子
func (h *ContentHandler) GetTopPosts(c *gin.Context) {
	h.HandleRequest(c, "content")
}

// GetPostStats 获取帖子统计数据
func (h *ContentHandler) GetPostStats(c *gin.Context) {
	h.HandleRequest(c, "content")
}

// ReportPost 举报帖子
func (h *ContentHandler) ReportPost(c *gin.Context) {
	h.HandleRequest(c, "content")
}

// UploadContent 上传内容文件(图片/视频)
func (h *ContentHandler) UploadContent(c *gin.Context) {
	// 添加详细日志
	log.Info().
		Str("method", c.Request.Method).
		Str("path", c.Request.URL.Path).
		Str("content_type", c.GetHeader("Content-Type")).
		Msg("Gateway收到上传请求")

	// 获取Authorization头，确保它被传递
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		log.Error().Msg("上传请求中缺少Authorization头")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供授权信息"})
		c.Abort()
		return
	}

	// 直接转发请求到content服务，不做任何处理
	h.HandleRequest(c, "content")
}

// GetPostByPermalink 通过用户名和永久链接ID获取帖子
func (h *ContentHandler) GetPostByPermalink(c *gin.Context) {
	// 路径形如 /:username/status/:permalink_id
	// 记录接收到的参数
	username := c.Param("username")
	permalinkID := c.Param("permalink_id")

	log.Info().
		Str("username", username).
		Str("permalink_id", permalinkID).
		Str("path", c.Request.URL.Path).
		Str("method", c.Request.Method).
		Msg("网关接收到通过永久链接获取帖子请求")

	if username == "" || permalinkID == "" {
		log.Error().
			Str("username", username).
			Str("permalink_id", permalinkID).
			Msg("缺少必要的参数")
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求，缺少用户名或帖子ID"})
		return
	}

	// 直接使用原始路径格式进行转发 - 不再需要转换为 "/content/permalink/{username}/{permalink_id}"
	// 内容服务已配置为也接受 "/{username}/status/{permalink_id}" 路径

	// 添加详细日志
	log.Info().
		Str("original_path", c.Request.URL.Path).
		Str("username", username).
		Str("permalink_id", permalinkID).
		Msg("直接转发帖子请求")

	// 使用原始路径转发请求
	h.HandleRequest(c, "content")
}

// GetPostMediaByIndex 获取帖子中特定索引的媒体文件
func (h *ContentHandler) GetPostMediaByIndex(c *gin.Context) {
	// 路径形如 /:username/status/:permalink_id/photo/:index
	username := c.Param("username")
	permalinkID := c.Param("permalink_id")
	index := c.Param("index")

	log.Info().
		Str("username", username).
		Str("permalink_id", permalinkID).
		Str("index", index).
		Msg("获取帖子媒体文件")

	// 直接使用原始路径转发请求
	h.HandleRequest(c, "content")
}

// UpdatePostByPermalink 通过用户名和永久链接ID更新帖子
func (h *ContentHandler) UpdatePostByPermalink(c *gin.Context) {
	// 路径形如 /:username/status/:permalink_id
	username := c.Param("username")
	permalinkID := c.Param("permalink_id")

	log.Info().
		Str("username", username).
		Str("permalink_id", permalinkID).
		Msg("通过永久链接更新帖子")

	// 直接使用原始路径转发请求
	h.HandleRequest(c, "content")
}

// DeletePostByPermalink 通过用户名和永久链接ID删除帖子
func (h *ContentHandler) DeletePostByPermalink(c *gin.Context) {
	// 路径形如 /:username/status/:permalink_id
	username := c.Param("username")
	permalinkID := c.Param("permalink_id")

	log.Info().
		Str("username", username).
		Str("permalink_id", permalinkID).
		Msg("通过永久链接删除帖子")

	// 直接使用原始路径转发请求
	h.HandleRequest(c, "content")
}

// SavePostByPermalink 通过用户名和永久链接ID保存帖子
func (h *ContentHandler) SavePostByPermalink(c *gin.Context) {
	// 路径形如 /:username/status/:permalink_id/save
	username := c.Param("username")
	permalinkID := c.Param("permalink_id")

	log.Info().
		Str("username", username).
		Str("permalink_id", permalinkID).
		Msg("通过永久链接保存帖子")

	// 直接使用原始路径转发请求
	h.HandleRequest(c, "content")
}

// UnsavePostByPermalink 通过用户名和永久链接ID取消保存帖子
func (h *ContentHandler) UnsavePostByPermalink(c *gin.Context) {
	// 路径形如 /:username/status/:permalink_id/save
	username := c.Param("username")
	permalinkID := c.Param("permalink_id")

	log.Info().
		Str("username", username).
		Str("permalink_id", permalinkID).
		Msg("通过永久链接取消保存帖子")

	// 直接使用原始路径转发请求
	h.HandleRequest(c, "content")
}
