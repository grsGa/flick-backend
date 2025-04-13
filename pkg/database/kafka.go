package database

import (
	"context"
	"fmt"
	"time"

	"backend/pkg/config"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// KafkaManager 管理Kafka连接
type KafkaManager struct {
	Writers map[string]*kafka.Writer
	Readers map[string]*kafka.Reader
	Brokers []string
	logger  *zap.Logger
	ctx     context.Context
}

// NewKafkaManager 创建Kafka管理器实例
func NewKafkaManager() (*KafkaManager, error) {
	cfg := config.GetConfig()
	log := config.GetLogger()
	ctx := context.Background()

	// 创建Kafka管理器
	km := &KafkaManager{
		Writers: make(map[string]*kafka.Writer),
		Readers: make(map[string]*kafka.Reader),
		Brokers: cfg.KafkaBrokers,
		logger:  log,
		ctx:     ctx,
	}

	// 为每个主题创建Writer
	for topicKey, topicName := range cfg.KafkaTopics {
		km.Writers[topicKey] = kafka.NewWriter(kafka.WriterConfig{
			Brokers:      cfg.KafkaBrokers,
			Topic:        topicName,
			Balancer:     &kafka.LeastBytes{},
			WriteTimeout: 10 * time.Second,
			// 批量写入配置
			BatchSize:    100,                    // 最多100条消息一批
			BatchTimeout: 100 * time.Millisecond, // 或者等待100ms
			// 重试配置
			MaxAttempts:  3, // 最多重试3次
			RequiredAcks: 1, // 至少一个broker确认
		})
		log.Info("Kafka writer created", zap.String("topic", topicName))
	}

	// 测试连接
	// 我们只需要测试一个连接，因为所有连接都使用相同的brokers
	if len(cfg.KafkaBrokers) > 0 {
		dialer := &kafka.Dialer{
			Timeout:   10 * time.Second,
			DualStack: true,
		}

		conn, err := dialer.DialContext(ctx, "tcp", cfg.KafkaBrokers[0])
		if err != nil {
			log.Error("Failed to connect to Kafka", zap.Error(err), zap.Strings("brokers", cfg.KafkaBrokers))
			return nil, err
		}
		conn.Close()
		log.Info("Kafka connection established", zap.Strings("brokers", cfg.KafkaBrokers))
	}

	return km, nil
}

// Close 关闭Kafka连接
func (km *KafkaManager) Close() {
	// 关闭所有Writer
	for topicKey, writer := range km.Writers {
		if err := writer.Close(); err != nil {
			km.logger.Error("Failed to close Kafka writer",
				zap.String("topic", topicKey),
				zap.Error(err),
			)
		}
	}

	// 关闭所有Reader
	for topicKey, reader := range km.Readers {
		if err := reader.Close(); err != nil {
			km.logger.Error("Failed to close Kafka reader",
				zap.String("topic", topicKey),
				zap.Error(err),
			)
		}
	}

	km.logger.Info("Kafka connections closed")
}

// GetWriter 获取指定主题的Writer
func (km *KafkaManager) GetWriter(topicKey string) (*kafka.Writer, error) {
	writer, exists := km.Writers[topicKey]
	if !exists {
		return nil, fmt.Errorf("writer for topic %s not found", topicKey)
	}
	return writer, nil
}

// CreateReader 创建新的Reader实例
func (km *KafkaManager) CreateReader(topicKey, groupID string) (*kafka.Reader, error) {
	cfg := config.GetConfig()

	topicName, exists := cfg.KafkaTopics[topicKey]
	if !exists {
		return nil, fmt.Errorf("topic %s not configured", topicKey)
	}

	// 如果没有提供groupID，使用配置中的默认值
	if groupID == "" {
		groupID = cfg.KafkaGroupID
	}

	// 创建Reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:         km.Brokers,
		Topic:           topicName,
		GroupID:         groupID,
		MinBytes:        10e3, // 10KB
		MaxBytes:        10e6, // 10MB
		MaxWait:         500 * time.Millisecond,
		ReadLagInterval: -1,                // 禁用滞后报告
		CommitInterval:  0,                 // 自动提交间隔
		StartOffset:     kafka.FirstOffset, // 从头开始读取
	})

	// 存储Reader引用
	km.Readers[topicKey] = reader

	km.logger.Info("Kafka reader created",
		zap.String("topic", topicName),
		zap.String("groupID", groupID),
	)

	return reader, nil
}

// PublishMessage 发布消息到指定主题
func (km *KafkaManager) PublishMessage(topicKey string, key string, value []byte) error {
	writer, err := km.GetWriter(topicKey)
	if err != nil {
		return err
	}

	message := kafka.Message{
		Key:   []byte(key),
		Value: value,
		Time:  time.Now(),
	}

	if err := writer.WriteMessages(km.ctx, message); err != nil {
		km.logger.Error("Failed to publish message to Kafka",
			zap.String("topic", topicKey),
			zap.String("key", key),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// PublishMessages 批量发布消息到指定主题
func (km *KafkaManager) PublishMessages(topicKey string, messages []kafka.Message) error {
	writer, err := km.GetWriter(topicKey)
	if err != nil {
		return err
	}

	if err := writer.WriteMessages(km.ctx, messages...); err != nil {
		km.logger.Error("Failed to publish messages to Kafka",
			zap.String("topic", topicKey),
			zap.Int("count", len(messages)),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// ConsumeMessages 消费指定主题的消息
func (km *KafkaManager) ConsumeMessages(ctx context.Context, topicKey, groupID string, handler func(message kafka.Message) error) error {
	// 创建Reader
	reader, err := km.CreateReader(topicKey, groupID)
	if err != nil {
		return err
	}

	// 在新goroutine中消费消息
	go func() {
		defer reader.Close()

		for {
			select {
			case <-ctx.Done():
				km.logger.Info("Context canceled, stopping message consumption",
					zap.String("topic", topicKey),
					zap.String("groupID", groupID),
				)
				return
			default:
				message, err := reader.ReadMessage(ctx)
				if err != nil {
					if ctx.Err() != nil {
						// 上下文取消，不记录错误
						return
					}
					km.logger.Error("Failed to read message from Kafka",
						zap.String("topic", topicKey),
						zap.String("groupID", groupID),
						zap.Error(err),
					)
					// 短暂暂停，避免过度记录错误
					time.Sleep(time.Second)
					continue
				}

				// 处理消息
				if err := handler(message); err != nil {
					km.logger.Error("Failed to process message",
						zap.String("topic", topicKey),
						zap.String("key", string(message.Key)),
						zap.Error(err),
					)
				}
			}
		}
	}()

	return nil
}
