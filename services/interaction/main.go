package main

import (
	"log"
	"os"

	"github.com/flick/backend/pkg/config"
	"github.com/flick/backend/pkg/database"
	"github.com/flick/backend/services/interaction/internal/repository"
	"github.com/flick/backend/services/interaction/internal/server"
	"github.com/flick/backend/services/interaction/internal/service"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库连接，但不执行迁移（避免多服务冲突）
	if err := database.InitDB(cfg, false); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化仓库
	interactionRepo := repository.NewInteractionRepository()

	// 初始化服务
	interactionService := service.NewInteractionService(interactionRepo)

	// 初始化服务端
	grpcServer := server.NewGRPCServer(interactionService)

	port := os.Getenv("INTERACTION_SERVICE_PORT")
	if port == "" {
		port = "50059"
	}

	log.Printf("Starting interaction service on port %s", port)

	if err := grpcServer.Run(port); err != nil {
		log.Fatalf("Failed to run interaction service: %v", err)
	}
}
