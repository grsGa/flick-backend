package main

import (
	"log"
	"os"
	
	"backend/services/interaction/internal/server"
	"backend/services/interaction/internal/service"
	"backend/services/interaction/internal/repository"
)

func main() {
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