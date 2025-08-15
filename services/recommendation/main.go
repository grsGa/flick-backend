package main

import (
	"log"
	"os"

	"github.com/flick/backend/services/recommendation/internal/repository"
	"github.com/flick/backend/services/recommendation/internal/server"
	"github.com/flick/backend/services/recommendation/internal/service"
)

func main() {
	// 初始化仓库
	recommendationRepo := repository.NewRecommendationRepository()

	// 初始化服务
	recommendationService := service.NewRecommendationService(recommendationRepo)

	// 初始化服务端
	grpcServer := server.NewGRPCServer(recommendationService)

	port := os.Getenv("RECOMMENDATION_SERVICE_PORT")
	if port == "" {
		port = "50057"
	}

	log.Printf("Starting recommendation service on port %s", port)

	if err := grpcServer.Run(port); err != nil {
		log.Fatalf("Failed to run recommendation service: %v", err)
	}
}
