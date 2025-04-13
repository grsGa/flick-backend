package postgres

import (
	"fmt"
	"time"

	"backend/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// DB 封装了GORM数据库实例
type DB struct {
	*gorm.DB
}

// NewPostgresDB 创建数据库连接
func NewPostgresDB(cfg interface{}) (*gorm.DB, error) {
	// 将传入的配置转换为Config类型
	config, ok := cfg.(*config.Config)
	if !ok {
		return nil, fmt.Errorf("无效的配置类型")
	}
	
	// 构建数据库连接字符串
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		config.DBHost, config.DBUser, config.DBPassword, config.DBName, config.DBPort, config.DBSSLMode, config.DBTimeZone,
	)
	
	// 配置GORM
	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "flick_", // 表前缀
			SingularTable: false,    // 使用复数表名
		},
		Logger: logger.Default.LogMode(getLogLevel(config.LogLevel)),
	}

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, err
	}

	// 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	
	// 连接池配置
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// 将字符串日志级别转换为GORM日志级别
func getLogLevel(logLevel string) logger.LogLevel {
	switch logLevel {
	case "debug":
		return logger.Info
	case "info":
		return logger.Info
	case "warn", "warning":
		return logger.Warn
	case "error":
		return logger.Error
	case "silent":
		return logger.Silent
	default:
		return logger.Info
	}
} 