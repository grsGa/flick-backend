package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"backend/pkg/config"
	"backend/pkg/database/postgres"
	"backend/pkg/database/redis"
	"backend/services/interaction/repository"
	"backend/services/interaction/service"
)

func main() {
	// 初始化日志
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	logger := log.With().Str("service", "interaction").Logger()

	// 加载配置
	cfg, err := config.LoadConfig("interaction")
	if err != nil {
		logger.Fatal().Err(err).Msg("无法加载配置")
	}

	// 初始化数据库连接
	db, err := postgres.NewPostgresDB(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("无法连接数据库")
	}

	// 初始化Redis客户端
	redisClient, err := redis.NewRedisClient(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("无法连接Redis")
	}

	// 初始化存储库
	interactionRepo := repository.NewRepository(db, logger)

	// 初始化服务
	interactionService := service.NewInteractionService(interactionRepo, redisClient, logger)

	// 设置路由
	r := mux.NewRouter()

	// API 路由
	r.HandleFunc("/interactions", interactionService.CreateInteractionHandler).Methods("POST")
	r.HandleFunc("/interactions", interactionService.ListInteractionsHandler).Methods("GET")
	r.HandleFunc("/interactions/{id}", interactionService.GetInteractionHandler).Methods("GET")
	r.HandleFunc("/interactions/{id}", interactionService.UpdateInteractionHandler).Methods("PUT")
	r.HandleFunc("/interactions/{id}", interactionService.DeleteInteractionHandler).Methods("DELETE")

	// 点赞相关路由
	r.HandleFunc("/posts/{id}/like", interactionService.LikePostHandler).Methods("POST")
	r.HandleFunc("/posts/{id}/unlike", interactionService.UnlikePostHandler).Methods("POST")
	r.HandleFunc("/posts/{id}/like", interactionService.UnlikePostHandler).Methods("DELETE")
	r.HandleFunc("/posts/{id}/likes", interactionService.GetPostLikesHandler).Methods("GET")
	r.HandleFunc("/posts/{id}/is-liked", interactionService.IsPostLikedHandler).Methods("GET")
	r.HandleFunc("/users/{id}/liked-posts", interactionService.GetUserLikedPostsHandler).Methods("GET")

	// 特别添加X风格URL的路由，直接处理permalink_id
	r.HandleFunc("/{username}/status/{permalink_id}/like", interactionService.LikePostHandler).Methods("POST")
	r.HandleFunc("/{username}/status/{permalink_id}/like", interactionService.UnlikePostHandler).Methods("DELETE")
	r.HandleFunc("/{username}/status/{permalink_id}/unlike", interactionService.UnlikePostHandler).Methods("POST")

	// 书签/收藏相关路由
	r.HandleFunc("/posts/{id}/save", interactionService.SavePostHandler).Methods("POST")
	r.HandleFunc("/posts/{id}/save", interactionService.UnsavePostHandler).Methods("DELETE")
	r.HandleFunc("/posts/{id}/is-saved", interactionService.IsPostBookmarkedHandler).Methods("GET")
	r.HandleFunc("/users/{id}/bookmarks", interactionService.GetUserBookmarksHandler).Methods("GET")

	// 特别添加X风格URL的收藏路由
	r.HandleFunc("/{username}/status/{permalink_id}/save", interactionService.SavePostHandler).Methods("POST")
	r.HandleFunc("/{username}/status/{permalink_id}/save", interactionService.UnsavePostHandler).Methods("DELETE")

	// 系统健康检查
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// 指标接口
	r.Handle("/metrics", promhttp.Handler())

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务器
	go func() {
		logger.Info().Msgf("交互服务启动在端口 %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				logger.Fatal().Err(err).Msg("服务器启动失败")
			}
		}
	}()

	// 优雅关闭
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// 创建关闭超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 关闭服务器
	logger.Info().Msg("正在关闭服务器...")
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal().Err(err).Msg("服务器关闭失败")
	}

	logger.Info().Msg("服务器已关闭")
}
