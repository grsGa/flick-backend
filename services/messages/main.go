package main

import (
	"log"
	"os"
	
	"backend/services/messages/internal/server"
	"backend/services/messages/internal/service"
	"backend/services/messages/internal/repository"
)

func main() {
	// 初始化仓库
	messageRepo := repository.NewMessageRepository()
	
	// 初始化服务
	messageService := service.NewMessageService(messageRepo)
	
	// 初始化服务端
	grpcServer := server.NewGRPCServer(messageService)
	
	port := os.Getenv("MESSAGES_SERVICE_PORT")
	if port == "" {
		port = "50055"
	}
	
	log.Printf("Starting messages service on port %s", port)
	
	if err := grpcServer.Run(port); err != nil {
		log.Fatalf("Failed to run messages service: %v", err)
	}
}