package main

import (
	"log"
	"os"

	"github.com/flick/backend/services/content/internal/repository"
	"github.com/flick/backend/services/content/internal/server"
	"github.com/flick/backend/services/content/internal/service"
)

func main() {
	// 初始化仓库
	contentRepo := repository.NewContentRepository()

	// 初始化服务
	contentService := service.NewContentService(contentRepo)

	// 初始化服务端
	grpcServer := server.NewGRPCServer(contentService)

	port := os.Getenv("CONTENT_SERVICE_PORT")
	if port == "" {
		port = "50052"
	}

	log.Printf("Starting content service on port %s", port)

	if err := grpcServer.Run(port); err != nil {
		log.Fatalf("Failed to run content service: %v", err)
	}
}
