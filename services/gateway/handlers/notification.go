package handlers

import (
	"backend/services/gateway/config"
	
	"github.com/gin-gonic/gin"
)

// NotificationHandler 处理通知相关的请求
type NotificationHandler struct {
	BaseHandler
}

// NewNotificationHandler 创建一个新的通知处理程序
func NewNotificationHandler(cfg *config.Config) *NotificationHandler {
	return &NotificationHandler{
		BaseHandler: BaseHandler{Config: cfg},
	}
}

// GetUserNotifications 获取用户通知
func (h *NotificationHandler) GetUserNotifications(c *gin.Context) {
	h.HandleRequest(c, "notification")
}

// MarkNotificationAsRead 标记通知为已读
func (h *NotificationHandler) MarkNotificationAsRead(c *gin.Context) {
	h.HandleRequest(c, "notification")
}

// MarkAllNotificationsAsRead 标记所有通知为已读
func (h *NotificationHandler) MarkAllNotificationsAsRead(c *gin.Context) {
	h.HandleRequest(c, "notification")
}

// GetNotificationSettings 获取通知设置
func (h *NotificationHandler) GetNotificationSettings(c *gin.Context) {
	h.HandleRequest(c, "notification")
}

// UpdateNotificationSettings 更新通知设置
func (h *NotificationHandler) UpdateNotificationSettings(c *gin.Context) {
	h.HandleRequest(c, "notification")
}

// RegisterDevice 注册设备
func (h *NotificationHandler) RegisterDevice(c *gin.Context) {
	h.HandleRequest(c, "notification")
}

// UnregisterDevice 注销设备
func (h *NotificationHandler) UnregisterDevice(c *gin.Context) {
	h.HandleRequest(c, "notification")
} 