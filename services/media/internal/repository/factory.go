package repository

import (
	"log"
	"os"

	"github.com/flick/backend/services/media/internal/storage"
)

// NewMediaRepositoryFactory 创建媒体仓储工厂
func NewMediaRepositoryFactory() MediaRepositoryFactory {
	return &mediaRepositoryFactory{}
}

// MediaRepositoryFactory 媒体仓储工厂接口
type MediaRepositoryFactory interface {
	Create() MediaRepository
}

// mediaRepositoryFactory 媒体仓储工厂实现
type mediaRepositoryFactory struct{}

// Create 创建媒体仓储实例
func (f *mediaRepositoryFactory) Create() MediaRepository {
	// 初始化MinIO存储配置
	minioConfig := storage.MediaStorageConfig{
		Endpoint:   getEnvOrDefault("MINIO_ENDPOINT", "minio:9000"),
		AccessKey:  getEnvOrDefault("MINIO_ACCESS_KEY", "minioadmin"),
		SecretKey:  getEnvOrDefault("MINIO_SECRET_KEY", "minioadmin"),
		UseSSL:     false, // 开发环境使用HTTP
		BucketName: "social-media", // 单一存储桶
		PublicURL:  "http://localhost:9000", // 浏览器可访问的公开URL
	}

	// 初始化MinIO客户端
	minioClient, err := storage.NewMinIOClient(
		minioConfig.Endpoint,
		minioConfig.AccessKey,
		minioConfig.SecretKey,
		minioConfig.BucketName,
	)
	if err != nil {
		log.Fatalf("Failed to initialize MinIO client: %v", err)
	}

	// 初始化存储仓库
	storageRepo := NewMinIOStorageRepository(minioClient)

	return NewMediaRepository(storageRepo)
}

// getEnvOrDefault 获取环境变量或返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
