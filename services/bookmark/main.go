package main

import (
	"log"
	"os"
	
	"backend/services/bookmark/internal/server"
	"backend/services/bookmark/internal/service"
	"backend/services/bookmark/internal/repository"
)

func main() {
	// 初始化仓库
	bookmarkRepo := repository.NewBookmarkRepository()
	
	// 初始化服务
	bookmarkService := service.NewBookmarkService(bookmarkRepo)
	
	// 初始化服务端
	grpcServer := server.NewGRPCServer(bookmarkService)
	
	port := os.Getenv("BOOKMARK_SERVICE_PORT")
	if port == "" {
		port = "50060"
	}
	
	log.Printf("Starting bookmark service on port %s", port)
	
	if err := grpcServer.Run(port); err != nil {
		log.Fatalf("Failed to run bookmark service: %v", err)
	}
}