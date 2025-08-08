package database

import (
	"fmt"
	"log"

	"backend/pkg/config"
	"backend/pkg/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

// InitDB initializes the database connection. It should be called once at startup.
func InitDB(cfg *config.Config, autoMigrate bool) error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.PostgresHost, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDB, cfg.PostgresPort)

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect database: %v", err)
	}

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
