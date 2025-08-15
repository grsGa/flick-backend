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

	ConsulAgentAddr string `mapstructure:"CONSUL_AGENT_ADDR"`

	PostgresUser     string `mapstructure:"POSTGRES_USER"`
	PostgresPassword string `mapstructure:"POSTGRES_PASSWORD"`
	PostgresDB       string `mapstructure:"POSTGRES_DB"`
	PostgresHost     string `mapstructure:"POSTGRES_HOST"`
	PostgresPort     string `mapstructure:"POSTGRES_PORT"`

	GithubClientID     string `mapstructure:"GITHUB_CLIENT_ID"`
	GithubClientSecret string `mapstructure:"GITHUB_CLIENT_SECRET"`
	GithubRedirectURL  string `mapstructure:"GITHUB_REDIRECT_URL"`

	GoogleClientID     string `mapstructure:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `mapstructure:"GOOGLE_CLIENT_SECRET"`
	GoogleRedirectURL  string `mapstructure:"GOOGLE_REDIRECT_URL"`

	JWTSecret        string `mapstructure:"JWT_SECRET"`
	CustomCaCertPath string `mapstructure:"CUSTOM_CA_CERT_PATH"`
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

	// OAuth defaults - these should be overridden by environment variables
	viper.SetDefault("GOOGLE_CLIENT_ID", "")
	viper.SetDefault("GOOGLE_CLIENT_SECRET", "")
	viper.SetDefault("GOOGLE_REDIRECT_URL", "")
	viper.SetDefault("GITHUB_CLIENT_ID", "")
	viper.SetDefault("GITHUB_CLIENT_SECRET", "")
	viper.SetDefault("GITHUB_REDIRECT_URL", "")

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
