package routes

import (
	"backend/services/gateway/config"
	"backend/services/gateway/handlers"
	"backend/services/gateway/middleware"

	"github.com/gin-gonic/gin"
	"net/http"
)

// AddRoute 添加自定义路由并处理可能包含参数的路径
func AddRoute(router *gin.Engine, method string, path string, handler gin.HandlerFunc) {
	// 构建API路径前缀
	apiPath := "/api/v1" + path

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
	api := router.Group("/api/v1")

	// 创建处理程序实例
	userHandler := handlers.NewUserHandler(cfg)
	contentHandler := handlers.NewContentHandler(cfg)
	interactionHandler := handlers.NewInteractionHandler(cfg)
	notificationHandler := handlers.NewNotificationHandler(cfg)
	recommendationHandler := handlers.NewRecommendationHandler(cfg)

	// 无需认证的路由
	// 用户认证相关
	api.POST("/auth/register", userHandler.Register)
	api.POST("/auth/login", userHandler.Login)
	api.POST("/auth/refresh", userHandler.RefreshToken)
	api.POST("/auth/forgot-password", userHandler.ForgotPassword)
	api.POST("/auth/reset-password", userHandler.ResetPassword)
	api.POST("/auth/verify-email", userHandler.VerifyEmail)

	// 内容相关的公开路由
	api.GET("/content/posts", contentHandler.GetPublicPosts)
	api.GET("/content/posts/:id", contentHandler.GetPostByID)

	// 需要认证的路由
	authRoutes := api.Group("")
	authRoutes.Use(middleware.JWTAuth(cfg.JwtSecret))

	// 用户相关
	authRoutes.GET("/users/me", userHandler.GetCurrentUser)
	authRoutes.PUT("/users/me", userHandler.UpdateCurrentUser)
	authRoutes.GET("/users/:id", userHandler.GetUserByID)
	authRoutes.GET("/users/:id/followers", userHandler.GetUserFollowers)
	authRoutes.GET("/users/:id/following", userHandler.GetUserFollowing)
	authRoutes.POST("/users/:id/follow", userHandler.FollowUser)
	authRoutes.DELETE("/users/:id/follow", userHandler.UnfollowUser)
	authRoutes.POST("/users/:id/block", userHandler.BlockUser)
	authRoutes.DELETE("/users/:id/block", userHandler.UnblockUser)
	authRoutes.GET("/users/search", userHandler.SearchUsers)

	// 内容相关
	authRoutes.POST("/content/posts", contentHandler.CreatePost)
	authRoutes.PUT("/content/posts/:id", contentHandler.UpdatePost)
	authRoutes.DELETE("/content/posts/:id", contentHandler.DeletePost)
	authRoutes.GET("/content/users/:id/posts", contentHandler.GetUserPosts)
	authRoutes.GET("/content/feed", contentHandler.GetUserFeed)

	// 帖子投票相关路由
	authRoutes.POST("/content/posts/:id/vote", contentHandler.VotePoll)
	authRoutes.GET("/content/posts/:id/poll", contentHandler.GetPollResults)

	// 帖子保存相关路由
	authRoutes.POST("/content/posts/:id/save", contentHandler.SavePost)
	authRoutes.DELETE("/content/posts/:id/save", contentHandler.UnsavePost)
	authRoutes.GET("/content/saved", contentHandler.GetSavedPosts)

	// 热门帖子路由
	authRoutes.GET("/content/posts/top", contentHandler.GetTopPosts)

	// 帖子统计数据路由
	authRoutes.GET("/content/posts/:id/stats", contentHandler.GetPostStats)

	// 举报帖子路由
	authRoutes.POST("/content/posts/:id/report", contentHandler.ReportPost)

	// 交互相关
	authRoutes.POST("/interactions/posts/:id/like", interactionHandler.LikePost)
	authRoutes.DELETE("/interactions/posts/:id/like", interactionHandler.UnlikePost)
	authRoutes.POST("/interactions/posts/:id/comments", interactionHandler.CommentOnPost)
	authRoutes.GET("/interactions/posts/:id/comments", interactionHandler.GetPostComments)
	authRoutes.PUT("/interactions/comments/:id", interactionHandler.UpdateComment)
	authRoutes.DELETE("/interactions/comments/:id", interactionHandler.DeleteComment)
	authRoutes.POST("/interactions/comments/:id/like", interactionHandler.LikeComment)
	authRoutes.DELETE("/interactions/comments/:id/like", interactionHandler.UnlikeComment)

	// 通知相关
	authRoutes.GET("/notifications", notificationHandler.GetUserNotifications)
	authRoutes.PUT("/notifications/:id/read", notificationHandler.MarkNotificationAsRead)
	authRoutes.PUT("/notifications/read-all", notificationHandler.MarkAllNotificationsAsRead)
	authRoutes.GET("/notifications/settings", notificationHandler.GetNotificationSettings)
	authRoutes.PUT("/notifications/settings", notificationHandler.UpdateNotificationSettings)
	authRoutes.POST("/notifications/devices", notificationHandler.RegisterDevice)
	authRoutes.DELETE("/notifications/devices/:id", notificationHandler.UnregisterDevice)

	// 推荐相关
	authRoutes.GET("/recommendations/for-you", recommendationHandler.GetRecommendations)
	authRoutes.GET("/recommendations/trending", recommendationHandler.GetTrendingContent)
	authRoutes.GET("/recommendations/similar/:type/:id", recommendationHandler.GetSimilarContent)
	authRoutes.POST("/recommendations/:id/feedback", recommendationHandler.RecordFeedback)
	authRoutes.PUT("/recommendations/:id/view", recommendationHandler.MarkAsViewed)
	authRoutes.PUT("/recommendations/:id/click", recommendationHandler.MarkAsClicked)

	// 管理员路由
	adminRoutes := authRoutes.Group("/admin")
	adminRoutes.Use(middleware.RoleAuth("admin"))

	// 用户管理
	adminRoutes.GET("/users", userHandler.ListUsers)
	adminRoutes.PUT("/users/:id/roles", userHandler.UpdateUserRoles)
	adminRoutes.DELETE("/users/:id", userHandler.DeleteUser)

	// 内容管理
	adminRoutes.GET("/content/posts/all", contentHandler.GetAllPosts)
	adminRoutes.PUT("/content/posts/:id/status", contentHandler.UpdatePostStatus)

	// 推荐模型管理
	adminRoutes.GET("/recommendations/models", recommendationHandler.ListModels)
	adminRoutes.POST("/recommendations/models", recommendationHandler.CreateModel)
	adminRoutes.PUT("/recommendations/models/:id", recommendationHandler.UpdateModel)
	adminRoutes.DELETE("/recommendations/models/:id", recommendationHandler.DeleteModel)

	// A/B测试管理
	adminRoutes.GET("/recommendations/ab-tests", recommendationHandler.ListABTests)
	adminRoutes.POST("/recommendations/ab-tests", recommendationHandler.CreateABTest)
	adminRoutes.GET("/recommendations/ab-tests/:id/metrics", recommendationHandler.GetABTestMetrics)
	authRoutes.PUT("/recommendations/ab-tests/:id", recommendationHandler.UpdateABTest)
	adminRoutes.DELETE("/recommendations/ab-tests/:id", recommendationHandler.DeleteABTest)

	// 用户资料相关路由 - 这些路由需要身份验证
	authRoutes.GET("/users/me/follow-stats", userHandler.GetUserFollowStats)

	// 用户头像和封面图片上传路由 - 支持POST和PUT两种方法
	authRoutes.POST("/users/me/avatar", userHandler.UploadAvatar)
	authRoutes.PUT("/users/me/avatar", userHandler.UploadAvatar) // 添加PUT方法支持
	authRoutes.POST("/users/me/cover-image", userHandler.UploadCoverImage)
	authRoutes.PUT("/users/me/cover-image", userHandler.UploadCoverImage) // 添加PUT方法支持

	// 添加一个健康检查路由
	api.GET("/upload-check", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "文件上传服务正常",
		})
	})
}
