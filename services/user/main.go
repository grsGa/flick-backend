package main

import (
	"log"
	"strconv"

	"backend/pkg/config"
	"backend/pkg/database"
	"backend/pkg/discovery"
	"backend/services/user/internal/repository"
	"backend/services/user/internal/server"
	"backend/services/user/internal/service"
)

const (
	serviceName = "user-service"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Database
	if err := database.InitDB(cfg, false); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 初始化仓库
	userRepo := repository.NewUserRepository()

	// 初始化服务
	userService := service.NewUserService(userRepo, cfg)

	// 初始化服务端
	grpcServer := server.NewGRPCServer(userService)

	port, err := strconv.Atoi(cfg.UserServicePort)
	if err != nil {
		log.Fatalf("Invalid port: %v", err)
	}

	// Service registration
	discovery.RegisterServiceToConsul(discovery.RegisterOptions{
		ServiceName:     serviceName,
		ServicePort:     port,
		HealthCheckType: "grpc",
	})

	log.Printf("Starting user service on port %d", port)

	if err := grpcServer.Run(cfg.UserServicePort); err != nil {
		log.Fatalf("Failed to run user service: %v", err)
	}
}
