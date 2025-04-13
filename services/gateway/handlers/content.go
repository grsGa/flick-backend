package handlers

import (
	"backend/services/gateway/config"
	
	"github.com/gin-gonic/gin"
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