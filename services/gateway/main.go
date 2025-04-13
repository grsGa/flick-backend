package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/services/gateway/config"
	"backend/services/gateway/middleware"
	"backend/services/gateway/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 设置Gin模式
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化Router
	router := gin.New()

	// 应用全局中间件
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(middleware.DefaultLoggerConfig()))
	router.Use(middleware.Cors())

	// 健康检查路由
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// 初始化API路由
	routes.SetupAPIRoutes(router, cfg)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeoutSeconds) * time.Second,
		IdleTimeout:  time.Duration(cfg.IdleTimeoutSeconds) * time.Second,
	}

	// 启动服务器（非阻塞）
	go func() {
		log.Printf("服务器启动在 %s 端口", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("监听失败: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("服务器强制关闭:", err)
	}

	log.Println("服务器已优雅关闭")
}
