package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// Config 应用配置
type Config struct {
	// 数据库设置
	DBHost         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBPort         string
	DBSSLMode      string
	DBTimeZone     string
	ReadDBHost     string
	ReadDBUser     string
	ReadDBPassword string
	ReadDBPort     string

	// Redis设置
	RedisAddrs    []string
	RedisPassword string

	// Kafka设置
	KafkaBrokers []string
	KafkaTopics  map[string]string
	KafkaGroupID string

	// JWT设置
	JWTSecret     string
	JWTExpiration time.Duration

	// 服务设置
	Server struct {
		Port string
	}

	// 日志设置
	LogLevel string

	// 数据库配置
	Database interface{}

	// Redis配置
	Redis interface{}

	// 其他配置
	ServiceName         string
	Environment         string
	ShutdownTimeoutSecs int

	// 其他服务URL
	UserServiceURL           string
	PostServiceURL           string
	InteractionServiceURL    string
	NotificationServiceURL   string
	RecommendationServiceURL string
	SearchServiceURL         string
	GatewayServiceURL        string
}

var (
	config *Config
	logger *zap.Logger
)

// LoadConfig 加载配置
func LoadConfig(serviceName string) (*Config, error) {
	// 加载环境变量
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}

	// 获取Redis主机和端口
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	// 创建默认配置
	config = &Config{
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBUser:         getEnv("DB_USER", "postgres"),
		DBPassword:     getEnv("DB_PASSWORD", "postgres"),
		DBName:         getEnv("DB_NAME", "flick"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBSSLMode:      getEnv("DB_SSL_MODE", "disable"),
		DBTimeZone:     getEnv("DB_TIMEZONE", "UTC"),
		ReadDBHost:     getEnv("READ_DB_HOST", ""),
		ReadDBUser:     getEnv("READ_DB_USER", ""),
		ReadDBPassword: getEnv("READ_DB_PASSWORD", ""),
		ReadDBPort:     getEnv("READ_DB_PORT", "5432"),
		RedisAddrs:     []string{redisAddr},
		RedisPassword:  getEnv("REDIS_PASSWORD", "redis123"),
		KafkaBrokers:   strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ","),
		KafkaGroupID:   getEnv("KAFKA_GROUP_ID", serviceName),
		KafkaTopics: map[string]string{
			"user":         getEnv("KAFKA_TOPIC_USER", "user-events"),
			"content":      getEnv("KAFKA_TOPIC_CONTENT", "content-events"),
			"interaction":  getEnv("KAFKA_TOPIC_INTERACTION", "interaction-events"),
			"notification": getEnv("KAFKA_TOPIC_NOTIFICATION", "notification-events"),
		},
		JWTSecret:                getEnv("JWT_SECRET", "your-secret-key"),
		JWTExpiration:            time.Duration(getEnvAsInt("JWT_EXPIRATION_HOURS", 24)) * time.Hour,
		LogLevel:                 getEnv("LOG_LEVEL", "info"),
		ServiceName:              serviceName,
		Environment:              getEnv("ENVIRONMENT", "development"),
		ShutdownTimeoutSecs:      getEnvAsInt("SHUTDOWN_TIMEOUT_SECS", 30),
		UserServiceURL:           getEnv("USER_SERVICE_URL", "http://user-service:8080"),
		PostServiceURL:           getEnv("POST_SERVICE_URL", "http://post-service:8080"),
		InteractionServiceURL:    getEnv("INTERACTION_SERVICE_URL", "http://interaction-service:8080"),
		NotificationServiceURL:   getEnv("NOTIFICATION_SERVICE_URL", "http://notification-service:8080"),
		RecommendationServiceURL: getEnv("RECOMMENDATION_SERVICE_URL", "http://recommendation-service:8080"),
		SearchServiceURL:         getEnv("SEARCH_SERVICE_URL", "http://search-service:8080"),
		GatewayServiceURL:        getEnv("GATEWAY_SERVICE_URL", "http://gateway-service:8080"),
	}

	// 服务端口设置 - 为不同服务分配不同的默认端口
	defaultPorts := map[string]string{
		"user":           "8081",
		"content":        "8082",
		"interaction":    "8083",
		"notification":   "8084",
		"recommendation": "8085",
		"gateway":        "8080",
	}

	// 首先尝试从环境变量获取特定服务的端口
	if serviceName != "" {
		portEnvVar := fmt.Sprintf("%s_PORT", strings.ToUpper(serviceName))
		servicePort := getEnv(portEnvVar, "")

		if servicePort != "" {
			// 如果环境变量中有指定端口，则使用指定的端口
			config.Server.Port = servicePort
		} else if defaultPort, exists := defaultPorts[serviceName]; exists {
			// 否则使用预定义的默认端口
			config.Server.Port = defaultPort
		} else {
			// 如果没有预定义端口，则使用通用默认端口
			config.Server.Port = getEnv("PORT", "8080")
		}
	} else {
		// 没有指定服务名称，使用通用默认端口
		config.Server.Port = getEnv("PORT", "8080")
	}

	// 设置数据库和Redis接口
	config.Database = config
	config.Redis = config

	return config, nil
}

// GetConfig 获取当前配置
func GetConfig() *Config {
	if config == nil {
		config, _ = LoadConfig("")
	}
	return config
}

// GetLogger 获取日志记录器
func GetLogger() *zap.Logger {
	if logger == nil {
		// 创建默认日志记录器
		logger, _ = zap.NewProduction()
	}
	return logger
}

// 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt 将环境变量值转换为整数
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
