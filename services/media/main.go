package main

import (
	"log"
	"os"
	"strconv"

	"github.com/flick/backend/pkg/config"
	"github.com/flick/backend/pkg/database"
	"github.com/flick/backend/pkg/discovery"
	"github.com/flick/backend/services/media/internal/repository"
	"github.com/flick/backend/services/media/internal/server"
	"github.com/flick/backend/services/media/internal/service"
	"github.com/flick/backend/services/media/internal/storage"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库连接（启用自动迁移）
	if err := database.InitDB(cfg, true); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("Database connection initialized successfully")

	// 初始化MinIO存储配置
	minioConfig := storage.MediaStorageConfig{
		Endpoint:   os.Getenv("MINIO_ENDPOINT"),
		AccessKey:  os.Getenv("MINIO_ACCESS_KEY"),
		SecretKey:  os.Getenv("MINIO_SECRET_KEY"),
		UseSSL:     false,                         // 开发环境使用HTTP
		BucketName: "social-media",                // 单一存储桶
		PublicURL:  os.Getenv("MINIO_PUBLIC_URL"), // 从环境变量获取公开URL
	}

	// 设置默认值
	if minioConfig.Endpoint == "" {
		minioConfig.Endpoint = "minio:9000"
	}
	if minioConfig.AccessKey == "" {
		minioConfig.AccessKey = "minioadmin"
	}
	if minioConfig.SecretKey == "" {
		minioConfig.SecretKey = "minioadmin123"
	}
	if minioConfig.PublicURL == "" {
		minioConfig.PublicURL = "http://localhost:9000" // 默认值，应在docker-compose中覆盖
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
	log.Println("MinIO client initialized successfully")

	// 初始化存储仓库
	storageRepo := repository.NewMinIOStorageRepository(minioClient)

	// 初始化媒体仓库
	mediaRepo := repository.NewMediaRepository(storageRepo)

	// 初始化服务
	tempDir := os.Getenv("TEMP_DIR")
	mediaService := service.NewMediaService(mediaRepo, storageRepo, tempDir)

	// 初始化服务端
	grpcServer := server.NewGRPCServer(mediaService)

	port := os.Getenv("MEDIA_SERVICE_PORT")
	if port == "" {
		port = "50054"
	}

	// 注册服务到 Consul
	portInt, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("Invalid port: %v", err)
	}

	discovery.RegisterServiceToConsul(discovery.RegisterOptions{
		ServiceName:     "media-service",
		ServicePort:     portInt,
		HealthCheckType: "grpc",
	})

	log.Printf("Starting media service on port %s", port)

	if err := grpcServer.Run(port); err != nil {
		log.Fatalf("Failed to run media service: %v", err)
	}
}
