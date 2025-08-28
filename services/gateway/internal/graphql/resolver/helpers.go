package resolver

import (
	content_proto "github.com/flick/backend/services/content/proto"
	"github.com/flick/backend/services/gateway/internal/graphql/model"
)

// postProtoToGql converts content service proto Post to GraphQL model
func (r *Resolver) postProtoToGql(post *content_proto.Post) *model.Post {
	if post == nil {
		return nil
	}

	// Convert media attachments
	mediaAttachments := make([]model.MediaAttachment, len(post.MediaAttachments))
	for i, media := range post.MediaAttachments {
		mediaAttachments[i] = model.MediaAttachment{
			ID:   media.Id,
			URL:  media.Url,
			Type: media.Type,
		}
	}

	// Convert poll if exists
	var poll *model.Poll
	if post.Poll != nil {
		pollOptions := make([]model.PollOption, len(post.Poll.Options))
		for i, option := range post.Poll.Options {
			pollOptions[i] = model.PollOption{
				ID:        option.Id,
				Text:      option.Text,
				VoteCount: int(option.VoteCount),
			}
		}

		poll = &model.Poll{
			ID:              post.Poll.Id,
			Question:        post.Poll.Question,
			Options:         pollOptions,
			DurationMinutes: int(post.Poll.DurationMinutes),
			ExpiresAt:       post.Poll.ExpiresAt,
		}
	}

	// Convert post stats - always provide default values
	stats := &model.PostStats{
		LikeCount:   0,
		ReplyCount:  0,
		RepostCount: 0,
		ViewCount:   0,
	}
	if post.Stats != nil {
		stats.LikeCount = int(post.Stats.LikeCount)
		stats.ReplyCount = int(post.Stats.ReplyCount)
		stats.RepostCount = int(post.Stats.RepostCount)
		stats.ViewCount = int(post.Stats.ViewCount)
	}

	// Convert visibility enum
	visibility := model.PostVisibilityPublic // default
	switch post.Visibility {
	case "public":
		visibility = model.PostVisibilityPublic
	case "private":
		visibility = model.PostVisibilityPrivate
	case "followers":
		visibility = model.PostVisibilityFollowers
	}

	// Convert reply permission enum
	replyPermission := model.ReplyPermissionEveryone // default
	switch post.ReplyPermission {
	case "EVERYONE":
		replyPermission = model.ReplyPermissionEveryone
	case "FOLLOWING":
		replyPermission = model.ReplyPermissionFollowing
	case "MENTIONED_ONLY":
		replyPermission = model.ReplyPermissionMentionedOnly
	}

	// Convert author information - ensure all required fields are present
	author := &model.User{
		ID:             post.UserId,
		Username:       "user_" + post.UserId, // Default fallback
		FollowersCount: 0,
		FollowingCount: 0,
		CreatedAt:      post.CreatedAt, // Use post creation time as fallback
	}

	if post.Author != nil {
		author.ID = post.Author.Id
		author.Username = post.Author.Username
		if post.Author.DisplayName != "" {
			displayName := post.Author.DisplayName
			author.DisplayName = &displayName
		}
		if post.Author.AvatarUrl != "" {
			avatarUrl := post.Author.AvatarUrl
			author.AvatarURL = &avatarUrl
		}
		if post.Author.IsVerified {
			isVerified := post.Author.IsVerified
			author.IsVerified = &isVerified
		}
	}

	// Convert media attachments to Media format for frontend
	media := make([]model.Media, len(post.MediaAttachments))
	for i, attachment := range post.MediaAttachments {
		mediaType := model.MediaTypeImage
		if attachment.Type == "video" {
			mediaType = model.MediaTypeVideo
		}

		// 构建variants信息
		var variants *model.MediaVariants
		if attachment.Variants != nil {
			variants = &model.MediaVariants{}

			if attachment.Variants.Thumbnail != nil {
				variants.Thumbnail = &model.MediaVariant{
					URL:    attachment.Variants.Thumbnail.Url,
					Width:  int(attachment.Variants.Thumbnail.Width),
					Height: int(attachment.Variants.Thumbnail.Height),
					Size:   int(attachment.Variants.Thumbnail.Size),
				}
			}
			if attachment.Variants.Small != nil {
				variants.Small = &model.MediaVariant{
					URL:    attachment.Variants.Small.Url,
					Width:  int(attachment.Variants.Small.Width),
					Height: int(attachment.Variants.Small.Height),
					Size:   int(attachment.Variants.Small.Size),
				}
			}
			if attachment.Variants.Medium != nil {
				variants.Medium = &model.MediaVariant{
					URL:    attachment.Variants.Medium.Url,
					Width:  int(attachment.Variants.Medium.Width),
					Height: int(attachment.Variants.Medium.Height),
					Size:   int(attachment.Variants.Medium.Size),
				}
			}
			if attachment.Variants.Large != nil {
				variants.Large = &model.MediaVariant{
					URL:    attachment.Variants.Large.Url,
					Width:  int(attachment.Variants.Large.Width),
					Height: int(attachment.Variants.Large.Height),
					Size:   int(attachment.Variants.Large.Size),
				}
			}
			if attachment.Variants.Original != nil {
				variants.Original = &model.MediaVariant{
					URL:    attachment.Variants.Original.Url,
					Width:  int(attachment.Variants.Original.Width),
					Height: int(attachment.Variants.Original.Height),
					Size:   int(attachment.Variants.Original.Size),
				}
			}
		}

		// 构建完整的Media对象
		var mimeType *string
		if attachment.MimeType != "" {
			mimeType = &attachment.MimeType
		}
		var width, height *int
		if attachment.Width > 0 {
			w := int(attachment.Width)
			width = &w
		}
		if attachment.Height > 0 {
			h := int(attachment.Height)
			height = &h
		}

		media[i] = model.Media{
			ID:       attachment.Id,
			URL:      attachment.Url,
			Type:     mediaType,
			MimeType: mimeType,
			Width:    width,
			Height:   height,
			Variants: variants,
		}
	}

	// Create interaction object from stats
	interaction := &model.Interaction{
		IsLiked:      false, // TODO: Get actual user interaction status
		IsBookmarked: false, // TODO: Get actual user interaction status
		IsReposted:   false, // TODO: Get actual user interaction status
		LikeCount:    stats.LikeCount,
		ReplyCount:   stats.ReplyCount,
		RepostCount:  stats.RepostCount,
	}

	return &model.Post{
		ID:               post.Id,
		Content:          post.Content,
		Author:           author,
		Visibility:       visibility,
		ReplyPermission:  replyPermission,
		ParentID:         stringToPointer(post.ParentId),
		RootID:           stringToPointer(post.RootId),
		RepostID:         stringToPointer(post.RepostId),
		IsReply:          post.IsReply,
		ReplyLevel:       int(post.ReplyLevel),
		HasMedia:         len(post.MediaAttachments) > 0,
		HasPoll:          post.Poll != nil,
		Media:            media,
		MediaAttachments: mediaAttachments,
		MentionedUsers:   post.MentionedUsers,
		Tags:             post.Tags,
		Poll:             poll,
		Stats:            stats,
		Interaction:      interaction,
		CreatedAt:        post.CreatedAt,
		UpdatedAt:        post.UpdatedAt,
	}
}

// Helper function to convert string to pointer
func stringToPointer(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
