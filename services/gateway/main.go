package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

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
	media_proto "github.com/flick/backend/services/media/proto"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/vektah/gqlparser/v2/gqlerror"
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
	if err := r.Run(":" + cfg.GatewayPort); err != nil {
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

func setupRoutes(r *gin.Engine) {
	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// GraphQL endpoint
	graphqlPath := "/graphql"
	queryHandler := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver.NewResolver(authServiceClient, userServiceClient, contentServiceClient)}))
	queryHandler.Use(extension.Introspection{})

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

			authHeader := c.Request.Header.Get("Authorization")
			if len(authHeader) > 20 {
				fmt.Printf("[Gateway] Authorization: %s...\n", authHeader[:20])
			} else {
				fmt.Printf("[Gateway] Authorization: %s\n", authHeader)
			}

			// The middleware already added the claims to the request context.
			// gqlgen will automatically pick it up.
			queryHandler.ServeHTTP(c.Writer, c.Request)
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

	// Media API routes
	api := r.Group("/api")
	{
		media := api.Group("/media")
		{
			// Legacy direct upload endpoint (keep for compatibility)
			media.POST("/upload", func(c *gin.Context) {
				// Debug logging for media upload
				fmt.Printf("[MEDIA] Upload request received\n")

				// Get user claims from context (set by auth middleware)
				claims := middleware.GetUserClaims(c.Request.Context())
				if claims == nil {
					fmt.Printf("[MEDIA] No user claims found in context\n")
					c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
					return
				}
				userID := claims.UserID
				fmt.Printf("[MEDIA] User authenticated: %s\n", userID)

				// Parse multipart form
				err := c.Request.ParseMultipartForm(10 << 20) // 10MB max
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
					return
				}

				file, header, err := c.Request.FormFile("file")
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
					return
				}
				defer file.Close()

				// Read file data
				fileData, err := io.ReadAll(file)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
					return
				}

				// Get file type and alt text from form
				fileType := c.PostForm("type")
				altText := c.PostForm("alt_text")

				// Call media service
				req := &media_proto.UploadFileRequest{
					UserId:   userID,
					Filename: header.Filename,
					FileData: fileData,
					Type:     fileType,
					AltText:  altText,
				}

				res, err := mediaServiceClient.UploadFile(c, req)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"url": res.File.Url,
					"id":  res.File.Id,
				})
			})
		}
	}
}
