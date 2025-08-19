package main

import (
	"log"
	"os"

	"github.com/flick/backend/pkg/config"
	"github.com/flick/backend/pkg/database"
	"github.com/flick/backend/services/content/internal/repository"
	"github.com/flick/backend/services/content/internal/server"
	"github.com/flick/backend/services/content/internal/service"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库连接（只有content服务执行迁移）
	if err := database.InitDB(cfg, true); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化仓库
	postRepo := repository.NewPostRepository()

	// 初始化服务（统一使用PostService）
	postService := service.NewPostService(postRepo)

	// 初始化服务端
	grpcServer := server.NewGRPCServer(postService)

	port := os.Getenv("CONTENT_SERVICE_PORT")
	if port == "" {
		port = "50055"
	}

	log.Printf("Starting content service on port %s", port)

	if err := grpcServer.Run(port); err != nil {
		log.Fatalf("Failed to run content service: %v", err)
	}
}
