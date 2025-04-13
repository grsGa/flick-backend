package service

import (
	"context"
	"errors"
	"time"

	"backend/pkg/models"
	"backend/pkg/utils"
	"backend/services/notification/repository"

	"github.com/rs/zerolog/log"
)

// PostgresService 实现通知服务接口
type PostgresService struct {
	repo repository.NotificationRepository
}

// NewPostgresService 创建新的PostgresSQL通知服务
func NewPostgresService(repo repository.NotificationRepository) *PostgresService {
	return &PostgresService{
		repo: repo,
	}
}

// CreateNotification 创建单个通知
func (s *PostgresService) CreateNotification(ctx context.Context, userID, senderID, entityID, entityType, notificationType, message string, data map[string]interface{}) (*models.Notification, error) {
	notification := &models.Notification{
		ID:         utils.GenerateUUID(),
		UserID:     userID,
		ActorID:    senderID,
		TargetID:   entityID,
		TargetType: entityType,
		Type:       models.NotificationType(notificationType),
		Data:       data,
		Status:     models.NotificationStatusUnread,
		CreatedAt:  time.Now(),
	}

	err := s.repo.CreateNotification(ctx, notification)
	if err != nil {
		log.Error().Err(err).
			Str("user_id", userID).
			Str("notification_type", notificationType).
			Msg("创建通知失败")
		return nil, err
	}

	// 检查用户通知设置，决定是否要发送推送通知
	settings, err := s.repo.GetUserNotificationSettings(ctx, userID)
	if err == nil && settings != nil && settings.PushEnabled {
		// 检查通知类型是否在用户允许的推送类型列表中
		if utils.Contains([]string{"all", notificationType}, notificationType) {
			// 检查是否在静音时间
			if !s.isQuietHours(settings.QuietHoursStart, settings.QuietHoursEnd) {
				// 获取用户设备并发送推送通知
				devices, err := s.repo.GetUserDevices(ctx, userID)
				if err == nil && len(devices) > 0 {
					// 获取通知模板
					template, _ := s.repo.GetNotificationTemplate(ctx, notificationType)
					title := notificationType
					body := message

					if template != nil {
						title = utils.RenderTemplate(template.Title, map[string]interface{}{
							"user_id":   userID,
							"sender_id": senderID,
							"entity_id": entityID,
							"type":      notificationType,
							"data":      data,
						})
						body = utils.RenderTemplate(template.Content, map[string]interface{}{
							"user_id":   userID,
							"sender_id": senderID,
							"entity_id": entityID,
							"type":      notificationType,
							"data":      data,
						})
					}

					for _, device := range devices {
						if device.NotificationsEnabled {
							// 在实际应用中，这里会调用FCM、APNS等推送服务
							// 这里只是记录日志
							log.Info().
								Str("device_token", device.DeviceToken).
								Str("device_type", device.DeviceType).
								Str("title", title).
								Str("body", body).
								Msg("发送推送通知")
						}
					}
				}
			}
		}
	}

	return notification, nil
}

// CreateBulkNotifications 批量创建通知
func (s *PostgresService) CreateBulkNotifications(ctx context.Context, userIDs []string, senderID, entityID, entityType, notificationType, message string, data map[string]interface{}) error {
	notifications := make([]*models.Notification, 0, len(userIDs))

	for _, userID := range userIDs {
		notification := &models.Notification{
			ID:         utils.GenerateUUID(),
			UserID:     userID,
			ActorID:    senderID,
			TargetID:   entityID,
			TargetType: entityType,
			Type:       models.NotificationType(notificationType),
			Data:       data,
			Status:     models.NotificationStatusUnread,
			CreatedAt:  time.Now(),
		}
		notifications = append(notifications, notification)
	}

	err := s.repo.CreateBulkNotifications(ctx, notifications)
	if err != nil {
		log.Error().Err(err).
			Int("user_count", len(userIDs)).
			Str("notification_type", notificationType).
			Msg("批量创建通知失败")
		return err
	}

	// 为每个用户处理推送通知
	for _, userID := range userIDs {
		// 注意：实际应用中应该用批量处理或消息队列优化
		settings, err := s.repo.GetUserNotificationSettings(ctx, userID)
		if err == nil && settings != nil && settings.PushEnabled {
			if utils.Contains([]string{"all", notificationType}, notificationType) {
				if !s.isQuietHours(settings.QuietHoursStart, settings.QuietHoursEnd) {
					devices, err := s.repo.GetUserDevices(ctx, userID)
					if err == nil && len(devices) > 0 {
						// 获取通知模板
						template, _ := s.repo.GetNotificationTemplate(ctx, notificationType)
						title := notificationType
						body := message

						if template != nil {
							title = utils.RenderTemplate(template.Title, map[string]interface{}{
								"user_id":   userID,
								"sender_id": senderID,
								"entity_id": entityID,
								"type":      notificationType,
								"data":      data,
							})
							body = utils.RenderTemplate(template.Content, map[string]interface{}{
								"user_id":   userID,
								"sender_id": senderID,
								"entity_id": entityID,
								"type":      notificationType,
								"data":      data,
							})
						}

						// 只记录日志，实际应用中应调用推送服务
						log.Info().
							Str("user_id", userID).
							Int("device_count", len(devices)).
							Str("title", title).
							Str("body", body).
							Msg("批量发送推送通知")
					}
				}
			}
		}
	}

	return nil
}

// GetNotificationByID 获取单个通知
func (s *PostgresService) GetNotificationByID(ctx context.Context, id string) (*models.Notification, error) {
	notification, err := s.repo.GetNotificationByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Str("notification_id", id).Msg("获取通知失败")
		return nil, err
	}
	return notification, nil
}

// GetUserNotifications 获取用户的通知
func (s *PostgresService) GetUserNotifications(ctx context.Context, userID string, page, pageSize int) ([]*models.Notification, int64, error) {
	return s.repo.GetUserNotifications(ctx, userID, page, pageSize)
}

// GetUserNotificationsByType 获取指定类型的用户通知
func (s *PostgresService) GetUserNotificationsByType(ctx context.Context, userID string, notificationType string, page, pageSize int) ([]*models.Notification, int64, error) {
	return s.repo.GetUserNotificationsByType(ctx, userID, notificationType, page, pageSize)
}

// GetUnreadNotificationCount 获取未读通知数量
func (s *PostgresService) GetUnreadNotificationCount(ctx context.Context, userID string) (int64, error) {
	return s.repo.GetUnreadNotificationCount(ctx, userID)
}

// MarkNotificationAsRead 标记通知为已读
func (s *PostgresService) MarkNotificationAsRead(ctx context.Context, id string) error {
	return s.repo.MarkNotificationAsRead(ctx, id)
}

// MarkNotificationsAsRead 标记多个通知为已读
func (s *PostgresService) MarkNotificationsAsRead(ctx context.Context, ids []string) error {
	return s.repo.MarkNotificationsAsRead(ctx, ids)
}

// MarkAllNotificationsAsRead 标记所有用户通知为已读
func (s *PostgresService) MarkAllNotificationsAsRead(ctx context.Context, userID string) error {
	return s.repo.MarkAllNotificationsAsRead(ctx, userID)
}

// DeleteNotification 删除通知
func (s *PostgresService) DeleteNotification(ctx context.Context, id string) error {
	return s.repo.DeleteNotification(ctx, id)
}

// DeleteUserNotifications 删除用户所有通知
func (s *PostgresService) DeleteUserNotifications(ctx context.Context, userID string) error {
	return s.repo.DeleteUserNotifications(ctx, userID)
}

// GetUserNotificationSettings 获取用户通知设置
func (s *PostgresService) GetUserNotificationSettings(ctx context.Context, userID string) (*models.NotificationPreference, error) {
	settings, err := s.repo.GetUserNotificationSettings(ctx, userID)
	if err != nil {
		// 如果未找到设置，创建默认设置
		if errors.Is(err, repository.ErrNotFound) {
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
				CreatedAt:         time.Now(),
				UpdatedAt:         time.Now(),
			}

			err = s.repo.CreateOrUpdateUserNotificationSettings(ctx, settings)
			if err != nil {
				log.Error().Err(err).Str("user_id", userID).Msg("创建默认通知设置失败")
				return nil, err
			}
			return settings, nil
		}

		log.Error().Err(err).Str("user_id", userID).Msg("获取用户通知设置失败")
		return nil, err
	}

	return settings, nil
}

// CreateOrUpdateUserNotificationSettings 创建或更新用户通知设置
func (s *PostgresService) CreateOrUpdateUserNotificationSettings(ctx context.Context, settings *models.NotificationPreference) error {
	settings.UpdatedAt = time.Now()
	return s.repo.CreateOrUpdateUserNotificationSettings(ctx, settings)
}

// EnablePushNotifications 启用或禁用推送通知
func (s *PostgresService) EnablePushNotifications(ctx context.Context, userID string, enabled bool) error {
	settings, err := s.GetUserNotificationSettings(ctx, userID)
	if err != nil {
		return err
	}

	settings.PushEnabled = enabled
	settings.UpdatedAt = time.Now()

	return s.repo.CreateOrUpdateUserNotificationSettings(ctx, settings)
}

// EnableEmailNotifications 启用或禁用邮件通知
func (s *PostgresService) EnableEmailNotifications(ctx context.Context, userID string, enabled bool) error {
	settings, err := s.GetUserNotificationSettings(ctx, userID)
	if err != nil {
		return err
	}

	settings.EmailEnabled = enabled
	settings.UpdatedAt = time.Now()

	return s.repo.CreateOrUpdateUserNotificationSettings(ctx, settings)
}

// EnableSMSNotifications 启用或禁用短信通知
func (s *PostgresService) EnableSMSNotifications(ctx context.Context, userID string, enabled bool) error {
	settings, err := s.GetUserNotificationSettings(ctx, userID)
	if err != nil {
		return err
	}

	// 更新设置
	settings.UpdatedAt = time.Now()

	return s.repo.CreateOrUpdateUserNotificationSettings(ctx, settings)
}

// UpdateQuietHours 更新静音时间
func (s *PostgresService) UpdateQuietHours(ctx context.Context, userID string, start, end string) error {
	settings, err := s.GetUserNotificationSettings(ctx, userID)
	if err != nil {
		return err
	}

	settings.QuietHoursStart = start
	settings.QuietHoursEnd = end
	settings.UpdatedAt = time.Now()

	return s.repo.CreateOrUpdateUserNotificationSettings(ctx, settings)
}

// RegisterUserDevice 注册用户设备
func (s *PostgresService) RegisterUserDevice(ctx context.Context, userID, deviceToken, deviceType string) error {
	return s.repo.CreateUserDevice(ctx, &models.UserDevice{
		UserID:               userID,
		DeviceToken:          deviceToken,
		DeviceType:           deviceType,
		NotificationsEnabled: true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	})
}

// GetUserDevices 获取用户设备列表
func (s *PostgresService) GetUserDevices(ctx context.Context, userID string) ([]*models.UserDevice, error) {
	return s.repo.GetUserDevices(ctx, userID)
}

// UpdateDeviceNotificationSettings 更新设备通知设置
func (s *PostgresService) UpdateDeviceNotificationSettings(ctx context.Context, userID, deviceToken string, enabled bool) error {
	return s.repo.UpdateDeviceNotificationSettings(ctx, userID, deviceToken, enabled)
}

// UnregisterUserDevice 注销用户设备
func (s *PostgresService) UnregisterUserDevice(ctx context.Context, userID, deviceToken string) error {
	return s.repo.DeleteUserDevice(ctx, userID, deviceToken)
}

// CreateOrUpdateNotificationTemplate 创建或更新通知模板
func (s *PostgresService) CreateOrUpdateNotificationTemplate(ctx context.Context, templateType, titleTemplate, bodyTemplate string) (*models.NotificationTemplate, error) {
	template := &models.NotificationTemplate{
		Type:      models.NotificationType(templateType),
		Title:     titleTemplate,
		Content:   bodyTemplate,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := s.repo.CreateOrUpdateNotificationTemplate(ctx, template)
	if err != nil {
		log.Error().Err(err).Str("type", templateType).Msg("创建或更新通知模板失败")
		return nil, err
	}

	return template, nil
}

// GetNotificationTemplate 获取通知模板
func (s *PostgresService) GetNotificationTemplate(ctx context.Context, templateType string) (*models.NotificationTemplate, error) {
	return s.repo.GetNotificationTemplate(ctx, templateType)
}

// DeleteNotificationTemplate 删除通知模板
func (s *PostgresService) DeleteNotificationTemplate(ctx context.Context, templateType string) error {
	return s.repo.DeleteNotificationTemplate(ctx, templateType)
}

// ListNotificationTemplates 列出所有通知模板
func (s *PostgresService) ListNotificationTemplates(ctx context.Context) ([]*models.NotificationTemplate, error) {
	return s.repo.ListNotificationTemplates(ctx)
}

// SendPushNotification 发送推送通知
func (s *PostgresService) SendPushNotification(ctx context.Context, userID, title, body string, data map[string]interface{}) error {
	// 检查用户通知设置
	settings, err := s.GetUserNotificationSettings(ctx, userID)
	if err != nil {
		return err
	}

	if !settings.PushEnabled {
		log.Info().Str("user_id", userID).Msg("用户已禁用推送通知")
		return nil
	}

	// 检查是否在静音时间
	if s.isQuietHours(settings.QuietHoursStart, settings.QuietHoursEnd) {
		log.Info().Str("user_id", userID).Msg("当前为静音时间，不发送推送通知")
		return nil
	}

	// 获取用户设备
	devices, err := s.repo.GetUserDevices(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("获取用户设备失败")
		return err
	}

	if len(devices) == 0 {
		log.Info().Str("user_id", userID).Msg("用户没有注册设备")
		return nil
	}

	// 向所有设备发送推送通知
	for _, device := range devices {
		if device.NotificationsEnabled {
			// 这里应实现实际推送逻辑，如FCM、APNS等
			log.Info().
				Str("user_id", userID).
				Str("device_token", device.DeviceToken).
				Str("device_type", device.DeviceType).
				Str("title", title).
				Str("body", body).
				Msg("发送推送通知")
		}
	}

	return nil
}

// SendEmailNotification 发送邮件通知
func (s *PostgresService) SendEmailNotification(ctx context.Context, userID, subject, content string) error {
	// 检查用户通知设置
	settings, err := s.GetUserNotificationSettings(ctx, userID)
	if err != nil {
		return err
	}

	if !settings.EmailEnabled {
		log.Info().Str("user_id", userID).Msg("用户已禁用邮件通知")
		return nil
	}

	// 实际应用中，需获取用户邮箱并调用邮件服务
	log.Info().
		Str("user_id", userID).
		Str("subject", subject).
		Str("content", content).
		Msg("发送邮件通知")

	return nil
}

// SendSMSNotification 发送短信通知
func (s *PostgresService) SendSMSNotification(ctx context.Context, userID, message string) error {
	log.Info().
		Str("user_id", userID).
		Str("message", message).
		Msg("发送短信通知")

	return nil
}

// NotifyUser 综合通知用户（创建通知记录并尝试通过所有通道发送）
func (s *PostgresService) NotifyUser(ctx context.Context, userID, notificationType, entityID, entityType, senderID string, data map[string]interface{}) error {
	// 获取通知模板
	template, err := s.repo.GetNotificationTemplate(ctx, notificationType)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		log.Error().Err(err).Str("type", notificationType).Msg("获取通知模板失败")
		return err
	}

	// 准备通知内容
	message := notificationType
	if template != nil {
		message = utils.RenderTemplate(template.Content, map[string]interface{}{
			"user_id":   userID,
			"sender_id": senderID,
			"entity_id": entityID,
			"type":      notificationType,
			"data":      data,
		})
	}

	// 创建通知记录
	_, err = s.CreateNotification(ctx, userID, senderID, entityID, entityType, notificationType, message, data)
	if err != nil {
		return err
	}

	// 获取用户通知设置
	settings, err := s.GetUserNotificationSettings(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("获取用户通知设置失败")
		return err
	}

	// 发送推送通知
	if settings.PushEnabled {
		title := notificationType
		body := message

		if template != nil {
			title = utils.RenderTemplate(template.Title, map[string]interface{}{
				"user_id":   userID,
				"sender_id": senderID,
				"entity_id": entityID,
				"type":      notificationType,
				"data":      data,
			})
		}

		_ = s.SendPushNotification(ctx, userID, title, body, data)
	}

	// 发送邮件通知
	if settings.EmailEnabled {
		subject := notificationType
		content := message

		if template != nil {
			subject = utils.RenderTemplate(template.Title, map[string]interface{}{
				"user_id":   userID,
				"sender_id": senderID,
				"entity_id": entityID,
				"type":      notificationType,
				"data":      data,
			})
		}

		_ = s.SendEmailNotification(ctx, userID, subject, content)
	}

	return nil
}

// NotifyUsers 通知多个用户
func (s *PostgresService) NotifyUsers(ctx context.Context, userIDs []string, notificationType, entityID, entityType, senderID string, data map[string]interface{}) error {
	// 首先批量创建通知记录
	err := s.CreateBulkNotifications(ctx, userIDs, senderID, entityID, entityType, notificationType, notificationType, data)
	if err != nil {
		return err
	}

	// 获取通知模板
	template, err := s.repo.GetNotificationTemplate(ctx, notificationType)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		log.Error().Err(err).Str("type", notificationType).Msg("获取通知模板失败")
	}

	// 为每个用户处理通知发送
	// 注意：实际应用中应该批量处理或使用消息队列
	for _, userID := range userIDs {
		// 获取用户通知设置
		settings, err := s.GetUserNotificationSettings(ctx, userID)
		if err != nil {
			log.Error().Err(err).Str("user_id", userID).Msg("获取用户通知设置失败")
			continue
		}

		message := notificationType
		if template != nil {
			message = utils.RenderTemplate(template.Content, map[string]interface{}{
				"user_id":   userID,
				"sender_id": senderID,
				"entity_id": entityID,
				"type":      notificationType,
				"data":      data,
			})
		}

		// 发送推送通知
		if settings.PushEnabled {
			title := notificationType
			if template != nil {
				title = utils.RenderTemplate(template.Title, map[string]interface{}{
					"user_id":   userID,
					"sender_id": senderID,
					"entity_id": entityID,
					"type":      notificationType,
					"data":      data,
				})
			}

			_ = s.SendPushNotification(ctx, userID, title, message, data)
		}

		// 发送邮件通知
		if settings.EmailEnabled {
			subject := notificationType
			if template != nil {
				subject = utils.RenderTemplate(template.Title, map[string]interface{}{
					"user_id":   userID,
					"sender_id": senderID,
					"entity_id": entityID,
					"type":      notificationType,
					"data":      data,
				})
			}

			_ = s.SendEmailNotification(ctx, userID, subject, message)
		}
	}

	return nil
}

// isQuietHours 检查当前是否是静音时间
func (s *PostgresService) isQuietHours(start, end string) bool {
	now := time.Now()

	// 解析静音时间
	startTime, err := time.Parse("15:04", start)
	if err != nil {
		return false
	}

	endTime, err := time.Parse("15:04", end)
	if err != nil {
		return false
	}

	// 将当前时间格式化为只有时分的格式
	nowTime, _ := time.Parse("15:04", now.Format("15:04"))

	// 如果结束时间小于开始时间，表示跨天
	if endTime.Before(startTime) {
		return nowTime.After(startTime) || nowTime.Before(endTime)
	}

	// 常规情况：当前时间在开始和结束时间之间
	return nowTime.After(startTime) && nowTime.Before(endTime)
}
