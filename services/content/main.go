package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"backend/pkg/config"
	"backend/pkg/database/postgres"
	"backend/pkg/database/redis"
	"backend/services/content/repository"
	"backend/services/content/service"
)

// 定义JWT声明结构
type JWTClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// JWT身份验证中间件
func JWTAuthMiddleware(jwtSecret string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 记录请求信息，帮助调试
			log.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Msg("收到请求")

			// 从请求头获取认证令牌
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Warn().Str("path", r.URL.Path).Msg("未授权：缺少授权头")
				http.Error(w, "未授权：缺少授权头", http.StatusUnauthorized)
				return
			}

			// 解析Bearer令牌
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				log.Warn().Str("path", r.URL.Path).Msg("未授权：无效的授权格式")
				http.Error(w, "未授权：无效的授权格式", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// 验证JWT令牌
			token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			})

			if err != nil {
				log.Warn().Err(err).Str("path", r.URL.Path).Msg("未授权：令牌无效")
				http.Error(w, "未授权："+err.Error(), http.StatusUnauthorized)
				return
			}

			if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
				// 将用户ID添加到请求上下文
				ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
				log.Info().Str("user_id", claims.UserID).Str("path", r.URL.Path).Msg("用户已认证")
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				log.Warn().Str("path", r.URL.Path).Msg("未授权：无效的令牌")
				http.Error(w, "未授权：无效的令牌", http.StatusUnauthorized)
			}
		})
	}
}

func main() {
	// 初始化日志
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	logger := log.With().Str("service", "content").Logger()

	// 加载配置
	cfg, err := config.LoadConfig("content")
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
	contentRepo := repository.NewRepository(db, logger)

	// 初始化服务
	contentService := service.NewContentService(contentRepo, redisClient, logger)

	// 设置路由
	r := mux.NewRouter()

	// 添加请求日志中间件
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 添加更多详细信息，包括请求参数和请求头
			logger.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Str("query", r.URL.RawQuery).
				Str("auth", fmt.Sprintf("%t", r.Header.Get("Authorization") != "")).
				Str("user_agent", r.Header.Get("User-Agent")).
				Msg("收到HTTP请求")

			// 对于X风格URL的请求路径，添加特殊日志
			if strings.Contains(r.URL.Path, "/status/") {
				// 解析X风格URL请求路径中的参数
				pathParts := strings.Split(r.URL.Path, "/")
				if len(pathParts) >= 4 && pathParts[2] == "status" {
					username := pathParts[1]
					permalinkID := pathParts[3]
					logger.Info().
						Str("username", username).
						Str("permalink_id", permalinkID).
						Msg("收到X风格URL访问请求")
				}
			}

			if r.URL.Path == "/content/upload" && r.Method == "POST" {
				logger.Info().Msg("收到文件上传请求")
			}

			next.ServeHTTP(w, r)
		})
	})

	// 添加JWT身份验证中间件 - 不再使用/api/v1前缀
	apiRouter := r.PathPrefix("").Subrouter()
	apiRouter.Use(JWTAuthMiddleware(cfg.JWTSecret))

	// 保留基础功能路由
	// 上传路由 - 处理文件上传
	uploadHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("content_type", r.Header.Get("Content-Type")).
			Msg("正在处理上传请求")

		// 设置较长的超时时间，处理大文件上传
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
		defer cancel()

		contentService.UploadContentHandler(w, r.WithContext(ctx))
	})
	apiRouter.Handle("/content/upload", uploadHandler).Methods("POST")

	// 添加公开帖子路由
	apiRouter.HandleFunc("/content/posts", contentService.ListContentHandler).Methods("GET")

	// Feed处理 - 添加用户Feed路由
	feedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("正在处理Feed请求")
		contentService.GetUserFeedHandler(w, r)
	})
	apiRouter.Handle("/content/feed", feedHandler).Methods("GET")

	// 添加创建帖子路由
	createPostHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("content_type", r.Header.Get("Content-Type")).
			Msg("处理创建帖子请求")
		contentService.CreatePostHandler(w, r)
	})
	apiRouter.Handle("/content/posts", createPostHandler).Methods("POST")

	// 添加永久链接路由，用于处理从网关转发的请求
	permalinkHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		username := vars["username"]
		permalinkID := vars["permalink_id"]

		logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("username", username).
			Str("permalink_id", permalinkID).
			Msg("正在处理永久链接请求")

		contentService.GetPostByPermalinkHandler(w, r)
	})
	// 添加与网关转发路径匹配的路由
	apiRouter.Handle("/content/permalink/{username}/{permalink_id}", permalinkHandler).Methods("GET")

	// 处理修改帖子的永久链接请求
	apiRouter.HandleFunc("/content/permalink/{username}/{permalink_id}", contentService.UpdatePostByPermalinkHandler).Methods("PUT")
	apiRouter.HandleFunc("/content/permalink/{username}/{permalink_id}", contentService.DeletePostByPermalinkHandler).Methods("DELETE")

	// 处理媒体文件的永久链接请求
	apiRouter.HandleFunc("/content/permalink/{username}/{permalink_id}/media/{index}", contentService.GetPostMediaByIndexHandler).Methods("GET")

	// 处理保存/取消保存的永久链接请求
	apiRouter.HandleFunc("/content/permalink/{username}/{permalink_id}/save", contentService.SavePostHandler).Methods("POST")
	apiRouter.HandleFunc("/content/permalink/{username}/{permalink_id}/save", contentService.UnsavePostHandler).Methods("DELETE")

	// X风格的API路由 - 支持直接访问，不需要网关路径转换
	// 注意：给X风格路由添加独立日志处理中间件
	xStyleGetHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		username := vars["username"]
		permalinkID := vars["permalink_id"]

		logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("username", username).
			Str("permalink_id", permalinkID).
			Msg("正在处理X风格的帖子请求")

		contentService.GetPostByPermalinkHandler(w, r)
	})

	// 直接注册X风格API路由，只在api前缀下注册
	apiRouter.Handle("/{username}/status/{permalink_id}", xStyleGetHandler).Methods("GET")

	// 添加评论公开路由，对应网关中的公开路由
	r.HandleFunc("/{username}/status/{permalink_id}/comments", contentService.GetPostCommentsHandler).Methods("GET")

	apiRouter.HandleFunc("/{username}/status/{permalink_id}", contentService.UpdatePostByPermalinkHandler).Methods("PUT")
	apiRouter.HandleFunc("/{username}/status/{permalink_id}", contentService.DeletePostByPermalinkHandler).Methods("DELETE")
	apiRouter.HandleFunc("/{username}/status/{permalink_id}/photo/{index}", contentService.GetPostMediaByIndexHandler).Methods("GET")

	// 添加X风格URL的评论相关路由
	apiRouter.HandleFunc("/{username}/status/{permalink_id}/comments", contentService.GetPostCommentsHandler).Methods("GET")
	apiRouter.HandleFunc("/{username}/status/{permalink_id}/comments", contentService.AddCommentToPostHandler).Methods("POST")
	apiRouter.HandleFunc("/{username}/status/{permalink_id}/like", contentService.LikePostHandler).Methods("POST")
	apiRouter.HandleFunc("/{username}/status/{permalink_id}/like", contentService.UnlikePostHandler).Methods("DELETE")
	apiRouter.HandleFunc("/{username}/status/{permalink_id}/save", contentService.SavePostHandler).Methods("POST")
	apiRouter.HandleFunc("/{username}/status/{permalink_id}/save", contentService.UnsavePostHandler).Methods("DELETE")

	// 添加获取用户帖子列表的路由
	getUserPostsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("正在处理获取用户帖子请求")

		vars := mux.Vars(r)
		username := vars["username"]
		logger.Info().Str("username", username).Msg("获取用户帖子")

		contentService.GetUserPostsByUsernameHandler(w, r)
	})
	// 只保留一个路由路径，避免重复
	apiRouter.Handle("/users/{username}/posts", getUserPostsHandler).Methods("GET")

	// 获取当前用户帖子的路由
	getCurrentUserPostsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("正在处理获取当前用户帖子请求")

		// 从上下文获取用户ID
		userID := ""
		if userIDVal := r.Context().Value("user_id"); userIDVal != nil {
			if id, ok := userIDVal.(string); ok {
				userID = id
			}
		}

		if userID == "" {
			http.Error(w, "未授权", http.StatusUnauthorized)
			return
		}

		logger.Info().Str("user_id", userID).Msg("获取当前用户帖子")

		// 保留原始请求的分页参数，并设置user_id参数
		query := r.URL.Query()
		query.Set("user_id", userID) // 添加user_id参数
		r.URL.RawQuery = query.Encode()

		// 使用新的通过用户ID获取帖子的处理函数
		contentService.GetUserPostsHandler(w, r)
	})
	apiRouter.Handle("/content/user", getCurrentUserPostsHandler).Methods("GET")

	// 系统健康检查 - 无需认证
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// 指标接口 - 无需认证
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
		logger.Info().Msgf("内容服务启动在端口 %s", cfg.Server.Port)
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
