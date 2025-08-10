package resolver

import (
	"backend/services/gateway/internal/client"
	"backend/services/gateway/internal/graphql/model"
	user_proto "backend/services/user/proto"
)

// Resolver is the root resolver.
type Resolver struct {
	AuthServiceClient client.AuthServiceClient
	UserServiceClient client.UserServiceClient
}

// NewResolver creates a new resolver.
func NewResolver(authServiceClient client.AuthServiceClient, userServiceClient client.UserServiceClient) *Resolver {
	return &Resolver{
		AuthServiceClient: authServiceClient,
		UserServiceClient: userServiceClient,
	}
}

func (r *Resolver) userProtoToGql(user *user_proto.User) *model.User {
	displayName := user.DisplayName
	avatarUrl := user.AvatarUrl
	bio := user.Bio
	bannerUrl := user.BannerUrl
	isFollowing := user.IsFollowing
	isVerified := user.IsVerified
	followersCount := int(user.FollowersCount)
	followingCount := int(user.FollowingCount)

	return &model.User{
		ID:             user.Id,
		Username:       user.Username,
		DisplayName:    &displayName,
		AvatarURL:      &avatarUrl,
		Bio:            &bio,
		BannerURL:      &bannerUrl,
		FollowersCount: followersCount,
		FollowingCount: followingCount,
		IsFollowing:    &isFollowing,
		IsVerified:     &isVerified,
		CreatedAt:      user.CreatedAt,
	}
}
