package routes

import (
	"backend/services/gateway/config"
	"backend/services/gateway/handlers"
	"backend/services/gateway/middleware"

	"net/http"

	"github.com/gin-gonic/gin"
)

// AddRoute 添加自定义路由并处理可能包含参数的路径
func AddRoute(router *gin.Engine, method string, path string, handler gin.HandlerFunc) {
	// 不再添加API路径前缀，由调用者负责提供完整路径
	apiPath := path

	// 根据HTTP方法注册路由
	switch method {
	case http.MethodGet:
		router.GET(apiPath, handler)
	case http.MethodPost:
		router.POST(apiPath, handler)
	case http.MethodPut:
		router.PUT(apiPath, handler)
	case http.MethodDelete:
		router.DELETE(apiPath, handler)
	case http.MethodPatch:
		router.PATCH(apiPath, handler)
	case http.MethodHead:
		router.HEAD(apiPath, handler)
	case http.MethodOptions:
		router.OPTIONS(apiPath, handler)
	default:
		// 默认为GET
		router.GET(apiPath, handler)
	}
}

// SetupAPIRoutes 配置API路由
func SetupAPIRoutes(router *gin.Engine, cfg *config.Config) {
	// 创建API组
	api := router.Group("")

	// 创建处理程序实例
	userHandler := handlers.NewUserHandler(cfg)
	contentHandler := handlers.NewContentHandler(cfg)
	interactionHandler := handlers.NewInteractionHandler(cfg)
	notificationHandler := handlers.NewNotificationHandler(cfg)
	recommendationHandler := handlers.NewRecommendationHandler(cfg)

	// <无需认证的路由>
	// 用户认证相关
	api.POST("/auth/register", userHandler.Register)
	api.POST("/auth/login", userHandler.Login)
	api.POST("/auth/refresh", userHandler.RefreshToken)
	api.POST("/auth/forgot-password", userHandler.ForgotPassword)
	api.POST("/auth/reset-password", userHandler.ResetPassword)
	api.POST("/auth/verify-email", userHandler.VerifyEmail)

	// 内容相关的公开路由
	api.GET("/content/posts", contentHandler.GetPublicPosts)
	// api.GET("/content/posts/:id", contentHandler.GetPostByID)

	// 支持X平台风格的URL结构 - (公开路由)
	api.GET("/:username/status/:permalink_id", contentHandler.GetPostByPermalink)                      // 例如：/testuser123/status/185747420474
	api.GET("/:username/status/:permalink_id/photo/:index", contentHandler.GetPostMediaByIndex)        // 例如：/testuser123/status/185747420474/photo/1
	api.GET("/:username/status/:permalink_id/comments", interactionHandler.GetPostCommentsByPermalink) // 公开获取帖子评论

	// 用户资料公开路由
	api.GET("/users/profile/:username", userHandler.GetUserProfileByUsername) // 根据用户名获取用户资料
	api.GET("/users/:username/posts", contentHandler.GetUserPosts)            // 获取用户的帖子列表

	// <需要认证的路由>
	authRoutes := api.Group("")
	authRoutes.Use(middleware.JWTAuth(cfg.JwtSecret))

	// 用户相关路由
	authRoutes.GET("/users/me", userHandler.GetCurrentUser)
	authRoutes.PUT("/users/me", userHandler.UpdateCurrentUser)
	authRoutes.GET("/users/by-id/:id", userHandler.GetUserByID)
	authRoutes.GET("/users/by-id/:id/followers", userHandler.GetUserFollowers)
	authRoutes.GET("/users/by-id/:id/following", userHandler.GetUserFollowing)
	authRoutes.POST("/users/by-id/:id/follow", userHandler.FollowUser)
	authRoutes.DELETE("/users/by-id/:id/follow", userHandler.UnfollowUser)
	authRoutes.POST("/users/by-id/:id/block", userHandler.BlockUser)
	authRoutes.DELETE("/users/by-id/:id/block", userHandler.UnblockUser)
	authRoutes.GET("/users/search", userHandler.SearchUsers)

	// 内容相关路由 - (X风格URL)
	// 基础内容功能
	authRoutes.POST("/content/posts", contentHandler.CreatePost)        // 创建帖子
	authRoutes.GET("/content/feed", contentHandler.GetUserFeed)         // 获取用户 Feed
	authRoutes.POST("/content/upload", contentHandler.UploadContent)    // 上传内容
	authRoutes.GET("/content/user", contentHandler.GetCurrentUserPosts) // 获取当前用户的帖子
	// 支持X平台风格的需要认证的URL操作
	authRoutes.PUT("/:username/status/:permalink_id", contentHandler.UpdatePostByPermalink)
	authRoutes.DELETE("/:username/status/:permalink_id", contentHandler.DeletePostByPermalink)

	// 交互相关新路由 - (X风格URL)
	authRoutes.POST("/:username/status/:permalink_id/like", interactionHandler.LikePostByPermalink)
	authRoutes.DELETE("/:username/status/:permalink_id/like", interactionHandler.UnlikePostByPermalink)
	authRoutes.POST("/:username/status/:permalink_id/comments", interactionHandler.CommentOnPostByPermalink)

	// 帖子保存相关路由 - (X风格URL)
	authRoutes.POST("/:username/status/:permalink_id/save", interactionHandler.BookmarkPostByPermalink)
	authRoutes.DELETE("/:username/status/:permalink_id/save", interactionHandler.UnbookmarkPostByPermalink)

	// (通知相关)
	authRoutes.GET("/notifications", notificationHandler.GetUserNotifications)
	authRoutes.PUT("/notifications/:id/read", notificationHandler.MarkNotificationAsRead)
	authRoutes.PUT("/notifications/read-all", notificationHandler.MarkAllNotificationsAsRead)
	authRoutes.GET("/notifications/settings", notificationHandler.GetNotificationSettings)
	authRoutes.PUT("/notifications/settings", notificationHandler.UpdateNotificationSettings)
	authRoutes.POST("/notifications/devices", notificationHandler.RegisterDevice)
	authRoutes.DELETE("/notifications/devices/:id", notificationHandler.UnregisterDevice)

	// (推荐相关)
	authRoutes.GET("/recommendations/for-you", recommendationHandler.GetRecommendations)
	authRoutes.GET("/recommendations/trending", recommendationHandler.GetTrendingContent)

	// 为admin用户名创建特殊路由，直接注册到authRoutes上，使用显式路径
	// 这样可以避免与管理员路由组冲突
	authRoutes.POST("/admin/status/:permalink_id/like", interactionHandler.LikePostByPermalink)
	authRoutes.DELETE("/admin/status/:permalink_id/like", interactionHandler.UnlikePostByPermalink)
	authRoutes.POST("/admin/status/:permalink_id/save", interactionHandler.BookmarkPostByPermalink)
	authRoutes.DELETE("/admin/status/:permalink_id/save", interactionHandler.UnbookmarkPostByPermalink)
	authRoutes.POST("/admin/status/:permalink_id/comments", interactionHandler.CommentOnPostByPermalink)

	// (管理员路由) - 必须在admin用户交互路由之后注册
	adminRoutes := authRoutes.Group("/admin")
	adminRoutes.Use(middleware.RoleAuth("admin"))

	// (用户管理)
	adminRoutes.GET("/users", userHandler.ListUsers)
	adminRoutes.PUT("/users/by-id/:id/roles", userHandler.UpdateUserRoles)
	adminRoutes.DELETE("/users/by-id/:id", userHandler.DeleteUser)

	// (内容管理)
	adminRoutes.GET("/content/posts/all", contentHandler.GetAllPosts)
	adminRoutes.PUT("/content/posts/:id/status", contentHandler.UpdatePostStatus)

	// (用户资料相关路由 - 这些路由需要身份验证)
	authRoutes.GET("/users/me/follow-stats", userHandler.GetUserFollowStats)

	// (用户头像和封面图片上传路由)
	authRoutes.POST("/users/me/avatar", userHandler.UploadAvatar)
	authRoutes.PUT("/users/me/avatar", userHandler.UploadAvatar)
	authRoutes.POST("/users/me/cover-image", userHandler.UploadCoverImage)
	// authRoutes.PUT("/users/me/cover-image", userHandler.UploadCoverImage) // 添加PUT方法支持

	// 添加一个健康检查路由
	api.GET("/upload-check", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "文件上传服务正常",
		})
	})
}
