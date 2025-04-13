package database

import (
	"fmt"
	"time"

	"backend/pkg/config"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// DBManager 管理数据库连接
type DBManager struct {
	WriteDB *gorm.DB
	ReadDB  *gorm.DB
	logger  *zap.Logger
}

// NewDBManager 创建数据库管理器实例
func NewDBManager() (*DBManager, error) {
	cfg := config.GetConfig()
	log := config.GetLogger()

	// 主库DSN
	writeDSN := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode, cfg.DBTimeZone,
	)

	// 读库DSN（如果配置了读库）
	var readDSN string
	if cfg.ReadDBHost != "" {
		// 使用读库配置
		readDBUser := cfg.ReadDBUser
		if readDBUser == "" {
			readDBUser = cfg.DBUser
		}
		readDBPassword := cfg.ReadDBPassword
		if readDBPassword == "" {
			readDBPassword = cfg.DBPassword
		}
		
		readDSN = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
			cfg.ReadDBHost, readDBUser, readDBPassword, cfg.DBName, cfg.ReadDBPort, cfg.DBSSLMode, cfg.DBTimeZone,
		)
	} else {
		// 如果没有配置读库，使用主库作为读库
		readDSN = writeDSN
	}

	// 配置GORM
	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "flick_", // 表前缀
			SingularTable: false,    // 使用复数表名
		},
		Logger: logger.Default.LogMode(getLogLevel(cfg.LogLevel)),
	}

	// 连接主库
	writeDB, err := gorm.Open(postgres.Open(writeDSN), gormConfig)
	if err != nil {
		log.Error("Failed to connect to write database", zap.Error(err))
		return nil, err
	}

	// 设置连接池
	sqlDB, err := writeDB.DB()
	if err != nil {
		log.Error("Failed to get database connection", zap.Error(err))
		return nil, err
	}
	
	// 连接池配置
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 连接读库
	readDB, err := gorm.Open(postgres.Open(readDSN), gormConfig)
	if err != nil {
		log.Error("Failed to connect to read database", zap.Error(err))
		return nil, err
	}

	// 设置读库连接池
	sqlReadDB, err := readDB.DB()
	if err != nil {
		log.Error("Failed to get read database connection", zap.Error(err))
		return nil, err
	}
	
	sqlReadDB.SetMaxIdleConns(10)
	sqlReadDB.SetMaxOpenConns(100)
	sqlReadDB.SetConnMaxLifetime(time.Hour)

	log.Info("Database connection established",
		zap.String("write_host", cfg.DBHost),
		zap.String("read_host", cfg.ReadDBHost),
		zap.String("database", cfg.DBName),
	)

	return &DBManager{
		WriteDB: writeDB,
		ReadDB:  readDB,
		logger:  log,
	}, nil
}

// Close 关闭数据库连接
func (dm *DBManager) Close() {
	sqlDB, err := dm.WriteDB.DB()
	if err == nil {
		_ = sqlDB.Close()
	}

	sqlReadDB, err := dm.ReadDB.DB()
	if err == nil {
		_ = sqlReadDB.Close()
	}

	dm.logger.Info("Database connections closed")
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