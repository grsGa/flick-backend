package main

import (
	"backend/pkg/config"
	"backend/pkg/database"
	"backend/pkg/discovery"
	"backend/pkg/logger"
	"backend/pkg/telemetry"
	"backend/services/auth/internal/repository"
	"backend/services/auth/internal/server"
	"backend/services/auth/internal/service"
	"context"
	"log"
	"strconv"

	"go.uber.org/zap"
)

const (
	serviceName = "auth-service"
)

func main() {
	// Initialize logger
	logger, err := logger.NewLogger()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Initialize tracer
	tp, err := telemetry.InitTracer(serviceName)
	if err != nil {
		logger.Fatal("Failed to init tracer", zap.Error(err))
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Error("Failed to shutdown tracer provider", zap.Error(err))
		}
	}()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// Initialize Database
	if err := database.InitDB(cfg, true); err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Initialize repository
	authRepo := repository.NewAuthRepository()

	// Initialize service
	authService := service.NewAuthService(authRepo, cfg, logger)

	// Initialize server
	grpcServer := server.NewGRPCServer(authService)

	port, err := strconv.Atoi(cfg.AuthServicePort)
	if err != nil {
		logger.Fatal("Invalid port", zap.Error(err))
	}

	// Service registration
	discovery.RegisterServiceToConsul(discovery.RegisterOptions{
		ServiceName:     serviceName,
		ServicePort:     port,
		HealthCheckType: "grpc",
	})

	logger.Info("Starting auth service", zap.Int("port", port))

	if err := grpcServer.Run(cfg.AuthServicePort); err != nil {
		logger.Fatal("Failed to run auth service", zap.Error(err))
	}
}
