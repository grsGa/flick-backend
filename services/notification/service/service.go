package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"backend/pkg/models"
	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog"
	"github.com/gorilla/mux"

	"backend/services/notification/repository"
)

// NotificationService 定义通知服务接口
type NotificationService interface {
	// 通知创建和管理
	CreateNotification(ctx context.Context, notification *models.Notification) error
	CreateBulkNotifications(ctx context.Context, notifications []*models.Notification) error
	GetNotificationByID(ctx context.Context, notificationID string) (*models.Notification, error)
	GetUserNotifications(ctx context.Context, userID string, page, pageSize int) ([]*models.Notification, int64, error)
	GetUserNotificationsByType(ctx context.Context, userID string, notificationType string, page, pageSize int) ([]*models.Notification, int64, error)
	GetUnreadNotificationCount(ctx context.Context, userID string) (int64, error)
	MarkNotificationAsRead(ctx context.Context, notificationID string) error
	MarkNotificationsAsRead(ctx context.Context, notificationIDs []string) error
	MarkAllNotificationsAsRead(ctx context.Context, userID string) error
	DeleteNotification(ctx context.Context, notificationID string) error
	DeleteUserNotifications(ctx context.Context, userID string) error
	
	// 通知生成和发送
	GenerateFollowNotification(ctx context.Context, followerID, followedID string) error
	GenerateLikeNotification(ctx context.Context, userID, contentType, contentID string) error
	GenerateCommentNotification(ctx context.Context, commentID, authorID string) error
	GenerateMentionNotification(ctx context.Context, mentionerID, mentionedID, contentType, contentID string) error
	GenerateDirectMessageNotification(ctx context.Context, messageID, senderID, receiverID string) error
	GenerateSystemNotification(ctx context.Context, userIDs []string, title, message string, metadata map[string]interface{}) error
	
	// 发送通知
	SendNotification(ctx context.Context, notification *models.Notification) error
	SendBulkNotifications(ctx context.Context, notifications []*models.Notification) error
	
	// 通知设置管理
	GetUserNotificationSettings(ctx context.Context, userID string) (*models.NotificationPreference, error)
	UpdateUserNotificationSettings(ctx context.Context, settings *models.NotificationPreference) error
	
	// 设备管理
	RegisterDevice(ctx context.Context, device *models.UserDevice) error
	UpdateUserDevice(ctx context.Context, device *models.UserDevice) error
	UnregisterDevice(ctx context.Context, userID, deviceToken string) error
	GetUserDevices(ctx context.Context, userID string) ([]*models.UserDevice, error)
	
	// 通知模板管理
	CreateNotificationTemplate(ctx context.Context, template *models.NotificationTemplate) error
	GetNotificationTemplate(ctx context.Context, templateType string) (*models.NotificationTemplate, error)
	UpdateNotificationTemplate(ctx context.Context, template *models.NotificationTemplate) error
	DeleteNotificationTemplate(ctx context.Context, templateType string) error
	ListNotificationTemplates(ctx context.Context) ([]*models.NotificationTemplate, error)
	RenderNotificationTemplate(ctx context.Context, templateType string, data map[string]interface{}) (string, string, error)
}

// NotificationCreateRequest 创建通知请求
type NotificationCreateRequest struct {
	RecipientID    string                 `json:"recipient_id" binding:"required"`
	SenderID       string                 `json:"sender_id"`
	Type           string                 `json:"type" binding:"required"`
	ReferenceID    string                 `json:"reference_id"`
	ReferenceType  string                 `json:"reference_type"`
	Message        string                 `json:"message" binding:"required"`
	AdditionalData map[string]interface{} `json:"additional_data"`
	DeliveryStatus string                 `json:"delivery_status"`
}

// NotificationSettingsRequest 通知设置请求
type NotificationSettingsRequest struct {
	UserID              string `json:"user_id" binding:"required"`
	LikesEnabled        bool   `json:"likes_enabled"`
	CommentsEnabled     bool   `json:"comments_enabled"`
	FollowsEnabled      bool   `json:"follows_enabled"`
	MentionsEnabled     bool   `json:"mentions_enabled"`
	SharesEnabled       bool   `json:"shares_enabled"`
	DirectMessagesEnabled bool  `json:"direct_messages_enabled"`
	SystemEnabled       bool   `json:"system_enabled"`
	EmailEnabled        bool   `json:"email_enabled"`
	PushEnabled         bool   `json:"push_enabled"`
	SMSEnabled          bool   `json:"sms_enabled"`
}

// DeviceRegisterRequest 设备注册请求
type DeviceRegisterRequest struct {
	UserID       string `json:"user_id" binding:"required"`
	DeviceToken  string `json:"device_token" binding:"required"`
	DeviceType   string `json:"device_type" binding:"required"` // ios, android, web
	AppVersion   string `json:"app_version"`
	DeviceModel  string `json:"device_model"`
	OSVersion    string `json:"os_version"`
}

// NotificationBatchRequest 批量通知请求
type NotificationBatchRequest struct {
	RecipientIDs []string               `json:"recipient_ids" binding:"required"`
	Type         string                 `json:"type" binding:"required"`
	Message      string                 `json:"message" binding:"required"`
	AdditionalData map[string]interface{} `json:"additional_data"`
}

// NotificationServiceImpl 通知服务实现
type NotificationServiceImpl struct {
	repo        *repository.Repository
	redisClient *redis.Client
	logger      zerolog.Logger
}

// NewNotificationService 创建通知服务实例
func NewNotificationService(repo *repository.Repository, redisClient *redis.Client, logger zerolog.Logger) *NotificationServiceImpl {
	return &NotificationServiceImpl{
		repo:        repo,
		redisClient: redisClient,
		logger:      logger,
	}
}

// GetNotificationsHandler 处理获取所有通知请求
func (s *NotificationServiceImpl) GetNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// 从请求参数中获取用户ID
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "缺少用户ID", http.StatusBadRequest)
		return
	}
	
	// 获取分页参数
	page := 1
	pageSize := 20
	
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}
	
	// 获取通知类型过滤参数
	notificationType := r.URL.Query().Get("type")
	
	var notifications []*models.Notification
	var total int64
	var err error
	
	// 根据是否有类型过滤调用不同的仓库方法
	if notificationType != "" {
		notifications, total, err = s.repo.GetUserNotificationsByType(ctx, userID, notificationType, page, pageSize)
	} else {
		notifications, total, err = s.repo.GetUserNotifications(ctx, userID, page, pageSize)
	}
	
	if err != nil {
		s.logger.Error().Err(err).
			Str("user_id", userID).
			Int("page", page).
			Int("page_size", pageSize).
			Str("type", notificationType).
			Msg("获取通知列表失败")
		http.Error(w, "获取通知列表失败", http.StatusInternalServerError)
		return
	}
	
	// 构建响应
	resp := struct {
		Notifications []*models.Notification `json:"notifications"`
		Total         int64                  `json:"total"`
		Page          int                    `json:"page"`
		PageSize      int                    `json:"page_size"`
	}{
		Notifications: notifications,
		Total:         total,
		Page:          page,
		PageSize:      pageSize,
	}
	
	// 序列化为JSON并返回
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error().Err(err).Msg("序列化通知响应失败")
		http.Error(w, "序列化通知响应失败", http.StatusInternalServerError)
		return
	}
}

// GetNotificationHandler 处理获取单个通知请求
func (s *NotificationServiceImpl) GetNotificationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// 从URL路径中获取通知ID
	vars := mux.Vars(r)
	notificationID := vars["id"]
	if notificationID == "" {
		http.Error(w, "缺少通知ID", http.StatusBadRequest)
		return
	}
	
	// 获取通知详情
	notification, err := s.repo.GetNotificationByID(ctx, notificationID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("notification_id", notificationID).
			Msg("获取通知详情失败")
			
		if errors.Is(err, repository.ErrNotificationNotFound) {
			http.Error(w, "通知不存在", http.StatusNotFound)
		} else {
			http.Error(w, "获取通知详情失败", http.StatusInternalServerError)
		}
		return
	}
	
	// 序列化为JSON并返回
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(notification); err != nil {
		s.logger.Error().Err(err).Msg("序列化通知详情失败")
		http.Error(w, "序列化通知详情失败", http.StatusInternalServerError)
		return
	}
}

// MarkAsReadHandler 处理标记通知为已读请求
func (s *NotificationServiceImpl) MarkAsReadHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// 从URL路径中获取通知ID
	vars := mux.Vars(r)
	notificationID := vars["id"]
	if notificationID == "" {
		http.Error(w, "缺少通知ID", http.StatusBadRequest)
		return
	}
	
	// 标记通知为已读
	err := s.repo.MarkNotificationAsRead(ctx, notificationID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("notification_id", notificationID).
			Msg("标记通知为已读失败")
			
		if errors.Is(err, repository.ErrNotificationNotFound) {
			http.Error(w, "通知不存在", http.StatusNotFound)
		} else {
			http.Error(w, "标记通知为已读失败", http.StatusInternalServerError)
		}
		return
	}
	
	// 返回成功响应
	resp := struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}{
		Success: true,
		Message: "通知已标记为已读",
	}
	
	// 序列化为JSON并返回
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error().Err(err).Msg("序列化响应失败")
		http.Error(w, "序列化响应失败", http.StatusInternalServerError)
		return
	}
}

// GetNotificationSettingsHandler 处理获取通知设置请求
func (s *NotificationServiceImpl) GetNotificationSettingsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// 从请求参数中获取用户ID
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "缺少用户ID", http.StatusBadRequest)
		return
	}
	
	// 获取用户通知设置
	settings, err := s.repo.GetUserNotificationSettings(ctx, userID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("user_id", userID).
			Msg("获取通知设置失败")
			
		if errors.Is(err, repository.ErrNotFound) {
			// 如果未找到设置，创建默认设置并返回
			settings = &models.NotificationPreference{
				UserID:            userID,
				LikeEnabled:       true,
				CommentEnabled:    true,
				MentionEnabled:    true,
				FollowEnabled:     true,
				SystemEnabled:     true,
				MessageEnabled:    true,
				ShareEnabled:      true,
				EmailEnabled:      true,
				PushEnabled:       true,
				QuietHoursEnabled: false,
				QuietHoursStart:   "22:00",
				QuietHoursEnd:     "08:00",
			}
		} else {
			http.Error(w, "获取通知设置失败", http.StatusInternalServerError)
			return
		}
	}
	
	// 序列化为JSON并返回
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(settings); err != nil {
		s.logger.Error().Err(err).Msg("序列化通知设置失败")
		http.Error(w, "序列化通知设置失败", http.StatusInternalServerError)
		return
	}
}

// UpdateNotificationSettingsHandler 处理更新通知设置请求
func (s *NotificationServiceImpl) UpdateNotificationSettingsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// 解析请求体
	var request NotificationSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.logger.Error().Err(err).Msg("解析请求体失败")
		http.Error(w, "请求格式错误", http.StatusBadRequest)
		return
	}
	
	// 验证用户ID
	if request.UserID == "" {
		http.Error(w, "缺少用户ID", http.StatusBadRequest)
		return
	}
	
	// 构建通知设置对象
	settings := &models.NotificationPreference{
		UserID:            request.UserID,
		LikeEnabled:       request.LikesEnabled,
		CommentEnabled:    request.CommentsEnabled,
		FollowEnabled:     request.FollowsEnabled,
		MentionEnabled:    request.MentionsEnabled,
		ShareEnabled:      request.SharesEnabled,
		MessageEnabled:    request.DirectMessagesEnabled,
		SystemEnabled:     request.SystemEnabled,
		EmailEnabled:      request.EmailEnabled,
		PushEnabled:       request.PushEnabled,
	}
	
	// 更新设置
	err := s.repo.CreateOrUpdateUserNotificationSettings(ctx, settings)
	if err != nil {
		s.logger.Error().Err(err).
			Str("user_id", request.UserID).
			Msg("更新通知设置失败")
		http.Error(w, "更新通知设置失败", http.StatusInternalServerError)
		return
	}
	
	// 返回成功响应
	resp := struct {
		Success  bool                          `json:"success"`
		Message  string                        `json:"message"`
		Settings *models.NotificationPreference `json:"settings"`
	}{
		Success:  true,
		Message:  "通知设置已更新",
		Settings: settings,
	}
	
	// 序列化为JSON并返回
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error().Err(err).Msg("序列化响应失败")
		http.Error(w, "序列化响应失败", http.StatusInternalServerError)
		return
	}
} 