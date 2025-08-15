package main

import (
	"log"
	"os"

	"github.com/flick/backend/services/messages/internal/repository"
	"github.com/flick/backend/services/messages/internal/server"
	"github.com/flick/backend/services/messages/internal/service"
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
