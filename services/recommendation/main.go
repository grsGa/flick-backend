package main

import (
	"log"
	"os"
	
	"backend/services/recommendation/internal/server"
	"backend/services/recommendation/internal/service"
	"backend/services/recommendation/internal/repository"
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