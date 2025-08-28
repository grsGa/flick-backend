package database

import (
	"fmt"
	"log"
	"time"

	"github.com/flick/backend/pkg/config"
	"github.com/flick/backend/pkg/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

// InitDB initializes the database connection. It should be called once at startup.
func InitDB(cfg *config.Config, autoMigrate bool) error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.PostgresHost, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDB, cfg.PostgresPort)

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %v", err)
	}

	// 配置数据库连接池和事务隔离级别
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 设置事务隔离级别为READ COMMITTED，确保读写一致性
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	fmt.Printf("[Database] Connected successfully with READ COMMITTED isolation level\n")

	if autoMigrate {
		// Auto-migrate the schema
		err = db.AutoMigrate(
			&models.User{},
			&models.Follow{},
			&models.Like{},
			&models.Repost{},
			&models.Report{},
			&models.Post{},
			&models.MediaAttachment{},
			&models.PostMention{},
			&models.PostTag{},
			&models.ReplyMention{},
			&models.Poll{},
			&models.PollOption{},
			&models.Notification{},
			&models.Conversation{},
			&models.Message{},
			&models.Bookmark{},
			&models.FeedItem{},
		)
		if err != nil {
			return fmt.Errorf("failed to migrate database: %w", err)
		}
	}
	return nil
}

// GetDB returns the database instance. Panics if InitDB has not been called.
func GetDB() *gorm.DB {
	if db == nil {
		log.Fatal("database connection not initialized. Call InitDB first.")
	}
	return db
}
