package main

import (
	"fmt"
	"os"

	"backend/pkg/config"
	"backend/pkg/database/postgres"
	"backend/pkg/database/redis"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// 初始化日志
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05"})
	logger := log.With().Str("script", "test_db_connection").Logger()

	// 加载配置
	logger.Info().Msg("加载配置...")
	cfg, err := config.LoadConfig("")
	if err != nil {
		logger.Fatal().Err(err).Msg("无法加载配置")
	}

	// 显示当前配置
	logger.Info().
		Str("数据库主机", cfg.DBHost).
		Str("数据库名称", cfg.DBName).
		Str("Redis地址", cfg.RedisAddrs[0]).
		Msg("当前配置")

	// 测试数据库连接
	logger.Info().Msg("测试PostgreSQL连接...")
	db, err := postgres.NewPostgresDB(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("无法连接数据库")
	}
	
	// 检查数据库连接
	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal().Err(err).Msg("无法获取SQL连接")
	}
	
	err = sqlDB.Ping()
	if err != nil {
		logger.Fatal().Err(err).Msg("无法ping数据库")
	}
	
	logger.Info().Msg("PostgreSQL连接成功!")

	// 测试Redis连接
	logger.Info().Msg("测试Redis连接...")
	redisClient, err := redis.NewRedisClient(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("无法连接Redis")
	}
	
	pong, err := redisClient.Ping(redisClient.Context()).Result()
	if err != nil {
		logger.Fatal().Err(err).Msg("无法ping Redis")
	}
	
	logger.Info().Str("响应", pong).Msg("Redis连接成功!")
	
	fmt.Println("所有测试通过，数据库和Redis连接正常!")
} 