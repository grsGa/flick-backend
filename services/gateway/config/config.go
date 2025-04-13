package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// ServiceConfig 表示微服务配置
type ServiceConfig struct {
	Name string
	URL  string
}

// Config 表示API网关的配置
type Config struct {
	Port                int
	Env                 string
	JwtSecret           string
	ReadTimeoutSeconds  int
	WriteTimeoutSeconds int
	IdleTimeoutSeconds  int
	Services            map[string]ServiceConfig
}

// LoadConfig 从环境变量加载配置
func LoadConfig() (*Config, error) {
	// 尝试加载.env文件，如果不存在则忽略
	_ = godotenv.Load()

	// 默认配置
	cfg := &Config{
		Port:                8080,
		Env:                 "development",
		JwtSecret:           "your-secret-key",
		ReadTimeoutSeconds:  10,
		WriteTimeoutSeconds: 10,
		IdleTimeoutSeconds:  60,
		Services:            make(map[string]ServiceConfig),
	}

	// 从环境变量覆盖配置
	if port := os.Getenv("PORT"); port != "" {
		portInt, err := strconv.Atoi(port)
		if err != nil {
			return nil, fmt.Errorf("无效的端口配置: %v", err)
		}
		cfg.Port = portInt
	}

	if env := os.Getenv("ENV"); env != "" {
		cfg.Env = env
	}

	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		cfg.JwtSecret = jwtSecret
	}

	if readTimeout := os.Getenv("READ_TIMEOUT_SECONDS"); readTimeout != "" {
		readTimeoutInt, err := strconv.Atoi(readTimeout)
		if err != nil {
			return nil, fmt.Errorf("无效的读取超时配置: %v", err)
		}
		cfg.ReadTimeoutSeconds = readTimeoutInt
	}

	if writeTimeout := os.Getenv("WRITE_TIMEOUT_SECONDS"); writeTimeout != "" {
		writeTimeoutInt, err := strconv.Atoi(writeTimeout)
		if err != nil {
			return nil, fmt.Errorf("无效的写入超时配置: %v", err)
		}
		cfg.WriteTimeoutSeconds = writeTimeoutInt
	}

	if idleTimeout := os.Getenv("IDLE_TIMEOUT_SECONDS"); idleTimeout != "" {
		idleTimeoutInt, err := strconv.Atoi(idleTimeout)
		if err != nil {
			return nil, fmt.Errorf("无效的空闲超时配置: %v", err)
		}
		cfg.IdleTimeoutSeconds = idleTimeoutInt
	}

	// 加载微服务配置
	// 格式 SERVICE_USER=user-service:http://localhost:8081/api/v1
	// 服务表示为name:url对
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "SERVICE_") {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) != 2 {
				continue
			}

			serviceName := strings.TrimPrefix(parts[0], "SERVICE_")
			serviceValue := parts[1]

			// 解析服务配置
			configParts := strings.SplitN(serviceValue, ":", 2)
			if len(configParts) != 2 {
				continue
			}

			cfg.Services[strings.ToLower(serviceName)] = ServiceConfig{
				Name: configParts[0],
				URL:  configParts[1],
			}
		}
	}

	// 设置默认服务配置（如果环境变量未配置）
	defaultServices := map[string]ServiceConfig{
		"user": {
			Name: "user-service",
			URL:  "http://localhost:8081/api/v1",
		},
		"content": {
			Name: "content-service",
			URL:  "http://localhost:8082/api/v1",
		},
		"interaction": {
			Name: "interaction-service",
			URL:  "http://localhost:8083/api/v1",
		},
		"notification": {
			Name: "notification-service",
			URL:  "http://localhost:8084/api/v1",
		},
		"recommendation": {
			Name: "recommendation-service",
			URL:  "http://localhost:8085/api/v1",
		},
	}

	// 合并默认配置
	for key, service := range defaultServices {
		if _, exists := cfg.Services[key]; !exists {
			cfg.Services[key] = service
		}
	}

	return cfg, nil
}
