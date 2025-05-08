package handlers

import (
	"backend/services/gateway/config"

	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// InteractionHandler 处理交互相关的请求
type InteractionHandler struct {
	BaseHandler
}

// NewInteractionHandler 创建一个新的交互处理程序
func NewInteractionHandler(cfg *config.Config) *InteractionHandler {
	return &InteractionHandler{
		BaseHandler: BaseHandler{Config: cfg},
	}
}

// LikePost 点赞内容
func (h *InteractionHandler) LikePost(c *gin.Context) {
	h.HandleRequest(c, "interaction")
}

// UnlikePost 取消点赞内容
func (h *InteractionHandler) UnlikePost(c *gin.Context) {
	h.HandleRequest(c, "interaction")
}

// CommentOnPost 评论内容
func (h *InteractionHandler) CommentOnPost(c *gin.Context) {
	h.HandleRequest(c, "interaction")
}

// GetPostComments 获取内容的评论
func (h *InteractionHandler) GetPostComments(c *gin.Context) {
	h.HandleRequest(c, "interaction")
}

// UpdateComment 更新评论
func (h *InteractionHandler) UpdateComment(c *gin.Context) {
	h.HandleRequest(c, "interaction")
}

// DeleteComment 删除评论
func (h *InteractionHandler) DeleteComment(c *gin.Context) {
	h.HandleRequest(c, "interaction")
}

// LikeComment 点赞评论
func (h *InteractionHandler) LikeComment(c *gin.Context) {
	h.HandleRequest(c, "interaction")
}

// UnlikeComment 取消点赞评论
func (h *InteractionHandler) UnlikeComment(c *gin.Context) {
	h.HandleRequest(c, "interaction")
}

// LikePostByPermalink 通过用户名和永久链接ID点赞帖子
func (h *InteractionHandler) LikePostByPermalink(c *gin.Context) {
	// 路径形如 /:username/status/:permalink_id/like
	username := c.Param("username")
	permalinkID := c.Param("permalink_id")

	log.Info().
		Str("username", username).
		Str("permalink_id", permalinkID).
		Msg("通过永久链接点赞帖子")

	// 修改请求路径为内部格式
	c.Request.URL.Path = fmt.Sprintf("/posts/%s/like", permalinkID)

	// 转发请求到interaction服务
	h.HandleRequest(c, "interaction")
}

// UnlikePostByPermalink 通过用户名和永久链接ID取消点赞帖子
func (h *InteractionHandler) UnlikePostByPermalink(c *gin.Context) {
	// 路径形如 /:username/status/:permalink_id/like
	username := c.Param("username")
	permalinkID := c.Param("permalink_id")

	log.Info().
		Str("username", username).
		Str("permalink_id", permalinkID).
		Msg("通过永久链接取消点赞帖子")

	// 修改请求路径为内部格式
	c.Request.URL.Path = fmt.Sprintf("/posts/%s/like", permalinkID)

	// 转发请求到interaction服务
	h.HandleRequest(c, "interaction")
}

// BookmarkPostByPermalink 通过用户名和永久链接ID收藏帖子
func (h *InteractionHandler) BookmarkPostByPermalink(c *gin.Context) {
	// 路径形如 /:username/status/:permalink_id/save
	username := c.Param("username")
	permalinkID := c.Param("permalink_id")

	log.Info().
		Str("username", username).
		Str("permalink_id", permalinkID).
		Msg("通过永久链接收藏帖子")

	// 修改请求路径为内部格式
	c.Request.URL.Path = fmt.Sprintf("/posts/%s/save", permalinkID)

	// 转发请求到interaction服务
	h.HandleRequest(c, "interaction")
}

// UnbookmarkPostByPermalink 通过用户名和永久链接ID取消收藏帖子
func (h *InteractionHandler) UnbookmarkPostByPermalink(c *gin.Context) {
	// 路径形如 /:username/status/:permalink_id/save
	username := c.Param("username")
	permalinkID := c.Param("permalink_id")

	log.Info().
		Str("username", username).
		Str("permalink_id", permalinkID).
		Msg("通过永久链接取消收藏帖子")

	// 修改请求路径为内部格式
	c.Request.URL.Path = fmt.Sprintf("/posts/%s/save", permalinkID)

	// 转发请求到interaction服务
	h.HandleRequest(c, "interaction")
}

// CommentOnPostByPermalink 通过用户名和永久链接ID评论帖子
func (h *InteractionHandler) CommentOnPostByPermalink(c *gin.Context) {
	// 路径形如 /:username/status/:permalink_id/comments
	username := c.Param("username")
	permalinkID := c.Param("permalink_id")

	log.Info().
		Str("username", username).
		Str("permalink_id", permalinkID).
		Msg("通过永久链接评论帖子")

	// 将请求委托给交互服务
	h.HandleRequest(c, "interaction")
}

// GetPostCommentsByPermalink 通过用户名和永久链接ID获取帖子评论
func (h *InteractionHandler) GetPostCommentsByPermalink(c *gin.Context) {
	// 路径形如 /:username/status/:permalink_id/comments
	username := c.Param("username")
	permalinkID := c.Param("permalink_id")

	log.Info().
		Str("username", username).
		Str("permalink_id", permalinkID).
		Msg("通过永久链接获取帖子评论")

	// 将请求委托给交互服务
	h.HandleRequest(c, "interaction")
}

// GetCurrentUserLikedPosts 获取当前用户点赞的帖子列表
func (h *InteractionHandler) GetCurrentUserLikedPosts(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	// 将用户ID添加到查询参数
	query := c.Request.URL.Query()
	query.Set("user_id", fmt.Sprintf("%v", userID))
	c.Request.URL.RawQuery = query.Encode()

	// 设置请求路径为内部格式 - 修改为与交互服务匹配的路径格式
	userIDStr := fmt.Sprintf("%v", userID)
	c.Request.URL.Path = fmt.Sprintf("/users/%s/liked-posts", userIDStr)

	log.Info().
		Str("user_id", userIDStr).
		Str("path", c.Request.URL.Path).
		Msg("获取用户点赞帖子")

	// 转发请求到interaction服务
	h.HandleRequest(c, "interaction")
}

// GetUserLikedPosts 获取用户点赞的帖子列表（简化版的转发）
func (h *InteractionHandler) GetUserLikedPosts(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	// 将用户ID添加到查询参数
	query := c.Request.URL.Query()
	query.Set("user_id", fmt.Sprintf("%v", userID))
	c.Request.URL.RawQuery = query.Encode()

	// 保持原始路径不变，直接使用/users/me/liked-posts
	// 不要修改路径，交互服务已经专门注册了这个路径的处理器

	// 记录详细转发信息
	log.Info().
		Str("user_id", fmt.Sprintf("%v", userID)).
		Str("path", c.Request.URL.Path).
		Str("method", c.Request.Method).
		Msg("获取用户点赞帖子 - 正在转发请求")

	// 转发请求到interaction服务
	h.HandleRequest(c, "interaction")
}

// GetUserBookmarks 获取当前用户的书签
func (h *InteractionHandler) GetUserBookmarks(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	// 将用户ID添加到查询参数
	query := c.Request.URL.Query()
	query.Set("user_id", fmt.Sprintf("%v", userID))
	c.Request.URL.RawQuery = query.Encode()

	// 保持原始路径不变，直接使用/users/me/bookmarks
	// 记录详细转发信息
	log.Info().
		Str("user_id", fmt.Sprintf("%v", userID)).
		Str("path", c.Request.URL.Path).
		Str("method", c.Request.Method).
		Msg("获取用户书签 - 正在转发请求")

	// 转发请求到interaction服务
	h.HandleRequest(c, "interaction")
}
