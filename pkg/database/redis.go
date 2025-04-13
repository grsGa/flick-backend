package database

import (
	"context"
	"time"

	"backend/pkg/config"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RedisManager 管理Redis连接
type RedisManager struct {
	Client *redis.ClusterClient
	logger *zap.Logger
	ctx    context.Context
}

// NewRedisManager 创建Redis管理器实例
func NewRedisManager() (*RedisManager, error) {
	cfg := config.GetConfig()
	log := config.GetLogger()
	ctx := context.Background()

	// 创建Redis集群客户端
	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    cfg.RedisAddrs,
		Password: cfg.RedisPassword,
		// 连接池配置
		PoolSize:     50,
		MinIdleConns: 10,
		// 连接超时设置
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		// 命令重试
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
	})

	// 测试连接
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Error("Failed to connect to Redis", zap.Error(err), zap.Strings("addrs", cfg.RedisAddrs))
		return nil, err
	}

	log.Info("Redis connection established", zap.Strings("addrs", cfg.RedisAddrs))

	return &RedisManager{
		Client: rdb,
		logger: log,
		ctx:    ctx,
	}, nil
}

// Close 关闭Redis连接
func (rm *RedisManager) Close() error {
	err := rm.Client.Close()
	if err != nil {
		rm.logger.Error("Failed to close Redis connection", zap.Error(err))
		return err
	}
	rm.logger.Info("Redis connection closed")
	return nil
}

// GetKey 获取键值
func (rm *RedisManager) GetKey(key string) (string, error) {
	val, err := rm.Client.Get(rm.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // 键不存在
		}
		rm.logger.Error("Failed to get Redis key", zap.String("key", key), zap.Error(err))
		return "", err
	}
	return val, nil
}

// SetKey 设置键值
func (rm *RedisManager) SetKey(key string, value interface{}, expiration time.Duration) error {
	err := rm.Client.Set(rm.ctx, key, value, expiration).Err()
	if err != nil {
		rm.logger.Error("Failed to set Redis key", 
			zap.String("key", key), 
			zap.Any("value", value), 
			zap.Duration("expiration", expiration),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// DelKey 删除键
func (rm *RedisManager) DelKey(key string) error {
	err := rm.Client.Del(rm.ctx, key).Err()
	if err != nil {
		rm.logger.Error("Failed to delete Redis key", zap.String("key", key), zap.Error(err))
		return err
	}
	return nil
}

// HasKey 检查键是否存在
func (rm *RedisManager) HasKey(key string) (bool, error) {
	val, err := rm.Client.Exists(rm.ctx, key).Result()
	if err != nil {
		rm.logger.Error("Failed to check if Redis key exists", zap.String("key", key), zap.Error(err))
		return false, err
	}
	return val > 0, nil
}

// IncrementKey 递增键值
func (rm *RedisManager) IncrementKey(key string) (int64, error) {
	val, err := rm.Client.Incr(rm.ctx, key).Result()
	if err != nil {
		rm.logger.Error("Failed to increment Redis key", zap.String("key", key), zap.Error(err))
		return 0, err
	}
	return val, nil
}

// SetKeyWithTTL 设置带过期时间的键值
func (rm *RedisManager) SetKeyWithTTL(key string, value interface{}, ttl time.Duration) error {
	return rm.SetKey(key, value, ttl)
}

// HashSet 设置哈希表字段
func (rm *RedisManager) HashSet(key, field string, value interface{}) error {
	err := rm.Client.HSet(rm.ctx, key, field, value).Err()
	if err != nil {
		rm.logger.Error("Failed to set hash field", 
			zap.String("key", key), 
			zap.String("field", field),
			zap.Any("value", value),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// HashGet 获取哈希表字段
func (rm *RedisManager) HashGet(key, field string) (string, error) {
	val, err := rm.Client.HGet(rm.ctx, key, field).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // 字段不存在
		}
		rm.logger.Error("Failed to get hash field", 
			zap.String("key", key),
			zap.String("field", field),
			zap.Error(err),
		)
		return "", err
	}
	return val, nil
}

// HashGetAll 获取哈希表所有字段
func (rm *RedisManager) HashGetAll(key string) (map[string]string, error) {
	val, err := rm.Client.HGetAll(rm.ctx, key).Result()
	if err != nil {
		rm.logger.Error("Failed to get all hash fields", zap.String("key", key), zap.Error(err))
		return nil, err
	}
	return val, nil
} 