package repository

import (
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// RepositoryImpl 推荐服务的数据存储层实现
type RepositoryImpl struct {
	*PostgresRepository
	logger zerolog.Logger
}

// NewRepository 创建一个新的Repository接口实现
func NewRepository(db *gorm.DB, logger zerolog.Logger) Repository {
	// 创建PostgreSQL仓库实现
	postgresRepo := NewPostgresRepository(db)
	
	return &RepositoryImpl{
		PostgresRepository: postgresRepo,
		logger: logger,
	}
} 