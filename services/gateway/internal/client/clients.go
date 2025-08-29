package client

import (
	"context"

	"google.golang.org/grpc"

	auth_proto "github.com/flick/backend/services/auth/proto"
	content_proto "github.com/flick/backend/services/content/proto"
	interaction_proto "github.com/flick/backend/services/interaction/proto"
	media_proto "github.com/flick/backend/services/media/proto"
	messages_proto "github.com/flick/backend/services/messages/proto"
	notification_proto "github.com/flick/backend/services/notification/proto"
	recommendation_proto "github.com/flick/backend/services/recommendation/proto"
	search_proto "github.com/flick/backend/services/search/proto"
	user_proto "github.com/flick/backend/services/user/proto"
)

// userServiceClient 用户服务gRPC客户端实现
type userServiceClient struct {
	client user_proto.UserServiceClient
}

// NewUserServiceClient 创建用户服务客户端
func NewUserServiceClient(conn *grpc.ClientConn) UserServiceClient {
	return &userServiceClient{
		client: user_proto.NewUserServiceClient(conn),
	}
}

// GetUser 获取用户
func (c *userServiceClient) GetUser(ctx context.Context, in *user_proto.GetUserRequest, opts ...interface{}) (*user_proto.GetUserResponse, error) {
	return c.client.GetUser(ctx, in)
}

// GetUserByUsername 获取用户
func (c *userServiceClient) GetUserByUsername(ctx context.Context, in *user_proto.GetUserByUsernameRequest, opts ...interface{}) (*user_proto.GetUserResponse, error) {
	return c.client.GetUserByUsername(ctx, in)
}

// UpdateUser 更新用户
func (c *userServiceClient) UpdateUser(ctx context.Context, in *user_proto.UpdateUserRequest, opts ...interface{}) (*user_proto.UpdateUserResponse, error) {
	return c.client.UpdateUser(ctx, in)
}

// DeleteUser 删除用户
func (c *userServiceClient) DeleteUser(ctx context.Context, in *user_proto.DeleteUserRequest, opts ...interface{}) (*user_proto.DeleteUserResponse, error) {
	return c.client.DeleteUser(ctx, in)
}

func (c *userServiceClient) GetFollowers(ctx context.Context, in *user_proto.GetFollowersRequest, opts ...interface{}) (*user_proto.GetFollowersResponse, error) {
	return c.client.GetFollowers(ctx, in)
}

func (c *userServiceClient) GetFollowing(ctx context.Context, in *user_proto.GetFollowingRequest, opts ...interface{}) (*user_proto.GetFollowingResponse, error) {
	return c.client.GetFollowing(ctx, in)
}

// Register 用户注册
func (c *userServiceClient) Register(ctx context.Context, in *user_proto.RegisterRequest, opts ...interface{}) (*user_proto.RegisterResponse, error) {
	return c.client.Register(ctx, in)
}

// Login 用户登录
func (c *userServiceClient) Login(ctx context.Context, in *user_proto.LoginRequest, opts ...interface{}) (*user_proto.LoginResponse, error) {
	return c.client.Login(ctx, in)
}

// UpdateProfile 更新个人资料
func (c *userServiceClient) UpdateProfile(ctx context.Context, in *user_proto.UpdateProfileRequest, opts ...interface{}) (*user_proto.UpdateProfileResponse, error) {
	return c.client.UpdateProfile(ctx, in)
}

// FollowUser 关注用户
func (c *userServiceClient) FollowUser(ctx context.Context, in *user_proto.FollowUserRequest, opts ...interface{}) (*user_proto.FollowUserResponse, error) {
	return c.client.FollowUser(ctx, in)
}

// UnfollowUser 取消关注用户
func (c *userServiceClient) UnfollowUser(ctx context.Context, in *user_proto.UnfollowUserRequest, opts ...interface{}) (*user_proto.UnfollowUserResponse, error) {
	return c.client.UnfollowUser(ctx, in)
}

// contentServiceClient 内容服务gRPC客户端实现
type contentServiceClient struct {
	client content_proto.ContentServiceClient
}

// NewContentServiceClient 创建内容服务客户端
func NewContentServiceClient(conn *grpc.ClientConn) ContentServiceClient {
	return &contentServiceClient{
		client: content_proto.NewContentServiceClient(conn),
	}
}

// CreatePost 创建帖子
func (c *contentServiceClient) CreatePost(ctx context.Context, in *content_proto.CreatePostRequest, opts ...grpc.CallOption) (*content_proto.CreatePostResponse, error) {
	return c.client.CreatePost(ctx, in, opts...)
}

// GetPost 获取帖子
func (c *contentServiceClient) GetPost(ctx context.Context, in *content_proto.GetPostRequest, opts ...grpc.CallOption) (*content_proto.GetPostResponse, error) {
	return c.client.GetPost(ctx, in, opts...)
}

// GetUserPosts 获取用户帖子列表
func (c *contentServiceClient) GetUserPosts(ctx context.Context, in *content_proto.GetUserPostsRequest, opts ...grpc.CallOption) (*content_proto.GetUserPostsResponse, error) {
	return c.client.GetUserPosts(ctx, in, opts...)
}

// GetTimeline 获取时间线
func (c *contentServiceClient) GetTimeline(ctx context.Context, in *content_proto.GetTimelineRequest, opts ...grpc.CallOption) (*content_proto.GetTimelineResponse, error) {
	return c.client.GetTimeline(ctx, in, opts...)
}

// GetFollowingTimeline 获取关注用户时间线
func (c *contentServiceClient) GetFollowingTimeline(ctx context.Context, in *content_proto.GetTimelineRequest, opts ...grpc.CallOption) (*content_proto.GetTimelineResponse, error) {
	return c.client.GetFollowingTimeline(ctx, in, opts...)
}

// DeletePost 删除帖子
func (c *contentServiceClient) DeletePost(ctx context.Context, in *content_proto.DeletePostRequest, opts ...grpc.CallOption) (*content_proto.DeletePostResponse, error) {
	return c.client.DeletePost(ctx, in, opts...)
}

// CheckReplyPermission 检查回复权限
func (c *contentServiceClient) CheckReplyPermission(ctx context.Context, in *content_proto.CheckReplyPermissionRequest, opts ...grpc.CallOption) (*content_proto.CheckReplyPermissionResponse, error) {
	return c.client.CheckReplyPermission(ctx, in, opts...)
}

// GetPostReplies 获取帖子回复
func (c *contentServiceClient) GetPostReplies(ctx context.Context, in *content_proto.GetPostRepliesRequest, opts ...grpc.CallOption) (*content_proto.GetPostRepliesResponse, error) {
	return c.client.GetPostReplies(ctx, in, opts...)
}

// GetConversationThread 获取对话线程
func (c *contentServiceClient) GetConversationThread(ctx context.Context, in *content_proto.GetConversationThreadRequest, opts ...grpc.CallOption) (*content_proto.GetConversationThreadResponse, error) {
	return c.client.GetConversationThread(ctx, in, opts...)
}

// DeleteReply 删除回复
func (c *contentServiceClient) DeleteReply(ctx context.Context, in *content_proto.DeleteReplyRequest, opts ...grpc.CallOption) (*content_proto.DeleteReplyResponse, error) {
	return c.client.DeleteReply(ctx, in, opts...)
}

// GetReplyMention 获取回复提及信息
func (c *contentServiceClient) GetReplyMention(ctx context.Context, in *content_proto.GetReplyMentionRequest, opts ...grpc.CallOption) (*content_proto.GetReplyMentionResponse, error) {
	return c.client.GetReplyMention(ctx, in, opts...)
}

// authServiceClient 认证服务gRPC客户端实现
type authServiceClient struct {
	client auth_proto.AuthServiceClient
}

// NewAuthServiceClient 创建认证服务客户端
func NewAuthServiceClient(conn *grpc.ClientConn) AuthServiceClient {
	return &authServiceClient{
		client: auth_proto.NewAuthServiceClient(conn),
	}
}

// Login 用户登录
func (c *authServiceClient) Login(ctx context.Context, in *auth_proto.LoginRequest, opts ...interface{}) (*auth_proto.LoginResponse, error) {
	return c.client.Login(ctx, in)
}

// Register 用户注册
func (c *authServiceClient) Register(ctx context.Context, in *auth_proto.RegisterRequest, opts ...interface{}) (*auth_proto.RegisterResponse, error) {
	return c.client.Register(ctx, in)
}

// ValidateToken 验证令牌
func (c *authServiceClient) ValidateToken(ctx context.Context, in *auth_proto.ValidateTokenRequest, opts ...interface{}) (*auth_proto.ValidateTokenResponse, error) {
	return c.client.ValidateToken(ctx, in)
}

// RefreshToken 刷新令牌
func (c *authServiceClient) RefreshToken(ctx context.Context, in *auth_proto.RefreshTokenRequest, opts ...interface{}) (*auth_proto.RefreshTokenResponse, error) {
	return c.client.RefreshToken(ctx, in)
}

// Logout 登出
func (c *authServiceClient) Logout(ctx context.Context, in *auth_proto.LogoutRequest, opts ...interface{}) (*auth_proto.LogoutResponse, error) {
	return c.client.Logout(ctx, in)
}

// GithubLogin Github登录
func (c *authServiceClient) GithubLogin(ctx context.Context, in *auth_proto.GithubLoginRequest, opts ...interface{}) (*auth_proto.GithubLoginResponse, error) {
	return c.client.GithubLogin(ctx, in)
}

// GithubCallback Github回调
func (c *authServiceClient) GithubCallback(ctx context.Context, in *auth_proto.GithubCallbackRequest, opts ...interface{}) (*auth_proto.GithubCallbackResponse, error) {
	return c.client.GithubCallback(ctx, in)
}

// GoogleLogin Google登录
func (c *authServiceClient) GoogleLogin(ctx context.Context, in *auth_proto.GoogleLoginRequest, opts ...interface{}) (*auth_proto.GoogleLoginResponse, error) {
	return c.client.GoogleLogin(ctx, in)
}

// GoogleCallback Google回调
func (c *authServiceClient) GoogleCallback(ctx context.Context, in *auth_proto.GoogleCallbackRequest, opts ...interface{}) (*auth_proto.GoogleCallbackResponse, error) {
	return c.client.GoogleCallback(ctx, in)
}

// mediaServiceClient 媒体服务gRPC客户端实现
type mediaServiceClient struct {
	client media_proto.MediaServiceClient
}

// NewMediaServiceClient 创建媒体服务客户端
func NewMediaServiceClient(conn *grpc.ClientConn) MediaServiceClient {
	return &mediaServiceClient{
		client: media_proto.NewMediaServiceClient(conn),
	}
}

// UploadFile 上传文件
func (c *mediaServiceClient) UploadFile(ctx context.Context, in *media_proto.UploadFileRequest, opts ...interface{}) (*media_proto.UploadFileResponse, error) {
	return c.client.UploadFile(ctx, in)
}

// GetFile 获取文件信息
func (c *mediaServiceClient) GetFile(ctx context.Context, in *media_proto.GetFileRequest, opts ...interface{}) (*media_proto.GetFileResponse, error) {
	return c.client.GetFile(ctx, in)
}

// DeleteFile 删除文件
func (c *mediaServiceClient) DeleteFile(ctx context.Context, in *media_proto.DeleteFileRequest, opts ...interface{}) (*media_proto.DeleteFileResponse, error) {
	return c.client.DeleteFile(ctx, in)
}

// ListFiles 获取文件列表
func (c *mediaServiceClient) ListFiles(ctx context.Context, in *media_proto.ListFilesRequest, opts ...interface{}) (*media_proto.ListFilesResponse, error) {
	return c.client.ListFiles(ctx, in)
}

// messageServiceClient 消息服务gRPC客户端实现
type messageServiceClient struct {
	client messages_proto.MessageServiceClient
}

// NewMessageServiceClient 创建消息服务客户端
func NewMessageServiceClient(conn *grpc.ClientConn) MessageServiceClient {
	return &messageServiceClient{
		client: messages_proto.NewMessageServiceClient(conn),
	}
}

// CreateConversation 创建会话
func (c *messageServiceClient) CreateConversation(ctx context.Context, in *messages_proto.CreateConversationRequest, opts ...interface{}) (*messages_proto.CreateConversationResponse, error) {
	return c.client.CreateConversation(ctx, in)
}

// ListConversations 获取会话列表
func (c *messageServiceClient) ListConversations(ctx context.Context, in *messages_proto.ListConversationsRequest, opts ...interface{}) (*messages_proto.ListConversationsResponse, error) {
	return c.client.ListConversations(ctx, in)
}

// GetConversation 获取会话详情
func (c *messageServiceClient) GetConversation(ctx context.Context, in *messages_proto.GetConversationRequest, opts ...interface{}) (*messages_proto.GetConversationResponse, error) {
	return c.client.GetConversation(ctx, in)
}

// SendMessage 发送消息
func (c *messageServiceClient) SendMessage(ctx context.Context, in *messages_proto.SendMessageRequest, opts ...interface{}) (*messages_proto.SendMessageResponse, error) {
	return c.client.SendMessage(ctx, in)
}

// ListMessages 获取消息列表
func (c *messageServiceClient) ListMessages(ctx context.Context, in *messages_proto.ListMessagesRequest, opts ...interface{}) (*messages_proto.ListMessagesResponse, error) {
	return c.client.ListMessages(ctx, in)
}

// MarkAsRead 标记消息为已读
func (c *messageServiceClient) MarkAsRead(ctx context.Context, in *messages_proto.MarkAsReadRequest, opts ...interface{}) (*messages_proto.MarkAsReadResponse, error) {
	return c.client.MarkAsRead(ctx, in)
}

// DeleteConversation 删除会话
func (c *messageServiceClient) DeleteConversation(ctx context.Context, in *messages_proto.DeleteConversationRequest, opts ...interface{}) (*messages_proto.DeleteConversationResponse, error) {
	return c.client.DeleteConversation(ctx, in)
}

// notificationServiceClient 通知服务gRPC客户端实现
type notificationServiceClient struct {
	client notification_proto.NotificationServiceClient
}

// NewNotificationServiceClient 创建通知服务客户端
func NewNotificationServiceClient(conn *grpc.ClientConn) NotificationServiceClient {
	return &notificationServiceClient{
		client: notification_proto.NewNotificationServiceClient(conn),
	}
}

// CreateNotification 创建通知
func (c *notificationServiceClient) CreateNotification(ctx context.Context, in *notification_proto.CreateNotificationRequest, opts ...interface{}) (*notification_proto.CreateNotificationResponse, error) {
	return c.client.CreateNotification(ctx, in)
}

// ListNotifications 获取用户通知列表
func (c *notificationServiceClient) ListNotifications(ctx context.Context, in *notification_proto.ListNotificationsRequest, opts ...interface{}) (*notification_proto.ListNotificationsResponse, error) {
	return c.client.ListNotifications(ctx, in)
}

// MarkAsRead 标记通知为已读
func (c *notificationServiceClient) MarkAsRead(ctx context.Context, in *notification_proto.MarkAsReadRequest, opts ...interface{}) (*notification_proto.MarkAsReadResponse, error) {
	return c.client.MarkAsRead(ctx, in)
}

// MarkAllAsRead 标记所有通知为已读
func (c *notificationServiceClient) MarkAllAsRead(ctx context.Context, in *notification_proto.MarkAllAsReadRequest, opts ...interface{}) (*notification_proto.MarkAllAsReadResponse, error) {
	return c.client.MarkAllAsRead(ctx, in)
}

// DeleteNotification 删除通知
func (c *notificationServiceClient) DeleteNotification(ctx context.Context, in *notification_proto.DeleteNotificationRequest, opts ...interface{}) (*notification_proto.DeleteNotificationResponse, error) {
	return c.client.DeleteNotification(ctx, in)
}

// GetUnreadCount 获取未读通知数
func (c *notificationServiceClient) GetUnreadCount(ctx context.Context, in *notification_proto.GetUnreadCountRequest, opts ...interface{}) (*notification_proto.GetUnreadCountResponse, error) {
	return c.client.GetUnreadCount(ctx, in)
}

// interactionServiceClient 互动服务gRPC客户端实现
type interactionServiceClient struct {
	client interaction_proto.InteractionServiceClient
}

// NewInteractionServiceClient 创建互动服务客户端
func NewInteractionServiceClient(conn *grpc.ClientConn) InteractionServiceClient {
	return &interactionServiceClient{
		client: interaction_proto.NewInteractionServiceClient(conn),
	}
}

// CreateFollow 创建关注
func (c *interactionServiceClient) CreateFollow(ctx context.Context, in *interaction_proto.CreateFollowRequest, opts ...interface{}) (*interaction_proto.CreateFollowResponse, error) {
	return c.client.CreateFollow(ctx, in)
}

// DeleteFollow 删除关注
func (c *interactionServiceClient) DeleteFollow(ctx context.Context, in *interaction_proto.DeleteFollowRequest, opts ...interface{}) (*interaction_proto.DeleteFollowResponse, error) {
	return c.client.DeleteFollow(ctx, in)
}

// IsFollowing 检查是否关注
func (c *interactionServiceClient) IsFollowing(ctx context.Context, in *interaction_proto.IsFollowingRequest, opts ...interface{}) (*interaction_proto.IsFollowingResponse, error) {
	return c.client.IsFollowing(ctx, in)
}

// GetFollowers 获取粉丝列表
func (c *interactionServiceClient) GetFollowers(ctx context.Context, in *interaction_proto.GetFollowersRequest, opts ...interface{}) (*interaction_proto.GetFollowersResponse, error) {
	return c.client.GetFollowers(ctx, in)
}

// GetFollowing 获取关注列表
func (c *interactionServiceClient) GetFollowing(ctx context.Context, in *interaction_proto.GetFollowingRequest, opts ...interface{}) (*interaction_proto.GetFollowingResponse, error) {
	return c.client.GetFollowing(ctx, in)
}

// CreateLike 创建点赞
func (c *interactionServiceClient) CreateLike(ctx context.Context, in *interaction_proto.CreateLikeRequest, opts ...interface{}) (*interaction_proto.CreateLikeResponse, error) {
	return c.client.CreateLike(ctx, in)
}

// DeleteLike 删除点赞
func (c *interactionServiceClient) DeleteLike(ctx context.Context, in *interaction_proto.DeleteLikeRequest, opts ...interface{}) (*interaction_proto.DeleteLikeResponse, error) {
	return c.client.DeleteLike(ctx, in)
}

// IsLiked 检查是否点赞
func (c *interactionServiceClient) IsLiked(ctx context.Context, in *interaction_proto.IsLikedRequest, opts ...interface{}) (*interaction_proto.IsLikedResponse, error) {
	return c.client.IsLiked(ctx, in)
}

// GetLikes 获取点赞列表
func (c *interactionServiceClient) GetLikes(ctx context.Context, in *interaction_proto.GetLikesRequest, opts ...interface{}) (*interaction_proto.GetLikesResponse, error) {
	return c.client.GetLikes(ctx, in)
}

// CreateRepost 创建转发
func (c *interactionServiceClient) CreateRepost(ctx context.Context, in *interaction_proto.CreateRepostRequest, opts ...interface{}) (*interaction_proto.CreateRepostResponse, error) {
	return c.client.CreateRepost(ctx, in)
}

// DeleteRepost 删除转发
func (c *interactionServiceClient) DeleteRepost(ctx context.Context, in *interaction_proto.DeleteRepostRequest, opts ...interface{}) (*interaction_proto.DeleteRepostResponse, error) {
	return c.client.DeleteRepost(ctx, in)
}

// CreateReport 创建举报
func (c *interactionServiceClient) CreateReport(ctx context.Context, in *interaction_proto.CreateReportRequest, opts ...interface{}) (*interaction_proto.CreateReportResponse, error) {
	return c.client.CreateReport(ctx, in)
}

// CreateReply 创建回复
func (c *interactionServiceClient) CreateReply(ctx context.Context, in *interaction_proto.CreateReplyRequest, opts ...interface{}) (*interaction_proto.CreateReplyResponse, error) {
	return c.client.CreateReply(ctx, in)
}

// DeleteReply 删除回复
func (c *interactionServiceClient) DeleteReply(ctx context.Context, in *interaction_proto.DeleteReplyRequest, opts ...interface{}) (*interaction_proto.DeleteReplyResponse, error) {
	return c.client.DeleteReply(ctx, in)
}

// GetReplies 获取回复列表
func (c *interactionServiceClient) GetReplies(ctx context.Context, in *interaction_proto.GetRepliesRequest, opts ...interface{}) (*interaction_proto.GetRepliesResponse, error) {
	return c.client.GetReplies(ctx, in)
}

// GetPostStats 获取帖子统计
func (c *interactionServiceClient) GetPostStats(ctx context.Context, in *interaction_proto.GetPostStatsRequest, opts ...interface{}) (*interaction_proto.GetPostStatsResponse, error) {
	return c.client.GetPostStats(ctx, in)
}

// UpdatePostStats 更新帖子统计
func (c *interactionServiceClient) UpdatePostStats(ctx context.Context, in *interaction_proto.UpdatePostStatsRequest, opts ...interface{}) (*interaction_proto.UpdatePostStatsResponse, error) {
	return c.client.UpdatePostStats(ctx, in)
}

// VotePoll 投票
func (c *interactionServiceClient) VotePoll(ctx context.Context, in *interaction_proto.VotePollRequest, opts ...interface{}) (*interaction_proto.VotePollResponse, error) {
	return c.client.VotePoll(ctx, in)
}

// CreateBookmark 创建收藏
func (c *interactionServiceClient) CreateBookmark(ctx context.Context, in *interaction_proto.CreateBookmarkRequest, opts ...interface{}) (*interaction_proto.CreateBookmarkResponse, error) {
	return c.client.CreateBookmark(ctx, in)
}

// DeleteBookmark 删除收藏
func (c *interactionServiceClient) DeleteBookmark(ctx context.Context, in *interaction_proto.DeleteBookmarkRequest, opts ...interface{}) (*interaction_proto.DeleteBookmarkResponse, error) {
	return c.client.DeleteBookmark(ctx, in)
}

// IsBookmarked 检查是否收藏
func (c *interactionServiceClient) IsBookmarked(ctx context.Context, in *interaction_proto.IsBookmarkedRequest, opts ...interface{}) (*interaction_proto.IsBookmarkedResponse, error) {
	return c.client.IsBookmarked(ctx, in)
}

// GetBookmarks 获取收藏列表
func (c *interactionServiceClient) GetBookmarks(ctx context.Context, in *interaction_proto.GetBookmarksRequest, opts ...interface{}) (*interaction_proto.GetBookmarksResponse, error) {
	return c.client.GetBookmarks(ctx, in)
}

// LikePost 点赞帖子 (高级接口，包含统计更新)
func (c *interactionServiceClient) LikePost(ctx context.Context, in *interaction_proto.LikePostRequest, opts ...interface{}) (*interaction_proto.LikePostResponse, error) {
	return c.client.LikePost(ctx, in)
}

// UnlikePost 取消点赞帖子 (高级接口，包含统计更新)
func (c *interactionServiceClient) UnlikePost(ctx context.Context, in *interaction_proto.UnlikePostRequest, opts ...interface{}) (*interaction_proto.UnlikePostResponse, error) {
	return c.client.UnlikePost(ctx, in)
}

// recommendationServiceClient 推荐服务gRPC客户端实现
type recommendationServiceClient struct {
	client recommendation_proto.RecommendationServiceClient
}

// NewRecommendationServiceClient 创建推荐服务客户端
func NewRecommendationServiceClient(conn *grpc.ClientConn) RecommendationServiceClient {
	return &recommendationServiceClient{
		client: recommendation_proto.NewRecommendationServiceClient(conn),
	}
}

// GetRecommendations 获取推荐内容
func (c *recommendationServiceClient) GetRecommendations(ctx context.Context, in *recommendation_proto.GetRecommendationsRequest, opts ...interface{}) (*recommendation_proto.GetRecommendationsResponse, error) {
	return c.client.GetRecommendations(ctx, in)
}

// RecordUserAction 记录用户行为
func (c *recommendationServiceClient) RecordUserAction(ctx context.Context, in *recommendation_proto.RecordUserActionRequest, opts ...interface{}) (*recommendation_proto.RecordUserActionResponse, error) {
	return c.client.RecordUserAction(ctx, in)
}

// UpdateUserInterest 更新用户兴趣
func (c *recommendationServiceClient) UpdateUserInterest(ctx context.Context, in *recommendation_proto.UpdateUserInterestRequest, opts ...interface{}) (*recommendation_proto.UpdateUserInterestResponse, error) {
	return c.client.UpdateUserInterest(ctx, in)
}

// searchServiceClient 搜索服务gRPC客户端实现
type searchServiceClient struct {
	client search_proto.SearchServiceClient
}

// NewSearchServiceClient 创建搜索服务客户端
func NewSearchServiceClient(conn *grpc.ClientConn) SearchServiceClient {
	return &searchServiceClient{
		client: search_proto.NewSearchServiceClient(conn),
	}
}

// SearchContent 搜索内容
func (c *searchServiceClient) SearchContent(ctx context.Context, in *search_proto.SearchContentRequest, opts ...interface{}) (*search_proto.SearchContentResponse, error) {
	return c.client.SearchContent(ctx, in)
}

// SearchUsers 搜索用户
func (c *searchServiceClient) SearchUsers(ctx context.Context, in *search_proto.SearchUsersRequest, opts ...interface{}) (*search_proto.SearchUsersResponse, error) {
	return c.client.SearchUsers(ctx, in)
}

// SearchHashtags 搜索标签
func (c *searchServiceClient) SearchHashtags(ctx context.Context, in *search_proto.SearchHashtagsRequest, opts ...interface{}) (*search_proto.SearchHashtagsResponse, error) {
	return c.client.SearchHashtags(ctx, in)
}
