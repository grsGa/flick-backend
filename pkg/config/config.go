package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	GatewayPort           string `mapstructure:"GATEWAY_PORT"`
	AuthServicePort       string `mapstructure:"AUTH_SERVICE_PORT"`
	UserServicePort       string `mapstructure:"USER_SERVICE_PORT"`
	ContentServicePort    string `mapstructure:"CONTENT_SERVICE_PORT"`
	MediaServicePort      string `mapstructure:"MEDIA_SERVICE_PORT"`
	MessagesServicePort   string `mapstructure:"MESSAGES_SERVICE_PORT"`
	NotificationSvcPort   string `mapstructure:"NOTIFICATION_SERVICE_PORT"`
	InteractionSvcPort    string `mapstructure:"INTERACTION_SERVICE_PORT"`
	RecommendationSvcPort string `mapstructure:"RECOMMENDATION_SERVICE_PORT"`
	SearchServicePort     string `mapstructure:"SEARCH_SERVICE_PORT"`
	BookmarkServicePort   string `mapstructure:"BOOKMARK_SERVICE_PORT"`

	ConsulAgentAddr string `mapstructure:"CONSUL_AGENT_ADDR"`

	PostgresUser     string `mapstructure:"POSTGRES_USER"`
	PostgresPassword string `mapstructure:"POSTGRES_PASSWORD"`
	PostgresDB       string `mapstructure:"POSTGRES_DB"`
	PostgresHost     string `mapstructure:"POSTGRES_HOST"`
	PostgresPort     string `mapstructure:"POSTGRES_PORT"`

	GithubClientID     string `mapstructure:"GITHUB_CLIENT_ID"`
	GithubClientSecret string `mapstructure:"GITHUB_CLIENT_SECRET"`

	GoogleClientID     string `mapstructure:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `mapstructure:"GOOGLE_CLIENT_SECRET"`

	JWTSecret string `mapstructure:"JWT_SECRET"`
}

// LoadConfig loads the configuration from environment variables.
func LoadConfig() (*Config, error) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set default values
	viper.SetDefault("GATEWAY_PORT", "8080")
	viper.SetDefault("AUTH_SERVICE_PORT", "50053")
	viper.SetDefault("USER_SERVICE_PORT", "50051")
	// ... set defaults for other services
	viper.SetDefault("GOOGLE_CLIENT_ID", "29310918760-nbaemcllg41b7r9g9dbi3mvfvmvmt59f.apps.googleusercontent.com")
	viper.SetDefault("GOOGLE_CLIENT_SECRET", "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")

	// Database defaults
	viper.SetDefault("POSTGRES_HOST", "postgres")
	viper.SetDefault("JWT_SECRET", "your_secret_key")
	viper.SetDefault("POSTGRES_PORT", "5432")
	viper.SetDefault("POSTGRES_USER", "user")
	viper.SetDefault("POSTGRES_PASSWORD", "password")
	viper.SetDefault("POSTGRES_DB", "flick")

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
