package main

import (
	"log"
	"os"

	"github.com/flick/backend/services/notification/internal/repository"
	"github.com/flick/backend/services/notification/internal/server"
	"github.com/flick/backend/services/notification/internal/service"
)

func main() {
	// 初始化仓库
	notificationRepo := repository.NewNotificationRepository()

	// 初始化服务
	notificationService := service.NewNotificationService(notificationRepo)

	// 初始化服务端
	grpcServer := server.NewGRPCServer(notificationService)

	port := os.Getenv("NOTIFICATION_SERVICE_PORT")
	if port == "" {
		port = "50056"
	}

	log.Printf("Starting notification service on port %s", port)

	if err := grpcServer.Run(port); err != nil {
		log.Fatalf("Failed to run notification service: %v", err)
	}
}
