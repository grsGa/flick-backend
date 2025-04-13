package handlers

import (
	"backend/services/gateway/config"
	
	"github.com/gin-gonic/gin"
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