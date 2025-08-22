package client

import (
	"context"

	auth_proto "github.com/flick/backend/services/auth/proto"
	content_proto "github.com/flick/backend/services/content/proto"
	interaction_proto "github.com/flick/backend/services/interaction/proto"
	media_proto "github.com/flick/backend/services/media/proto"
	messages_proto "github.com/flick/backend/services/messages/proto"
	notification_proto "github.com/flick/backend/services/notification/proto"
	recommendation_proto "github.com/flick/backend/services/recommendation/proto"
	search_proto "github.com/flick/backend/services/search/proto"
	user_proto "github.com/flick/backend/services/user/proto"
	"google.golang.org/grpc"
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

	// GetFollowers 获取关注者
	GetFollowers(ctx context.Context, in *user_proto.GetFollowersRequest, opts ...interface{}) (*user_proto.GetFollowersResponse, error)

	// GetFollowing 获取正在关注
	GetFollowing(ctx context.Context, in *user_proto.GetFollowingRequest, opts ...interface{}) (*user_proto.GetFollowingResponse, error)

	// UpdateProfile 更新个人资料
	UpdateProfile(ctx context.Context, in *user_proto.UpdateProfileRequest, opts ...interface{}) (*user_proto.UpdateProfileResponse, error)

	// FollowUser 关注用户
	FollowUser(ctx context.Context, in *user_proto.FollowUserRequest, opts ...interface{}) (*user_proto.FollowUserResponse, error)

	// UnfollowUser 取消关注用户
	UnfollowUser(ctx context.Context, in *user_proto.UnfollowUserRequest, opts ...interface{}) (*user_proto.UnfollowUserResponse, error)
}

// ContentServiceClient 定义内容服务客户端接口
type ContentServiceClient interface {
	// CreatePost 创建帖子
	CreatePost(ctx context.Context, in *content_proto.CreatePostRequest, opts ...grpc.CallOption) (*content_proto.CreatePostResponse, error)

	// GetPost 获取帖子
	GetPost(ctx context.Context, in *content_proto.GetPostRequest, opts ...grpc.CallOption) (*content_proto.GetPostResponse, error)

	// GetUserPosts 获取用户帖子列表
	GetUserPosts(ctx context.Context, in *content_proto.GetUserPostsRequest, opts ...grpc.CallOption) (*content_proto.GetUserPostsResponse, error)

	// GetTimeline 获取时间线
	GetTimeline(ctx context.Context, in *content_proto.GetTimelineRequest, opts ...grpc.CallOption) (*content_proto.GetTimelineResponse, error)

	// DeletePost 删除帖子
	DeletePost(ctx context.Context, in *content_proto.DeletePostRequest, opts ...grpc.CallOption) (*content_proto.DeletePostResponse, error)

	// CheckReplyPermission 检查回复权限
	CheckReplyPermission(ctx context.Context, in *content_proto.CheckReplyPermissionRequest, opts ...grpc.CallOption) (*content_proto.CheckReplyPermissionResponse, error)
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

	// CreateComment 创建评论
	CreateComment(ctx context.Context, in *interaction_proto.CreateCommentRequest, opts ...interface{}) (*interaction_proto.CreateCommentResponse, error)

	// DeleteComment 删除评论
	DeleteComment(ctx context.Context, in *interaction_proto.DeleteCommentRequest, opts ...interface{}) (*interaction_proto.DeleteCommentResponse, error)

	// GetComments 获取评论列表
	GetComments(ctx context.Context, in *interaction_proto.GetCommentsRequest, opts ...interface{}) (*interaction_proto.GetCommentsResponse, error)

	// GetPostStats 获取帖子统计
	GetPostStats(ctx context.Context, in *interaction_proto.GetPostStatsRequest, opts ...interface{}) (*interaction_proto.GetPostStatsResponse, error)

	// UpdatePostStats 更新帖子统计
	UpdatePostStats(ctx context.Context, in *interaction_proto.UpdatePostStatsRequest, opts ...interface{}) (*interaction_proto.UpdatePostStatsResponse, error)

	// VotePoll 投票
	VotePoll(ctx context.Context, in *interaction_proto.VotePollRequest, opts ...interface{}) (*interaction_proto.VotePollResponse, error)

	// CreateBookmark 创建收藏
	CreateBookmark(ctx context.Context, in *interaction_proto.CreateBookmarkRequest, opts ...interface{}) (*interaction_proto.CreateBookmarkResponse, error)

	// DeleteBookmark 删除收藏
	DeleteBookmark(ctx context.Context, in *interaction_proto.DeleteBookmarkRequest, opts ...interface{}) (*interaction_proto.DeleteBookmarkResponse, error)

	// IsBookmarked 检查是否收藏
	IsBookmarked(ctx context.Context, in *interaction_proto.IsBookmarkedRequest, opts ...interface{}) (*interaction_proto.IsBookmarkedResponse, error)

	// GetBookmarks 获取收藏列表
	GetBookmarks(ctx context.Context, in *interaction_proto.GetBookmarksRequest, opts ...interface{}) (*interaction_proto.GetBookmarksResponse, error)
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
