package handlers

import (
	"backend/services/gateway/config"
	
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
	
	// 将请求委托给交互服务
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
	
	// 将请求委托给交互服务
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