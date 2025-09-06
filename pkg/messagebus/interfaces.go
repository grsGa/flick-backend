package messagebus

import (
	"context"
	"time"
)

// MessageBus defines the interface for message bus operations
type MessageBus interface {
	// Publish publishes a message to a topic
	Publish(ctx context.Context, topic string, message []byte) error
	
	// Subscribe subscribes to a topic and returns a channel for messages
	Subscribe(ctx context.Context, topic string) (<-chan Message, error)
	
	// Close closes the message bus connection
	Close() error
}

// Message represents a message received from the message bus
type Message struct {
	Topic     string
	Data      []byte
	Timestamp time.Time
}

// Event types for user profile updates
const (
	TopicUserProfileUpdated = "user.profile.updated"
	TopicUserAvatarUpdated  = "user.avatar.updated"
)

// UserProfileUpdatedEvent represents a user profile update event
type UserProfileUpdatedEvent struct {
	UserID        string    `json:"user_id"`
	Username      string    `json:"username"`
	DisplayName   string    `json:"display_name,omitempty"`
	AvatarURL     string    `json:"avatar_url,omitempty"`
	AvatarVersion int32     `json:"avatar_version,omitempty"`
	UpdatedAt     time.Time `json:"updated_at"`
	EventType     string    `json:"event_type"` // "avatar_updated", "profile_updated"
}
