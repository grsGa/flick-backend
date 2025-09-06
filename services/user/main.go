package main

import (
	"strconv"

	"github.com/flick/backend/pkg/config"
	"github.com/flick/backend/pkg/database"
	"github.com/flick/backend/pkg/discovery"
	"github.com/flick/backend/pkg/logger"
	"github.com/flick/backend/services/user/internal/repository"
	"github.com/flick/backend/services/user/internal/server"
	"github.com/flick/backend/services/user/internal/service"

	"go.uber.org/zap"
)

const (
	serviceName = "user-service"
)

func main() {
	// Initialize logger
	appLogger, err := logger.NewLogger()
	if err != nil {
		// Fallback to standard logger if zap fails
		zap.S().Fatalf("Failed to create logger: %v", err)
	}
	zap.ReplaceGlobals(appLogger)
	defer appLogger.Sync()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		zap.S().Fatalf("Failed to load config: %v", err)
	}

	// Initialize Database
	if err := database.InitDB(cfg, false); err != nil {
		zap.S().Fatalf("Failed to connect to database: %v", err)
	}

	// 初始化仓库
	userRepo := repository.NewUserRepository()

	// 初始化服务
	userService := service.NewUserService(userRepo, cfg, appLogger, nil)

	// 初始化服务端
	grpcServer := server.NewGRPCServer(userService)

	port, err := strconv.Atoi(cfg.UserServicePort)
	if err != nil {
		zap.S().Fatalf("Invalid port: %v", err)
	}

	// Service registration
	discovery.RegisterServiceToConsul(discovery.RegisterOptions{
		ServiceName:     serviceName,
		ServicePort:     port,
		HealthCheckType: "grpc",
	})

	zap.S().Infof("Starting user service on port %d", port)

	if err := grpcServer.Run(cfg.UserServicePort); err != nil {
		zap.S().Fatalf("Failed to run user service: %v", err)
	}
}
