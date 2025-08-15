package main

import (
	"log"
	"os"

	"github.com/flick/backend/services/search/internal/repository"
	"github.com/flick/backend/services/search/internal/server"
	"github.com/flick/backend/services/search/internal/service"
)

func main() {
	// 初始化仓库
	searchRepo := repository.NewSearchRepository()

	// 初始化服务
	searchService := service.NewSearchService(searchRepo)

	// 初始化服务端
	grpcServer := server.NewGRPCServer(searchService)

	port := os.Getenv("SEARCH_SERVICE_PORT")
	if port == "" {
		port = "50058"
	}

	log.Printf("Starting search service on port %s", port)

	if err := grpcServer.Run(port); err != nil {
		log.Fatalf("Failed to run search service: %v", err)
	}
}
