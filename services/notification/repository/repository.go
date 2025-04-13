package repository

import (
	"context"
	"errors"

	"backend/pkg/models"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// ErrNotFound 错误定义
var (
	ErrNotFound = errors.New("记录未找到")
)

// NotificationRepository 定义通知仓库接口
type NotificationRepository interface {
	// 通知管理
	CreateNotification(ctx context.Context, notification *models.Notification) error
	CreateBulkNotifications(ctx context.Context, notifications []*models.Notification) error
	GetNotificationByID(ctx context.Context, id string) (*models.Notification, error)
	GetUserNotifications(ctx context.Context, userID string, page, pageSize int) ([]*models.Notification, int64, error)
	GetUserNotificationsByType(ctx context.Context, userID string, notificationType string, page, pageSize int) ([]*models.Notification, int64, error)
	GetUnreadNotificationCount(ctx context.Context, userID string) (int64, error)
	MarkNotificationAsRead(ctx context.Context, notificationID string) error
	MarkNotificationsAsRead(ctx context.Context, notificationIDs []string) error
	MarkAllNotificationsAsRead(ctx context.Context, userID string) error
	DeleteNotification(ctx context.Context, notificationID string) error
	DeleteUserNotifications(ctx context.Context, userID string) error

	// 用户通知设置管理
	GetUserNotificationSettings(ctx context.Context, userID string) (*models.NotificationPreference, error)
	UpdateUserNotificationSettings(ctx context.Context, settings *models.NotificationPreference) error
	CreateOrUpdateUserNotificationSettings(ctx context.Context, settings *models.NotificationPreference) error

	// 用户设备管理
	RegisterDevice(ctx context.Context, device *models.UserDevice) error
	UpdateUserDevice(ctx context.Context, device *models.UserDevice) error
	UnregisterDevice(ctx context.Context, userID string, deviceToken string) error
	GetUserDevices(ctx context.Context, userID string) ([]*models.UserDevice, error)
	CreateUserDevice(ctx context.Context, device *models.UserDevice) error
	UpdateDeviceNotificationSettings(ctx context.Context, userID, deviceToken string, enabled bool) error
	DeleteUserDevice(ctx context.Context, userID, deviceToken string) error

	// 通知模板管理
	CreateNotificationTemplate(ctx context.Context, template *models.NotificationTemplate) error
	GetNotificationTemplate(ctx context.Context, templateType string) (*models.NotificationTemplate, error)
	UpdateNotificationTemplate(ctx context.Context, template *models.NotificationTemplate) error
	DeleteNotificationTemplate(ctx context.Context, templateType string) error
	ListNotificationTemplates(ctx context.Context) ([]*models.NotificationTemplate, error)
	CreateOrUpdateNotificationTemplate(ctx context.Context, template *models.NotificationTemplate) error
}

// Repository 通知数据存储层
type Repository struct {
	*PostgresRepository
	logger zerolog.Logger
}

// NewRepository 创建一个新的Repository实例
func NewRepository(db *gorm.DB, logger zerolog.Logger) *Repository {
	// 创建PostgreSQL仓库实现
	postgresRepo := NewPostgresRepository(db)
	
	return &Repository{
		PostgresRepository: postgresRepo,
		logger: logger,
	}
}
