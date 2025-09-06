package messagebus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// redisMessageBus implements MessageBus using Redis pub/sub
type redisMessageBus struct {
	client *redis.Client
}

// NewRedisMessageBus creates a new Redis-based message bus
func NewRedisMessageBus(redisURL string) (MessageBus, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &redisMessageBus{
		client: client,
	}, nil
}

// Publish publishes a message to a Redis topic
func (r *redisMessageBus) Publish(ctx context.Context, topic string, message []byte) error {
	return r.client.Publish(ctx, topic, message).Err()
}

// Subscribe subscribes to a Redis topic and returns a channel for messages
func (r *redisMessageBus) Subscribe(ctx context.Context, topic string) (<-chan Message, error) {
	pubsub := r.client.Subscribe(ctx, topic)

	// Test subscription
	_, err := pubsub.Receive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
	}

	msgChan := make(chan Message, 100) // Buffer for messages

	go func() {
		defer close(msgChan)
		defer pubsub.Close()

		ch := pubsub.Channel()
		for {
			select {
			case msg := <-ch:
				if msg == nil {
					return
				}
				msgChan <- Message{
					Topic:     msg.Channel,
					Data:      []byte(msg.Payload),
					Timestamp: time.Now(),
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return msgChan, nil
}

// Close closes the Redis connection
func (r *redisMessageBus) Close() error {
	return r.client.Close()
}

// PublishUserProfileUpdated publishes a user profile updated event
func PublishUserProfileUpdated(ctx context.Context, mb MessageBus, event *UserProfileUpdatedEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	return mb.Publish(ctx, TopicUserProfileUpdated, data)
}

// PublishUserAvatarUpdated publishes a user avatar updated event
func PublishUserAvatarUpdated(ctx context.Context, mb MessageBus, userID, username, avatarURL string, avatarVersion int32) error {
	event := &UserProfileUpdatedEvent{
		UserID:        userID,
		Username:      username,
		AvatarURL:     avatarURL,
		AvatarVersion: avatarVersion,
		UpdatedAt:     time.Now(),
		EventType:     "avatar_updated",
	}

	return PublishUserProfileUpdated(ctx, mb, event)
}
