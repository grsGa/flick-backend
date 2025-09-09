package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/flick/backend/pkg/config"
	"github.com/flick/backend/pkg/database"
	"github.com/flick/backend/pkg/discovery"
	"github.com/flick/backend/pkg/logger"
	"github.com/flick/backend/pkg/telemetry"
	auth_proto "github.com/flick/backend/services/auth/proto"
	"github.com/flick/backend/services/gateway/internal/client"
	"github.com/flick/backend/services/gateway/internal/graphql/generated"
	"github.com/flick/backend/services/gateway/internal/graphql/resolver"
	"github.com/flick/backend/services/gateway/internal/middleware"
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
		fmt.Printf("Failed to create logger: %v", err)
		return
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

	// Initialize service discovery
	serviceDiscovery, err := client.NewConsulServiceDiscovery()
	if err != nil {
		logger.Fatal("Failed to create consul service discovery", zap.Error(err))
	}

	// Initialize gRPC clients with retry logic
	authClient := client.NewAuthServiceClient(getServiceConnWithRetry(serviceDiscovery, "auth-service", logger))
	userClient := client.NewUserServiceClient(getServiceConnWithRetry(serviceDiscovery, "user-service", logger))
	contentClient := client.NewContentServiceClient(getServiceConnWithRetry(serviceDiscovery, "content-service", logger))
	mediaClient := client.NewMediaServiceClient(getServiceConnWithRetry(serviceDiscovery, "media-service", logger))
	interactionClient := client.NewInteractionServiceClient(getServiceConnWithRetry(serviceDiscovery, "interaction-service", logger))

	// Set gin run mode
	gin.SetMode(gin.ReleaseMode)

	// Initialize Database
	if err := database.InitDB(cfg, false); err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Store clients globally for use in routes
	authServiceClient = authClient
	userServiceClient = userClient
	contentServiceClient = contentClient
	mediaServiceClient = mediaClient
	interactionServiceClient = interactionClient

	// Create Gin engine
	r := gin.Default()

	// Set multipart memory limit to 100MB (matching media service limit)
	r.MaxMultipartMemory = 100 << 20 // 100MB

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Auth middleware
	r.Use(middleware.AuthMiddleware(cfg))

	// Setup routes
	setupRoutes(r)

	// Service registration
	port, err := strconv.Atoi(cfg.GatewayPort)
	if err != nil {
		logger.Fatal("Invalid port", zap.Error(err))
	}

	discovery.RegisterServiceToConsul(discovery.RegisterOptions{
		ServiceName:     serviceName,
		ServicePort:     port,
		HealthCheckType: "http",
	})

	logger.Info("Starting gateway service", zap.String("port", cfg.GatewayPort))

	// Start server
	fmt.Printf("[Gateway] Starting HTTP server on port %s\n", cfg.GatewayPort)
	if err := r.Run("0.0.0.0:" + cfg.GatewayPort); err != nil {
		logger.Fatal("Failed to run gateway service", zap.Error(err))
	}
}

// getServiceConnWithRetry is a helper function to get service connection with retry logic
func getServiceConnWithRetry(serviceDiscovery client.ServiceDiscovery, serviceName string, logger *zap.Logger) *grpc.ClientConn {
	maxRetries := 10
	baseDelay := 2 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		conn, err := serviceDiscovery.GetServiceConn(serviceName)
		if err == nil {
			logger.Info("Successfully connected to service", zap.String("service", serviceName), zap.Int("attempt", attempt))
			return conn
		}

		if attempt == maxRetries {
			logger.Fatal("Failed to get client connection after all retries",
				zap.String("service", serviceName),
				zap.Int("attempts", maxRetries),
				zap.Error(err))
		}

		delay := time.Duration(attempt) * baseDelay
		logger.Warn("Failed to connect to service, retrying...",
			zap.String("service", serviceName),
			zap.Int("attempt", attempt),
			zap.Duration("retry_in", delay),
			zap.Error(err))

		time.Sleep(delay)
	}

	return nil // This should never be reached due to Fatal above
}

// getServiceConn is a helper function to get service connection with error handling
func getServiceConn(serviceDiscovery client.ServiceDiscovery, serviceName string, logger *zap.Logger) *grpc.ClientConn {
	conn, err := serviceDiscovery.GetServiceConn(serviceName)
	if err != nil {
		logger.Fatal("Failed to get client connection", zap.String("service", serviceName), zap.Error(err))
	}
	return conn
}

// responseCapture captures HTTP response for logging
type responseCapture struct {
	gin.ResponseWriter
	statusCode int
	body       []byte
}

func (w *responseCapture) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseCapture) Write(data []byte) (int, error) {
	w.body = append(w.body, data...)
	return w.ResponseWriter.Write(data)
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

	// Create GraphQL server with proper transport configuration
	// IMPORTANT: Use handler.New() and add transports manually to avoid conflicts
	srv := handler.New(generated.NewExecutableSchema(generated.Config{Resolvers: resolver.NewResolver(authServiceClient, userServiceClient, contentServiceClient, interactionServiceClient, mediaServiceClient, nil)}))

	// Configure multipart upload transport FIRST with proper configuration
	srv.AddTransport(&transport.MultipartForm{
		MaxUploadSize: 100 << 20, // 100MB max file size
		MaxMemory:     32 << 20,  // 32MB in memory, rest goes to temp files
	})

	// Add other transports in correct order
	srv.AddTransport(&transport.POST{})
	srv.AddTransport(&transport.GET{})
	srv.AddTransport(&transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
	})

	srv.Use(extension.Introspection{})
	queryHandler := srv

	// Add error handling
	queryHandler.SetErrorPresenter(func(ctx context.Context, e error) *gqlerror.Error {
		fmt.Printf("[Gateway] GraphQL Error: %v\n", e)
		return graphql.DefaultErrorPresenter(ctx, e)
	})

	queryHandler.SetRecoverFunc(func(ctx context.Context, err interface{}) error {
		fmt.Printf("[Gateway] GraphQL Panic: %v\n", err)
		return fmt.Errorf("internal server error")
	})

	// GraphQL路由
	graphql := r.Group(graphqlPath)
	{
		graphql.POST("", func(c *gin.Context) {
			fmt.Printf("[Gateway] GraphQL request received: %s %s\n", c.Request.Method, c.Request.URL.Path)
			fmt.Printf("[Gateway] Content-Type: %s\n", c.Request.Header.Get("Content-Type"))
			fmt.Printf("[Gateway] Content-Length: %s\n", c.Request.Header.Get("Content-Length"))

			authHeader := c.Request.Header.Get("Authorization")
			if len(authHeader) > 20 {
				fmt.Printf("[Gateway] Authorization: %s...\n", authHeader[:20])
			} else {
				fmt.Printf("[Gateway] Authorization: %s\n", authHeader)
			}

			// Check if this is a multipart request
			contentType := c.Request.Header.Get("Content-Type")
			if strings.Contains(contentType, "multipart/form-data") {
				fmt.Printf("[Gateway] MULTIPART REQUEST DETECTED\n")
				fmt.Printf("[Gateway] Content-Type: %s\n", contentType)
				fmt.Printf("[Gateway] Request body size: %d bytes\n", c.Request.ContentLength)
				fmt.Printf("[Gateway] Request method: %s\n", c.Request.Method)
				// Don't parse multipart form here - let gqlgen handle it
			}

			// The middleware already added the claims to the request context.
			// gqlgen will automatically pick it up.
			fmt.Printf("[Gateway] Calling GraphQL handler\n")

			// Capture response to log errors
			responseWriter := &responseCapture{ResponseWriter: c.Writer}
			queryHandler.ServeHTTP(responseWriter, c.Request)

			// Log response details for multipart requests
			if strings.Contains(contentType, "multipart/form-data") {
				fmt.Printf("[Gateway] MULTIPART RESPONSE STATUS: %d\n", responseWriter.statusCode)
				if len(responseWriter.body) > 0 {
					fmt.Printf("[Gateway] MULTIPART RESPONSE BODY: %s\n", string(responseWriter.body))
				}
			}

			fmt.Printf("[Gateway] GraphQL handler completed\n")
		})
	}

	l, _ := simplelru.NewLRU(100, nil)
	queryHandler.Use(extension.AutomaticPersistedQuery{
		Cache: &lruCache{l},
	})

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

	// REST API endpoints removed - all media operations now use GraphQL mutations
	// This ensures consistent authentication and unified API interface
}
