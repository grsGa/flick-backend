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
	"backend/services/recommendation/repository"
	"backend/services/recommendation/service"
)

func main() {
	// 初始化日志
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	logger := log.With().Str("service", "recommendation").Logger()

	// 加载配置
	cfg, err := config.LoadConfig("recommendation")
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
	recommendationRepo := repository.NewRepository(db, logger)

	// 初始化服务
	recommendationService := service.NewRecommendationService(recommendationRepo, redisClient, logger)

	// 设置路由
	r := mux.NewRouter()

	// API 路由
	r.HandleFunc("recommendations", recommendationService.GetRecommendationsHandler).Methods("GET")
	r.HandleFunc("recommendations/feedback", recommendationService.RecordFeedbackHandler).Methods("POST")
	r.HandleFunc("recommendations/viewed", recommendationService.MarkAsViewedHandler).Methods("POST")
	r.HandleFunc("recommendations/clicked", recommendationService.MarkAsClickedHandler).Methods("POST")

	// 模型相关路由
	r.HandleFunc("recommendation-models", recommendationService.ListModelsHandler).Methods("GET")
	r.HandleFunc("recommendation-models", recommendationService.CreateModelHandler).Methods("POST")
	r.HandleFunc("recommendation-models/{id}", recommendationService.UpdateModelHandler).Methods("PUT")
	r.HandleFunc("recommendation-models/{id}", recommendationService.DeleteModelHandler).Methods("DELETE")

	// A/B测试相关路由
	r.HandleFunc("recommendation-abtests", recommendationService.ListABTestsHandler).Methods("GET")
	r.HandleFunc("recommendation-abtests", recommendationService.CreateABTestHandler).Methods("POST")
	r.HandleFunc("recommendation-abtests/{id}/metrics", recommendationService.GetABTestMetricsHandler).Methods("GET")
	r.HandleFunc("recommendation-abtests/{id}", recommendationService.UpdateABTestHandler).Methods("PUT")
	r.HandleFunc("recommendation-abtests/{id}", recommendationService.DeleteABTestHandler).Methods("DELETE")

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
		logger.Info().Msgf("推荐服务启动在端口 %s", cfg.Server.Port)
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
