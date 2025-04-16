package main

import (
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"backend/pkg/config"
	"backend/pkg/database/postgres"
	"backend/pkg/models"
)

func main() {
	// 设置日志
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: zerolog.NewConsoleWriter().Out})
	logger := log.With().Str("script", "create_test_user").Logger()

	// 加载配置
	cfg, err := config.LoadConfig("")
	if err != nil {
		logger.Fatal().Err(err).Msg("无法加载配置")
	}

	// 连接数据库
	db, err := postgres.NewPostgresDB(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("无法连接到数据库")
	}

	// 创建测试用户
	createTestUsers(db, logger)

	logger.Info().Msg("测试用户创建完成!")
}

func createTestUsers(db *gorm.DB, logger zerolog.Logger) {
	// 检查是否已存在测试管理员用户
	var adminUser models.User
	result := db.Where("username = ?", "admin").First(&adminUser)
	if result.RowsAffected == 0 {
		// 创建管理员用户
		logger.Info().Msg("创建管理员用户...")
		adminUser = models.User{
			Username:      "admin",
			Email:         "admin@flick.com",
			DisplayName:   "系统管理员",
			PhoneNumber:   "13800000000",
			AccountStatus: "active",
			VerifiedEmail: true,
			LastLogin:     &time.Time{},
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		// 设置密码
		err := adminUser.SetPassword("admin123")
		if err != nil {
			logger.Error().Err(err).Msg("管理员密码设置失败")
			return
		}

		// 保存用户
		if err := db.Create(&adminUser).Error; err != nil {
			logger.Error().Err(err).Msg("创建管理员用户失败")
			return
		}

		// 分配管理员角色
		var adminRole models.Role
		if err := db.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
			logger.Error().Err(err).Msg("获取管理员角色失败")
		} else {
			if err := db.Model(&adminUser).Association("Roles").Append(&adminRole); err != nil {
				logger.Error().Err(err).Msg("为管理员分配角色失败")
			}
		}

		// 创建通知偏好设置
		notifPref := models.NotificationPreference{
			UserID:          adminUser.ID,
			LikeEnabled:     true,
			CommentEnabled:  true,
			MentionEnabled:  true,
			FollowEnabled:   true,
			SystemEnabled:   true,
			MessageEnabled:  true,
			ShareEnabled:    true,
			EmailEnabled:    true,
			PushEnabled:     true,
			QuietHoursStart: "22:00",
			QuietHoursEnd:   "08:00",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		if err := db.Create(&notifPref).Error; err != nil {
			logger.Error().Err(err).Msg("创建管理员通知设置失败")
		}
	}

	// 检查是否已存在测试普通用户
	var testUser models.User
	result = db.Where("username = ?", "testuser").First(&testUser)
	if result.RowsAffected == 0 {
		// 创建普通测试用户
		logger.Info().Msg("创建普通测试用户...")
		testUser = models.User{
			Username:      "testuser",
			Email:         "test@flick.com",
			DisplayName:   "测试用户",
			PhoneNumber:   "13900000000",
			Bio:           "这是一个用于测试的普通用户账号",
			AccountStatus: "active",
			VerifiedEmail: true,
			LastLogin:     &time.Time{},
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		// 设置密码
		err := testUser.SetPassword("password123")
		if err != nil {
			logger.Error().Err(err).Msg("测试用户密码设置失败")
			return
		}

		// 保存用户
		if err := db.Create(&testUser).Error; err != nil {
			logger.Error().Err(err).Msg("创建测试用户失败")
			return
		}

		// 分配普通用户角色
		var userRole models.Role
		if err := db.Where("name = ?", "user").First(&userRole).Error; err != nil {
			logger.Error().Err(err).Msg("获取普通用户角色失败")
		} else {
			if err := db.Model(&testUser).Association("Roles").Append(&userRole); err != nil {
				logger.Error().Err(err).Msg("为测试用户分配角色失败")
			}
		}

		// 创建通知偏好设置
		notifPref := models.NotificationPreference{
			UserID:          testUser.ID,
			LikeEnabled:     true,
			CommentEnabled:  true,
			MentionEnabled:  true,
			FollowEnabled:   true,
			SystemEnabled:   true,
			MessageEnabled:  true,
			ShareEnabled:    true,
			EmailEnabled:    true,
			PushEnabled:     true,
			QuietHoursStart: "23:00",
			QuietHoursEnd:   "07:00",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		if err := db.Create(&notifPref).Error; err != nil {
			logger.Error().Err(err).Msg("创建测试用户通知设置失败")
		}
	}

	logger.Info().Msg("测试用户已准备就绪")
}
