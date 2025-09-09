package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

// ReplyEvent 回复事件结构
type ReplyEvent struct {
	Type      string    `json:"type"`      // "reply_created", "reply_deleted"
	ReplyID   string    `json:"reply_id"`
	PostID    string    `json:"post_id"`    // 被回复的帖子ID
	RootID    string    `json:"root_id"`    // 根帖子ID
	ParentID  string    `json:"parent_id"`  // 父回复ID
	UserID    string    `json:"user_id"`
	Level     int       `json:"level"`      // 回复层级
	Timestamp time.Time `json:"timestamp"`
}

// ReplyStatsUpdateEvent 回复统计更新事件
type ReplyStatsUpdateEvent struct {
	PostID      string `json:"post_id"`
	ReplyCount  int32  `json:"reply_count"`
	Increment   bool   `json:"increment"`   // true为增加，false为减少
	Timestamp   time.Time `json:"timestamp"`
}

// ReplyEventPublisher 回复事件发布器
type ReplyEventPublisher struct {
	nc *nats.Conn
}

// NewReplyEventPublisher 创建回复事件发布器
func NewReplyEventPublisher(nc *nats.Conn) *ReplyEventPublisher {
	return &ReplyEventPublisher{nc: nc}
}

// PublishReplyCreated 发布回复创建事件
func (p *ReplyEventPublisher) PublishReplyCreated(ctx context.Context, replyID, postID, rootID, parentID, userID string, level int) error {
	event := ReplyEvent{
		Type:      "reply_created",
		ReplyID:   replyID,
		PostID:    postID,
		RootID:    rootID,
		ParentID:  parentID,
		UserID:    userID,
		Level:     level,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal reply created event: %w", err)
	}

	// 发布到多个主题以支持不同的消费者
	subjects := []string{
		"content.reply.created",
		fmt.Sprintf("content.post.%s.reply.created", postID),
		"interaction.stats.update", // 通知统计服务更新
	}

	for _, subject := range subjects {
		if err := p.nc.Publish(subject, data); err != nil {
			log.Printf("[Reply Events] Failed to publish to %s: %v", subject, err)
		}
	}

	log.Printf("[Reply Events] Published reply created event: %s -> %s", replyID, postID)
	return nil
}

// PublishReplyDeleted 发布回复删除事件
func (p *ReplyEventPublisher) PublishReplyDeleted(ctx context.Context, replyID, postID, rootID, parentID, userID string, level int) error {
	event := ReplyEvent{
		Type:      "reply_deleted",
		ReplyID:   replyID,
		PostID:    postID,
		RootID:    rootID,
		ParentID:  parentID,
		UserID:    userID,
		Level:     level,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal reply deleted event: %w", err)
	}

	subjects := []string{
		"content.reply.deleted",
		fmt.Sprintf("content.post.%s.reply.deleted", postID),
		"interaction.stats.update",
	}

	for _, subject := range subjects {
		if err := p.nc.Publish(subject, data); err != nil {
			log.Printf("[Reply Events] Failed to publish to %s: %v", subject, err)
		}
	}

	log.Printf("[Reply Events] Published reply deleted event: %s -> %s", replyID, postID)
	return nil
}

// PublishStatsUpdate 发布统计更新事件
func (p *ReplyEventPublisher) PublishStatsUpdate(ctx context.Context, postID string, increment bool) error {
	event := ReplyStatsUpdateEvent{
		PostID:    postID,
		Increment: increment,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal stats update event: %w", err)
	}

	if err := p.nc.Publish("interaction.reply.stats.update", data); err != nil {
		return fmt.Errorf("failed to publish stats update event: %w", err)
	}

	log.Printf("[Reply Events] Published stats update event for post %s (increment: %v)", postID, increment)
	return nil
}

// ReplyEventHandler 回复事件处理器接口
type ReplyEventHandler interface {
	HandleReplyCreated(ctx context.Context, event *ReplyEvent) error
	HandleReplyDeleted(ctx context.Context, event *ReplyEvent) error
}

// ReplyEventSubscriber 回复事件订阅器
type ReplyEventSubscriber struct {
	nc      *nats.Conn
	handler ReplyEventHandler
}

// NewReplyEventSubscriber 创建回复事件订阅器
func NewReplyEventSubscriber(nc *nats.Conn, handler ReplyEventHandler) *ReplyEventSubscriber {
	return &ReplyEventSubscriber{
		nc:      nc,
		handler: handler,
	}
}

// Subscribe 订阅回复事件
func (s *ReplyEventSubscriber) Subscribe(ctx context.Context) error {
	// 订阅回复创建事件
	_, err := s.nc.Subscribe("content.reply.created", func(msg *nats.Msg) {
		var event ReplyEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("[Reply Events] Failed to unmarshal reply created event: %v", err)
			return
		}

		if err := s.handler.HandleReplyCreated(ctx, &event); err != nil {
			log.Printf("[Reply Events] Failed to handle reply created event: %v", err)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to reply created events: %w", err)
	}

	// 订阅回复删除事件
	_, err = s.nc.Subscribe("content.reply.deleted", func(msg *nats.Msg) {
		var event ReplyEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("[Reply Events] Failed to unmarshal reply deleted event: %v", err)
			return
		}

		if err := s.handler.HandleReplyDeleted(ctx, &event); err != nil {
			log.Printf("[Reply Events] Failed to handle reply deleted event: %v", err)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to reply deleted events: %w", err)
	}

	log.Println("[Reply Events] Successfully subscribed to reply events")
	return nil
}
