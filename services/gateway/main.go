package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/services/gateway/config"
	"backend/services/gateway/middleware"
	"backend/services/gateway/routes"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("无法加载配置")
	}

	// 设置日志级别
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(level)
	}

	// 初始化路由
	r := gin.Default()

	// 设置最大multipart表单内存限制，增加到50MB
	r.MaxMultipartMemory = 50 << 20 // 50MB

	// 设置跨域中间件
	r.Use(middleware.Cors())

	// 设置请求日志中间件
	loggerConfig := middleware.DefaultLoggerConfig()
	loggerConfig.LogRequestBody = true
	loggerConfig.LogResponseBody = true
	loggerConfig.MaxBodySize = 10 << 20 // 10MB
	r.Use(middleware.Logger(loggerConfig))

	// 设置请求ID中间件
	r.Use(middleware.RequestID())

	// 健康检查路由
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// 初始化API路由
	routes.SetupAPIRoutes(r, cfg)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: r,
	}

	// 在后台启动服务器
	go func() {
		log.Info().Msgf("服务器启动在 %s 端口", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("启动服务器失败")
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("服务器强制关闭")
	}

	log.Info().Msg("服务器已优雅关闭")
}
