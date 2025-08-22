package resolver

import (
	"github.com/flick/backend/services/gateway/internal/client"
	"github.com/flick/backend/services/gateway/internal/graphql/model"
	user_proto "github.com/flick/backend/services/user/proto"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	AuthServiceClient        client.AuthServiceClient
	UserServiceClient        client.UserServiceClient
	ContentServiceClient     client.ContentServiceClient
	InteractionServiceClient client.InteractionServiceClient
}

// NewResolver creates a new resolver instance
func NewResolver(authServiceClient client.AuthServiceClient, userServiceClient client.UserServiceClient, contentServiceClient client.ContentServiceClient, interactionServiceClient client.InteractionServiceClient) *Resolver {
	return &Resolver{
		AuthServiceClient:        authServiceClient,
		UserServiceClient:        userServiceClient,
		ContentServiceClient:     contentServiceClient,
		InteractionServiceClient: interactionServiceClient,
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
