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
	"backend/services/user/repository"
	"backend/services/user/service"
)

func main() {
	// 初始化日志
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	logger := log.With().Str("service", "user").Logger()

	// 加载配置
	cfg, err := config.LoadConfig("user")
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
	userRepo := repository.NewRepository(db, logger)

	// 初始化服务
	userService := service.NewUserService(userRepo, redisClient, logger)

	// 设置路由
	r := mux.NewRouter()

	// API 路由 - 移除/api/v1前缀
	r.HandleFunc("/users/me", userService.GetCurrentUserHandler).Methods("GET")
	r.HandleFunc("/users/me", userService.UpdateCurrentUserHandler).Methods("PUT")
	r.HandleFunc("/users/me/follow-stats", userService.GetFollowStatsHandler).Methods("GET")
	r.HandleFunc("/users/me/avatar", userService.UploadAvatarHandler).Methods("POST", "PUT")
	r.HandleFunc("/users/me/cover-image", userService.UploadCoverImageHandler).Methods("POST", "PUT")
	r.HandleFunc("/users/{id}/followers", userService.GetUserFollowersHandler).Methods("GET")
	r.HandleFunc("/users/{id}/following", userService.GetUserFollowingHandler).Methods("GET")
	r.HandleFunc("/users/{id}/follow", userService.FollowUserHandler).Methods("POST")
	r.HandleFunc("/users/{id}/follow", userService.UnfollowUserHandler).Methods("DELETE")
	r.HandleFunc("/users/profile/{username}", userService.GetUserProfileByUsernameHandler).Methods("GET")
	r.HandleFunc("/users", userService.ListUsersHandler).Methods("GET")
	r.HandleFunc("/users", userService.CreateUserHandler).Methods("POST")
	r.HandleFunc("/users/{id}", userService.GetUserHandler).Methods("GET")
	r.HandleFunc("/users/{id}", userService.UpdateUserHandler).Methods("PUT")
	r.HandleFunc("/users/{id}", userService.DeleteUserHandler).Methods("DELETE")
	r.HandleFunc("/auth/register", userService.RegisterHandler).Methods("POST")
	r.HandleFunc("/auth/login", userService.LoginHandler).Methods("POST")
	r.HandleFunc("/auth/logout", userService.LogoutHandler).Methods("POST")
	r.HandleFunc("/auth/refresh", userService.RefreshTokenHandler).Methods("POST")

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
		logger.Info().Msgf("用户服务启动在端口 %s", cfg.Server.Port)
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
