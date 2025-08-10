package client

import (
	"context"

	auth_proto "backend/services/auth/proto"
	bookmark_proto "backend/services/bookmark/proto"
	content_proto "backend/services/content/proto"
	interaction_proto "backend/services/interaction/proto"
	media_proto "backend/services/media/proto"
	messages_proto "backend/services/messages/proto"
	notification_proto "backend/services/notification/proto"
	recommendation_proto "backend/services/recommendation/proto"
	search_proto "backend/services/search/proto"
	user_proto "backend/services/user/proto"
)

// UserServiceClient 定义用户服务客户端接口
type UserServiceClient interface {
	// Register 用户注册
	Register(ctx context.Context, in *user_proto.RegisterRequest, opts ...interface{}) (*user_proto.RegisterResponse, error)

	// Login 用户登录
	Login(ctx context.Context, in *user_proto.LoginRequest, opts ...interface{}) (*user_proto.LoginResponse, error)

	// GetUser 获取用户
	GetUser(ctx context.Context, in *user_proto.GetUserRequest, opts ...interface{}) (*user_proto.GetUserResponse, error)

	// GetUserByUsername 获取用户
	GetUserByUsername(ctx context.Context, in *user_proto.GetUserByUsernameRequest, opts ...interface{}) (*user_proto.GetUserResponse, error)

	// UpdateUser 更新用户
	UpdateUser(ctx context.Context, in *user_proto.UpdateUserRequest, opts ...interface{}) (*user_proto.UpdateUserResponse, error)

	// DeleteUser 删除用户
	DeleteUser(ctx context.Context, in *user_proto.DeleteUserRequest, opts ...interface{}) (*user_proto.DeleteUserResponse, error)
}

// ContentServiceClient 定义内容服务客户端接口
type ContentServiceClient interface {
	// GetContent 获取内容
	GetContent(ctx context.Context, in *content_proto.GetContentRequest, opts ...interface{}) (*content_proto.GetContentResponse, error)

	// CreateContent 创建内容
	CreateContent(ctx context.Context, in *content_proto.CreateContentRequest, opts ...interface{}) (*content_proto.CreateContentResponse, error)

	// UpdateContent 更新内容
	UpdateContent(ctx context.Context, in *content_proto.UpdateContentRequest, opts ...interface{}) (*content_proto.UpdateContentResponse, error)

	// DeleteContent 删除内容
	DeleteContent(ctx context.Context, in *content_proto.DeleteContentRequest, opts ...interface{}) (*content_proto.DeleteContentResponse, error)

	// ListContent 列出内容
	ListContent(ctx context.Context, in *content_proto.ListContentRequest, opts ...interface{}) (*content_proto.ListContentResponse, error)
}

// AuthServiceClient 定义认证服务客户端接口
type AuthServiceClient interface {
	// Login 用户登录
	Login(ctx context.Context, in *auth_proto.LoginRequest, opts ...interface{}) (*auth_proto.LoginResponse, error)

	// Register 用户注册
	Register(ctx context.Context, in *auth_proto.RegisterRequest, opts ...interface{}) (*auth_proto.RegisterResponse, error)

	// ValidateToken 验证令牌
	ValidateToken(ctx context.Context, in *auth_proto.ValidateTokenRequest, opts ...interface{}) (*auth_proto.ValidateTokenResponse, error)

	// RefreshToken 刷新令牌
	RefreshToken(ctx context.Context, in *auth_proto.RefreshTokenRequest, opts ...interface{}) (*auth_proto.RefreshTokenResponse, error)

	// Logout 登出
	Logout(ctx context.Context, in *auth_proto.LogoutRequest, opts ...interface{}) (*auth_proto.LogoutResponse, error)

	// GithubLogin Github登录
	GithubLogin(ctx context.Context, in *auth_proto.GithubLoginRequest, opts ...interface{}) (*auth_proto.GithubLoginResponse, error)

	// GithubCallback Github回调
	GithubCallback(ctx context.Context, in *auth_proto.GithubCallbackRequest, opts ...interface{}) (*auth_proto.GithubCallbackResponse, error)

	// GoogleLogin Google登录
	GoogleLogin(ctx context.Context, in *auth_proto.GoogleLoginRequest, opts ...interface{}) (*auth_proto.GoogleLoginResponse, error)

	// GoogleCallback Google回调
	GoogleCallback(ctx context.Context, in *auth_proto.GoogleCallbackRequest, opts ...interface{}) (*auth_proto.GoogleCallbackResponse, error)
}

// MediaServiceClient 定义媒体服务客户端接口
type MediaServiceClient interface {
	// UploadFile 上传文件
	UploadFile(ctx context.Context, in *media_proto.UploadFileRequest, opts ...interface{}) (*media_proto.UploadFileResponse, error)

	// GetFile 获取文件信息
	GetFile(ctx context.Context, in *media_proto.GetFileRequest, opts ...interface{}) (*media_proto.GetFileResponse, error)

	// DeleteFile 删除文件
	DeleteFile(ctx context.Context, in *media_proto.DeleteFileRequest, opts ...interface{}) (*media_proto.DeleteFileResponse, error)

	// ListFiles 获取文件列表
	ListFiles(ctx context.Context, in *media_proto.ListFilesRequest, opts ...interface{}) (*media_proto.ListFilesResponse, error)
}

// MessageServiceClient 定义消息服务客户端接口
type MessageServiceClient interface {
	// CreateConversation 创建会话
	CreateConversation(ctx context.Context, in *messages_proto.CreateConversationRequest, opts ...interface{}) (*messages_proto.CreateConversationResponse, error)

	// ListConversations 获取会话列表
	ListConversations(ctx context.Context, in *messages_proto.ListConversationsRequest, opts ...interface{}) (*messages_proto.ListConversationsResponse, error)

	// GetConversation 获取会话详情
	GetConversation(ctx context.Context, in *messages_proto.GetConversationRequest, opts ...interface{}) (*messages_proto.GetConversationResponse, error)

	// SendMessage 发送消息
	SendMessage(ctx context.Context, in *messages_proto.SendMessageRequest, opts ...interface{}) (*messages_proto.SendMessageResponse, error)

	// ListMessages 获取消息列表
	ListMessages(ctx context.Context, in *messages_proto.ListMessagesRequest, opts ...interface{}) (*messages_proto.ListMessagesResponse, error)

	// MarkAsRead 标记消息为已读
	MarkAsRead(ctx context.Context, in *messages_proto.MarkAsReadRequest, opts ...interface{}) (*messages_proto.MarkAsReadResponse, error)

	// DeleteConversation 删除会话
	DeleteConversation(ctx context.Context, in *messages_proto.DeleteConversationRequest, opts ...interface{}) (*messages_proto.DeleteConversationResponse, error)
}

// NotificationServiceClient 定义通知服务客户端接口
type NotificationServiceClient interface {
	// CreateNotification 创建通知
	CreateNotification(ctx context.Context, in *notification_proto.CreateNotificationRequest, opts ...interface{}) (*notification_proto.CreateNotificationResponse, error)

	// ListNotifications 获取用户通知列表
	ListNotifications(ctx context.Context, in *notification_proto.ListNotificationsRequest, opts ...interface{}) (*notification_proto.ListNotificationsResponse, error)

	// MarkAsRead 标记通知为已读
	MarkAsRead(ctx context.Context, in *notification_proto.MarkAsReadRequest, opts ...interface{}) (*notification_proto.MarkAsReadResponse, error)

	// MarkAllAsRead 标记所有通知为已读
	MarkAllAsRead(ctx context.Context, in *notification_proto.MarkAllAsReadRequest, opts ...interface{}) (*notification_proto.MarkAllAsReadResponse, error)

	// DeleteNotification 删除通知
	DeleteNotification(ctx context.Context, in *notification_proto.DeleteNotificationRequest, opts ...interface{}) (*notification_proto.DeleteNotificationResponse, error)

	// GetUnreadCount 获取未读通知数
	GetUnreadCount(ctx context.Context, in *notification_proto.GetUnreadCountRequest, opts ...interface{}) (*notification_proto.GetUnreadCountResponse, error)
}

// InteractionServiceClient 定义互动服务客户端接口
type InteractionServiceClient interface {
	// CreateFollow 创建关注
	CreateFollow(ctx context.Context, in *interaction_proto.CreateFollowRequest, opts ...interface{}) (*interaction_proto.CreateFollowResponse, error)

	// DeleteFollow 删除关注
	DeleteFollow(ctx context.Context, in *interaction_proto.DeleteFollowRequest, opts ...interface{}) (*interaction_proto.DeleteFollowResponse, error)

	// IsFollowing 检查是否关注
	IsFollowing(ctx context.Context, in *interaction_proto.IsFollowingRequest, opts ...interface{}) (*interaction_proto.IsFollowingResponse, error)

	// GetFollowers 获取粉丝列表
	GetFollowers(ctx context.Context, in *interaction_proto.GetFollowersRequest, opts ...interface{}) (*interaction_proto.GetFollowersResponse, error)

	// GetFollowing 获取关注列表
	GetFollowing(ctx context.Context, in *interaction_proto.GetFollowingRequest, opts ...interface{}) (*interaction_proto.GetFollowingResponse, error)

	// CreateLike 创建点赞
	CreateLike(ctx context.Context, in *interaction_proto.CreateLikeRequest, opts ...interface{}) (*interaction_proto.CreateLikeResponse, error)

	// DeleteLike 删除点赞
	DeleteLike(ctx context.Context, in *interaction_proto.DeleteLikeRequest, opts ...interface{}) (*interaction_proto.DeleteLikeResponse, error)

	// IsLiked 检查是否点赞
	IsLiked(ctx context.Context, in *interaction_proto.IsLikedRequest, opts ...interface{}) (*interaction_proto.IsLikedResponse, error)

	// GetLikes 获取点赞列表
	GetLikes(ctx context.Context, in *interaction_proto.GetLikesRequest, opts ...interface{}) (*interaction_proto.GetLikesResponse, error)

	// CreateRepost 创建转发
	CreateRepost(ctx context.Context, in *interaction_proto.CreateRepostRequest, opts ...interface{}) (*interaction_proto.CreateRepostResponse, error)

	// DeleteRepost 删除转发
	DeleteRepost(ctx context.Context, in *interaction_proto.DeleteRepostRequest, opts ...interface{}) (*interaction_proto.DeleteRepostResponse, error)

	// CreateReport 创建举报
	CreateReport(ctx context.Context, in *interaction_proto.CreateReportRequest, opts ...interface{}) (*interaction_proto.CreateReportResponse, error)
}

// RecommendationServiceClient 定义推荐服务客户端接口
type RecommendationServiceClient interface {
	// GetRecommendations 获取推荐内容
	GetRecommendations(ctx context.Context, in *recommendation_proto.GetRecommendationsRequest, opts ...interface{}) (*recommendation_proto.GetRecommendationsResponse, error)

	// RecordUserAction 记录用户行为
	RecordUserAction(ctx context.Context, in *recommendation_proto.RecordUserActionRequest, opts ...interface{}) (*recommendation_proto.RecordUserActionResponse, error)

	// UpdateUserInterest 更新用户兴趣
	UpdateUserInterest(ctx context.Context, in *recommendation_proto.UpdateUserInterestRequest, opts ...interface{}) (*recommendation_proto.UpdateUserInterestResponse, error)
}

// SearchServiceClient 定义搜索服务客户端接口
type SearchServiceClient interface {
	// SearchContent 搜索内容
	SearchContent(ctx context.Context, in *search_proto.SearchContentRequest, opts ...interface{}) (*search_proto.SearchContentResponse, error)

	// SearchUsers 搜索用户
	SearchUsers(ctx context.Context, in *search_proto.SearchUsersRequest, opts ...interface{}) (*search_proto.SearchUsersResponse, error)

	// SearchHashtags 搜索标签
	SearchHashtags(ctx context.Context, in *search_proto.SearchHashtagsRequest, opts ...interface{}) (*search_proto.SearchHashtagsResponse, error)
}

// BookmarkServiceClient 定义书签服务客户端接口
type BookmarkServiceClient interface {
	// CreateBookmark 创建书签
	CreateBookmark(ctx context.Context, in *bookmark_proto.CreateBookmarkRequest, opts ...interface{}) (*bookmark_proto.CreateBookmarkResponse, error)

	// DeleteBookmark 删除书签
	DeleteBookmark(ctx context.Context, in *bookmark_proto.DeleteBookmarkRequest, opts ...interface{}) (*bookmark_proto.DeleteBookmarkResponse, error)

	// IsBookmarked 检查是否已收藏
	IsBookmarked(ctx context.Context, in *bookmark_proto.IsBookmarkedRequest, opts ...interface{}) (*bookmark_proto.IsBookmarkedResponse, error)

	// ListBookmarks 获取用户书签列表
	ListBookmarks(ctx context.Context, in *bookmark_proto.ListBookmarksRequest, opts ...interface{}) (*bookmark_proto.ListBookmarksResponse, error)
}
