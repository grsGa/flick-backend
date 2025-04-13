package main

import (
	"fmt"
	"os"
	"strings"

	"backend/pkg/config"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// 初始化日志
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05"})
	logger := log.With().Str("script", "check_config").Logger()

	// 检查每个服务的配置
	services := []string{"user", "content", "notification", "interaction", "recommendation", "gateway"}

	for _, service := range services {
		checkServiceConfig(service, logger)
	}
}

func checkServiceConfig(serviceName string, logger zerolog.Logger) {
	logger.Info().Msgf("检查服务 %s 的配置", serviceName)

	// 加载配置
	cfg, err := config.LoadConfig(serviceName)
	if err != nil {
		logger.Error().Err(err).Msgf("无法加载 %s 服务配置", serviceName)
		return
	}

	// 打印服务配置信息
	logger.Info().
		Str("服务端口", cfg.Server.Port).
		Str("数据库主机", cfg.DBHost).
		Str("数据库名称", cfg.DBName).
		Str("数据库用户", cfg.DBUser).
		Str("数据库密码", maskPassword(cfg.DBPassword)).
		Str("数据库端口", cfg.DBPort).
		Str("Redis地址", strings.Join(cfg.RedisAddrs, ",")).
		Msg("服务配置信息")

	// 构建数据库DSN
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode, cfg.DBTimeZone,
	)
	logger.Info().Str("数据库DSN", maskDSN(dsn)).Msg("数据库连接串")
}

// 掩盖密码以避免显示敏感信息
func maskPassword(password string) string {
	if len(password) <= 2 {
		return "***"
	}
	return password[:1] + "***" + password[len(password)-1:]
}

// 掩盖DSN中的密码
func maskDSN(dsn string) string {
	parts := strings.Split(dsn, " ")
	for i, part := range parts {
		if strings.HasPrefix(part, "password=") {
			password := strings.TrimPrefix(part, "password=")
			parts[i] = "password=" + maskPassword(password)
			break
		}
	}
	return strings.Join(parts, " ")
} 