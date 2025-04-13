package redis

import (
	"context"
	"fmt"
	"time"

	"backend/pkg/config"
	"github.com/go-redis/redis/v8"
)

// Client 封装了Redis客户端
type Client struct {
	*redis.Client
	ctx context.Context
}

// NewRedisClient 创建Redis客户端连接
func NewRedisClient(cfg interface{}) (*redis.Client, error) {
	// 将传入的配置转换为Config类型
	config, ok := cfg.(*config.Config)
	if !ok {
		return nil, fmt.Errorf("无效的配置类型")
	}
	
	// 使用配置中的Redis地址和密码
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddrs[0], // 使用第一个Redis地址
		Password: config.RedisPassword, // 从配置获取密码
		DB:       0,                   // 默认DB
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
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return rdb, nil
}

// Close 关闭Redis连接
func (c *Client) Close() error {
	return c.Client.Close()
}
