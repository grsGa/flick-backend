package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/flick/backend/pkg/config"
	"github.com/flick/backend/pkg/database"
	"github.com/flick/backend/pkg/discovery"
	"github.com/flick/backend/pkg/models"
	"github.com/flick/backend/services/interaction/internal/repository"
	"github.com/flick/backend/services/interaction/internal/service"
	"github.com/flick/backend/services/interaction/proto"
)

// interactionServiceServer gRPC服务器包装器
type interactionServiceServer struct {
	proto.UnimplementedInteractionServiceServer
	service *service.InteractionService
}

// CreateFollow 创建关注
func (s *interactionServiceServer) CreateFollow(ctx context.Context, req *proto.CreateFollowRequest) (*proto.CreateFollowResponse, error) {
	return s.service.CreateFollow(ctx, req)
}

// DeleteFollow 删除关注
func (s *interactionServiceServer) DeleteFollow(ctx context.Context, req *proto.DeleteFollowRequest) (*proto.DeleteFollowResponse, error) {
	return s.service.DeleteFollow(ctx, req)
}

// IsFollowing 检查是否关注
func (s *interactionServiceServer) IsFollowing(ctx context.Context, req *proto.IsFollowingRequest) (*proto.IsFollowingResponse, error) {
	return s.service.IsFollowing(ctx, req)
}

// GetFollowers 获取粉丝列表
func (s *interactionServiceServer) GetFollowers(ctx context.Context, req *proto.GetFollowersRequest) (*proto.GetFollowersResponse, error) {
	return s.service.GetFollowers(ctx, req)
}

// GetFollowing 获取关注列表
func (s *interactionServiceServer) GetFollowing(ctx context.Context, req *proto.GetFollowingRequest) (*proto.GetFollowingResponse, error) {
	return s.service.GetFollowing(ctx, req)
}

// CreateLike 创建点赞
func (s *interactionServiceServer) CreateLike(ctx context.Context, req *proto.CreateLikeRequest) (*proto.CreateLikeResponse, error) {
	return s.service.CreateLike(ctx, req)
}

// DeleteLike 删除点赞
func (s *interactionServiceServer) DeleteLike(ctx context.Context, req *proto.DeleteLikeRequest) (*proto.DeleteLikeResponse, error) {
	return s.service.DeleteLike(ctx, req)
}

// IsLiked 检查是否点赞
func (s *interactionServiceServer) IsLiked(ctx context.Context, req *proto.IsLikedRequest) (*proto.IsLikedResponse, error) {
	return s.service.IsLiked(ctx, req)
}

// GetLikes 获取点赞列表
func (s *interactionServiceServer) GetLikes(ctx context.Context, req *proto.GetLikesRequest) (*proto.GetLikesResponse, error) {
	return s.service.GetLikes(ctx, req)
}

// CreateRepost 创建转发
func (s *interactionServiceServer) CreateRepost(ctx context.Context, req *proto.CreateRepostRequest) (*proto.CreateRepostResponse, error) {
	return s.service.CreateRepost(ctx, req)
}

// DeleteRepost 删除转发
func (s *interactionServiceServer) DeleteRepost(ctx context.Context, req *proto.DeleteRepostRequest) (*proto.DeleteRepostResponse, error) {
	return s.service.DeleteRepost(ctx, req)
}

// CreateReply 创建回复
func (s *interactionServiceServer) CreateReply(ctx context.Context, req *proto.CreateReplyRequest) (*proto.CreateReplyResponse, error) {
	return s.service.CreateReply(ctx, req)
}

// DeleteReply 删除回复
func (s *interactionServiceServer) DeleteReply(ctx context.Context, req *proto.DeleteReplyRequest) (*proto.DeleteReplyResponse, error) {
	return s.service.DeleteReply(ctx, req)
}

// GetReplies 获取回复列表
func (s *interactionServiceServer) GetReplies(ctx context.Context, req *proto.GetRepliesRequest) (*proto.GetRepliesResponse, error) {
	return s.service.GetReplies(ctx, req)
}

// GetPostStats 获取帖子统计
func (s *interactionServiceServer) GetPostStats(ctx context.Context, req *proto.GetPostStatsRequest) (*proto.GetPostStatsResponse, error) {
	return s.service.GetPostStats(ctx, req)
}

// UpdatePostStats 更新帖子统计
func (s *interactionServiceServer) UpdatePostStats(ctx context.Context, req *proto.UpdatePostStatsRequest) (*proto.UpdatePostStatsResponse, error) {
	return s.service.UpdatePostStats(ctx, req)
}

// VotePoll 投票
func (s *interactionServiceServer) VotePoll(ctx context.Context, req *proto.VotePollRequest) (*proto.VotePollResponse, error) {
	return s.service.VotePoll(ctx, req)
}

// CreateBookmark 创建收藏
func (s *interactionServiceServer) CreateBookmark(ctx context.Context, req *proto.CreateBookmarkRequest) (*proto.CreateBookmarkResponse, error) {
	return s.service.CreateBookmark(ctx, req)
}

// DeleteBookmark 删除收藏
func (s *interactionServiceServer) DeleteBookmark(ctx context.Context, req *proto.DeleteBookmarkRequest) (*proto.DeleteBookmarkResponse, error) {
	return s.service.DeleteBookmark(ctx, req)
}

// IsBookmarked 检查是否收藏
func (s *interactionServiceServer) IsBookmarked(ctx context.Context, req *proto.IsBookmarkedRequest) (*proto.IsBookmarkedResponse, error) {
	return s.service.IsBookmarked(ctx, req)
}

// GetBookmarks 获取收藏列表
func (s *interactionServiceServer) GetBookmarks(ctx context.Context, req *proto.GetBookmarksRequest) (*proto.GetBookmarksResponse, error) {
	return s.service.GetBookmarks(ctx, req)
}

// CreateReport 创建举报
func (s *interactionServiceServer) CreateReport(ctx context.Context, req *proto.CreateReportRequest) (*proto.CreateReportResponse, error) {
	return s.service.CreateReport(ctx, req)
}

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库连接并执行迁移
	if err := database.InitDB(cfg, true); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 执行interaction相关的数据库迁移
	db := database.GetDB()
	if err := db.AutoMigrate(
		&models.Follow{},
		&models.Like{},
		&models.Repost{},
		&models.PostStats{},
		&models.PollVote{},
		&models.Bookmark{},
		&models.Report{},
	); err != nil {
		log.Fatalf("Failed to migrate interaction models: %v", err)
	}

	// 初始化仓库
	interactionRepo := repository.NewInteractionRepository()

	// 初始化服务
	interactionService := service.NewInteractionService(interactionRepo)

	port := os.Getenv("INTERACTION_SERVICE_PORT")
	if port == "" {
		port = "50056"
	}

	// 创建监听器
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	// 创建gRPC服务器实例
	s := grpc.NewServer()

	// 注册服务
	proto.RegisterInteractionServiceServer(s, &interactionServiceServer{service: interactionService})

	// 启用反射（用于调试）
	reflection.Register(s)

	// 解析端口用于Consul注册
	portInt, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("Invalid port: %v", err)
	}

	// 注册到Consul
	discovery.RegisterServiceToConsul(discovery.RegisterOptions{
		ServiceName:     "interaction-service",
		ServicePort:     portInt,
		HealthCheckType: "grpc",
	})

	// 设置优雅关闭
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Shutting down interaction service...")
		s.GracefulStop()
	}()

	log.Printf("Starting interaction service on port %s", port)

	// 启动服务器
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve interaction service: %v", err)
	}
}
