package main

import (
	"os"

	"backend/pkg/config"
	"backend/pkg/database/postgres"
	"backend/pkg/models"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// 初始化日志
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05"})
	logger := log.With().Str("script", "init_database").Logger()

	// 加载配置
	logger.Info().Msg("加载配置...")
	cfg, err := config.LoadConfig("")
	if err != nil {
		logger.Fatal().Err(err).Msg("无法加载配置")
	}

	// 显示当前配置
	logger.Info().
		Str("数据库主机", cfg.DBHost).
		Str("数据库名称", cfg.DBName).
		Msg("当前配置")

	// 连接数据库
	logger.Info().Msg("连接PostgreSQL数据库...")
	db, err := postgres.NewPostgresDB(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("无法连接数据库")
	}

	// 检查数据库连接
	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal().Err(err).Msg("无法获取SQL连接")
	}

	err = sqlDB.Ping()
	if err != nil {
		logger.Fatal().Err(err).Msg("无法ping数据库")
	}

	logger.Info().Msg("数据库连接成功，开始初始化表...")

	// 使用GORM自动迁移创建表
	// 用户相关表
	logger.Info().Msg("迁移用户相关表...")
	err = db.AutoMigrate(
		&models.User{},
		&models.UserCredential{},
		&models.UserSession{},
		&models.Role{},
		&models.UserVerification{},
		&models.UserActivity{},
		&models.UserFollow{},
		&models.UserBlock{},
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("用户表迁移失败")
	}

	// 内容相关表
	logger.Info().Msg("迁移内容相关表...")
	err = db.AutoMigrate(
		&models.Post{},
		&models.Tag{},
		&models.Category{},
		&models.PostAuditLog{},
		&models.PostReport{},
		&models.SavedPost{},
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("内容表迁移失败")
	}

	// 交互相关表
	logger.Info().Msg("迁移交互相关表...")
	err = db.AutoMigrate(
		&models.Comment{},
		&models.CommentReport{},
		&models.CommentReply{},
		&models.Reaction{},
		&models.Mention{},
		&models.PollVote{},
		&models.Share{},
		&models.UserInteraction{},
		&models.PostLike{},
		&models.CommentLike{},
		&models.PostShare{},
		&models.Bookmark{},
		&models.BookmarkCollection{},
		&models.InteractionHistory{},
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("交互表迁移失败")
	}

	// 通知相关表
	logger.Info().Msg("迁移通知相关表...")
	err = db.AutoMigrate(
		&models.Notification{},
		&models.NotificationPreference{},
		&models.DeviceToken{},
		&models.NotificationCount{},
		&models.NotificationBatch{},
		&models.NotificationTemplate{},
		&models.UserDevice{},
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("通知表迁移失败")
	}

	// 创建管理员角色
	var adminRole models.Role
	result := db.Where("name = ?", "admin").First(&adminRole)
	if result.RowsAffected == 0 {
		logger.Info().Msg("创建管理员角色...")
		adminRole = models.Role{
			Name:        "admin",
			Description: "系统管理员角色",
		}
		if err := db.Create(&adminRole).Error; err != nil {
			logger.Error().Err(err).Msg("创建管理员角色失败")
		}
	}

	// 创建用户角色
	var userRole models.Role
	result = db.Where("name = ?", "user").First(&userRole)
	if result.RowsAffected == 0 {
		logger.Info().Msg("创建用户角色...")
		userRole = models.Role{
			Name:        "user",
			Description: "普通用户角色",
		}
		if err := db.Create(&userRole).Error; err != nil {
			logger.Error().Err(err).Msg("创建用户角色失败")
		}
	}

	logger.Info().Msg("数据库初始化完成!")
} 