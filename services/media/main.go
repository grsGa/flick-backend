package main

import (
	"log"
	"os"
	
	"backend/services/media/internal/server"
	"backend/services/media/internal/service"
	"backend/services/media/internal/repository"
)

func main() {
	// 初始化仓库
	mediaRepo := repository.NewMediaRepository()
	
	// 初始化服务
	mediaService := service.NewMediaService(mediaRepo)
	
	// 初始化服务端
	grpcServer := server.NewGRPCServer(mediaService)
	
	port := os.Getenv("MEDIA_SERVICE_PORT")
	if port == "" {
		port = "50054"
	}
	
	log.Printf("Starting media service on port %s", port)
	
	if err := grpcServer.Run(port); err != nil {
		log.Fatalf("Failed to run media service: %v", err)
	}
}