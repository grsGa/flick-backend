package resolver

import (
	"context"
	"fmt"

	"github.com/flick/backend/pkg/messagebus"
	"github.com/flick/backend/services/gateway/internal/client"
	"github.com/flick/backend/services/gateway/internal/graphql/model"
	"github.com/flick/backend/services/gateway/internal/middleware"
	user_proto "github.com/flick/backend/services/user/proto"
	"google.golang.org/grpc/metadata"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	AuthServiceClient        client.AuthServiceClient
	UserServiceClient        client.UserServiceClient
	ContentServiceClient     client.ContentServiceClient
	InteractionServiceClient client.InteractionServiceClient
	MediaServiceClient       client.MediaServiceClient
	MessageBus               messagebus.MessageBus
}

// NewResolver creates a new resolver instance
func NewResolver(authServiceClient client.AuthServiceClient, userServiceClient client.UserServiceClient, contentServiceClient client.ContentServiceClient, interactionServiceClient client.InteractionServiceClient, mediaServiceClient client.MediaServiceClient, messageBus messagebus.MessageBus) *Resolver {
	return &Resolver{
		AuthServiceClient:        authServiceClient,
		UserServiceClient:        userServiceClient,
		ContentServiceClient:     contentServiceClient,
		InteractionServiceClient: interactionServiceClient,
		MediaServiceClient:       mediaServiceClient,
		MessageBus:               messageBus,
	}
}

// userProtoToGql converts user proto to GraphQL model
func (r *Resolver) userProtoToGql(user *user_proto.User) *model.User {
	if user == nil {
		return nil
	}

	// Convert string to *string for optional fields
	var displayName, bio, location, website, avatarUrl, bannerUrl *string
	if user.DisplayName != "" {
		displayName = &user.DisplayName
	}
	if user.Bio != "" {
		bio = &user.Bio
	}
	if user.Location != "" {
		location = &user.Location
	}
	if user.WebsiteUrl != "" {
		website = &user.WebsiteUrl
	}
	if user.AvatarUrl != "" {
		avatarUrl = &user.AvatarUrl
	}
	if user.BannerUrl != "" {
		bannerUrl = &user.BannerUrl
	}

	// Convert bool to *bool for optional fields
	var isVerified *bool
	isFollowing := &user.IsFollowing // Always set isFollowing, regardless of value
	if user.IsVerified {
		isVerified = &user.IsVerified
	}

	return &model.User{
		ID:             user.Id,
		Username:       user.Username,
		DisplayName:    displayName,
		Bio:            bio,
		Location:       location,
		Website:        website,
		AvatarURL:      avatarUrl,
		BannerURL:      bannerUrl,
		FollowersCount: int(user.FollowersCount),
		FollowingCount: int(user.FollowingCount),
		IsFollowing:    isFollowing,
		IsVerified:     isVerified,
		CreatedAt:      user.CreatedAt,
	}
}

// createAuthenticatedContext creates a context with JWT token for gRPC calls
func (r *Resolver) createAuthenticatedContext(ctx context.Context) (context.Context, error) {
	// Extract token from GraphQL context using middleware helper
	token := middleware.GetTokenFromContext(ctx)
	if token == "" {
		return nil, fmt.Errorf("authentication required: no valid token found")
	}

	// Create gRPC metadata with Authorization header
	md := metadata.Pairs("authorization", "Bearer "+token)
	return metadata.NewOutgoingContext(ctx, md), nil
}
