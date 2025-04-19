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
	// 强制刷新所有输出，确保Docker环境中也能看到日志
	// 这会使日志直接写入文件描述符而不是缓冲
	defer os.Stdout.Sync()
	defer os.Stderr.Sync()

	// 配置日志输出 - 使用更可靠的配置方式
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	
	// 日志输出到标准输出，确保Docker能捕获
	log.Logger = zerolog.New(os.Stdout).With().
		Timestamp().
		Caller().
		Logger()
	
	// 记录启动日志 - 使用fmt直接输出，确保即使zerolog有问题也能看到
	startMsg := "网关服务初始化中...[" + time.Now().Format(time.RFC3339) + "]"
	fmt.Println(startMsg)
	log.Info().Msg(startMsg)

	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		errMsg := fmt.Sprintf("无法加载配置: %v", err)
		fmt.Println(errMsg)
		log.Fatal().Err(err).Msg(errMsg)
		// 确保Fatal消息能被看到
		os.Exit(1)
	}

	// 设置日志级别
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Warn().Err(err).Msg("无效的日志级别配置，使用默认级别：Info")
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(level)
		log.Info().Str("level", level.String()).Msg("设置日志级别")
	}

	// 初始化路由 - 使用 New() 而不是 Default()
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// 手动添加Recovery中间件
	r.Use(gin.Recovery())
	
	// 添加zerolog集成的中间件
	r.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		
		c.Next()
		
		end := time.Now()
		latency := end.Sub(start)
		
		if raw != "" {
			path = path + "?" + raw
		}
		
		// 使用zerolog记录请求信息
		logger := log.With().
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", c.Writer.Status()).
			Str("ip", c.ClientIP()).
			Dur("latency", latency).
			Str("user_agent", c.Request.UserAgent()).
			Logger()
		
		msg := fmt.Sprintf("%s %s %d", c.Request.Method, path, c.Writer.Status())
		if c.Writer.Status() >= 400 {
			logger.Warn().Msg(msg)
		} else {
			logger.Info().Msg(msg)
		}
	})

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
		log.Info().Msg("健康检查请求")
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

	// 服务器启动日志 - 直接输出确保可见
	serverStartMsg := fmt.Sprintf("服务器启动在 %s 端口", srv.Addr)
	fmt.Println(serverStartMsg)
	log.Info().Msg(serverStartMsg)

	// 在后台启动服务器
	go func() {
		// 在goroutine内再次输出启动消息，确保在Docker中可见
		fmt.Println(fmt.Sprintf("HTTP服务正在监听 %s...", srv.Addr))
		log.Info().Msgf("HTTP服务正在监听 %s...", srv.Addr)
		
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errMsg := fmt.Sprintf("启动服务器失败: %v", err)
			fmt.Println(errMsg)
			log.Fatal().Err(err).Msg(errMsg)
			// 确保goroutine中的Fatal消息被看到
			os.Exit(1)
		}
	}()

	// 输出等待关闭信号的日志
	log.Info().Msg("服务器运行中，等待关闭信号...")
	fmt.Println("服务器运行中，等待关闭信号...")

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	shutdownMsg := "正在关闭服务器..."
	log.Info().Msg(shutdownMsg)
	fmt.Println(shutdownMsg)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		errMsg := fmt.Sprintf("服务器强制关闭: %v", err)
		fmt.Println(errMsg)
		log.Fatal().Err(err).Msg(errMsg)
		os.Exit(1)
	}

	// 确保关闭消息被看到
	closeMsg := "服务器已优雅关闭"
	log.Info().Msg(closeMsg)
	fmt.Println(closeMsg)
	
	// 强制刷新所有输出流
	os.Stdout.Sync()
	os.Stderr.Sync()
}
