package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"backend/pkg/config"
	"backend/pkg/database"
	"backend/pkg/discovery"
	"backend/pkg/logger"
	"backend/pkg/telemetry"
	auth_proto "backend/services/auth/proto"
	"backend/services/gateway/internal/client"
	"backend/services/gateway/internal/graphql/generated"
	"backend/services/gateway/internal/graphql/resolver"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/golang-lru/simplelru"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// 全局gRPC客户端变量
var (
	userServiceClient           client.UserServiceClient
	contentServiceClient        client.ContentServiceClient
	authServiceClient           client.AuthServiceClient
	mediaServiceClient          client.MediaServiceClient
	messageServiceClient        client.MessageServiceClient
	notificationServiceClient   client.NotificationServiceClient
	interactionServiceClient    client.InteractionServiceClient
	recommendationServiceClient client.RecommendationServiceClient
	searchServiceClient         client.SearchServiceClient
	bookmarkServiceClient       client.BookmarkServiceClient
)

const (
	serviceName = "gateway-service"
)

type lruCache struct {
	*simplelru.LRU
}

func (l *lruCache) Add(ctx context.Context, key string, value string) {
	l.LRU.Add(key, value)
}

func (l *lruCache) Get(ctx context.Context, key string) (string, bool) {
	val, ok := l.LRU.Get(key)
	if !ok {
		return "", false
	}
	return val.(string), true
}

func main() {
	// Initialize logger
	logger, err := logger.NewLogger()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Initialize tracer
	tp, err := telemetry.InitTracer(serviceName)
	if err != nil {
		logger.Fatal("Failed to init tracer", zap.Error(err))
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Error("Failed to shutdown tracer provider", zap.Error(err))
		}
	}()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// Set gin run mode
	gin.SetMode(gin.ReleaseMode)

	// Initialize Database
	if err := database.InitDB(cfg, false); err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Initialize gRPC clients
	initGRPCClients()

	// Create Gin engine
	r := gin.Default()

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Setup routes
	setupRoutes(r)

	port, err := strconv.Atoi(cfg.GatewayPort)
	if err != nil {
		logger.Fatal("Invalid port", zap.Error(err))
	}

	// Service registration
	discovery.RegisterServiceToConsul(discovery.RegisterOptions{
		ServiceName:     serviceName,
		ServicePort:     port,
		HealthCheckType: "http",
	})

	logger.Info("Starting gateway service", zap.String("port", cfg.GatewayPort))

	// Start server
	if err := r.Run(":" + cfg.GatewayPort); err != nil {
		logger.Fatal("Failed to run gateway service", zap.Error(err))
	}
}

// initGRPCClients initializes all gRPC client connections using service discovery.
func initGRPCClients() {
	logger := zap.L().Named("grpc_clients")

	discovery, err := client.NewConsulServiceDiscovery()
	if err != nil {
		logger.Fatal("Failed to create consul service discovery", zap.Error(err))
	}

	// A helper function to reduce boilerplate
	getClient := func(serviceName string) *grpc.ClientConn {
		conn, err := discovery.GetServiceConn(serviceName)
		if err != nil {
			logger.Fatal("Failed to get client connection", zap.String("service", serviceName), zap.Error(err))
		}
		return conn
	}

	// Initialize all service clients
	userServiceClient = client.NewUserServiceClient(getClient("user-service"))
	// contentServiceClient = client.NewContentServiceClient(getClient("content-service"))
	authServiceClient = client.NewAuthServiceClient(getClient("auth-service"))
	// mediaServiceClient = client.NewMediaServiceClient(getClient("media-service"))
	// messageServiceClient = client.NewMessageServiceClient(getClient("messages-service"))
	// notificationServiceClient = client.NewNotificationServiceClient(getClient("notification-service"))
	// interactionServiceClient = client.NewInteractionServiceClient(getClient("interaction-service"))
	// recommendationServiceClient = client.NewRecommendationServiceClient(getClient("recommendation-service"))
	// searchServiceClient = client.NewSearchServiceClient(getClient("search-service"))
	// bookmarkServiceClient = client.NewBookmarkServiceClient(getClient("bookmark-service"))
}

func setupRoutes(r *gin.Engine) {
	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// GraphQL endpoint
	graphqlPath := "/graphql"
	queryHandler := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver.NewResolver(authServiceClient, userServiceClient)}))
	queryHandler.Use(extension.Introspection{})
	l, _ := simplelru.NewLRU(100, nil)
	queryHandler.Use(extension.AutomaticPersistedQuery{
		Cache: &lruCache{l},
	})

	// GraphQL路由
	graphql := r.Group(graphqlPath)
	{
		graphql.POST("", func(c *gin.Context) {
			queryHandler.ServeHTTP(c.Writer, c.Request)
		})
	}

	// Playground
	r.GET("/", func(c *gin.Context) {
		playground.Handler("GraphQL playground", graphqlPath).ServeHTTP(c.Writer, c.Request)
	})

	// 其他API路由可以在这里添加
	auth := r.Group("/auth")
	{
		auth.GET("/github/login", func(c *gin.Context) {
			res, err := authServiceClient.GithubLogin(c, &auth_proto.GithubLoginRequest{})
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.Redirect(http.StatusTemporaryRedirect, res.RedirectUrl)
		})

		auth.GET("/github/callback", func(c *gin.Context) {
			code := c.Query("code")
			res, err := authServiceClient.GithubCallback(c, &auth_proto.GithubCallbackRequest{Code: code})
			if err != nil {
				// Redirect to an error page on the frontend
				c.Redirect(http.StatusTemporaryRedirect, "http://localhost:3000/login?error=github_failed")
				return
			}

			// On success, redirect to a frontend callback page with the token and user info
			userJSON, _ := json.Marshal(res.User)
			redirectURL := fmt.Sprintf("http://localhost:3000/auth/callback?token=%s&user=%s", res.AccessToken, url.QueryEscape(string(userJSON)))
			c.Redirect(http.StatusTemporaryRedirect, redirectURL)
		})

		auth.GET("/google/login", func(c *gin.Context) {
			res, err := authServiceClient.GoogleLogin(c, &auth_proto.GoogleLoginRequest{})
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.Redirect(http.StatusTemporaryRedirect, res.RedirectUrl)
		})

		auth.GET("/google/callback", func(c *gin.Context) {
			code := c.Query("code")
			res, err := authServiceClient.GoogleCallback(c, &auth_proto.GoogleCallbackRequest{Code: code})
			if err != nil {
				// Redirect to an error page on the frontend
				c.Redirect(http.StatusTemporaryRedirect, "http://localhost:3000/login?error=google_failed")
				return
			}

			// On success, redirect to a frontend callback page with the token and user info
			userJSON, _ := json.Marshal(res.User)
			redirectURL := fmt.Sprintf("http://localhost:3000/auth/callback?token=%s&user=%s", res.AccessToken, url.QueryEscape(string(userJSON)))
			c.Redirect(http.StatusTemporaryRedirect, redirectURL)
		})
	}
}
