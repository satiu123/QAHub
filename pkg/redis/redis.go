package redis

import (
	"context"

	"qahub/pkg/config"

	"github.com/redis/go-redis/v9"
)

// NewClient 根据提供的配置创建一个新的 Redis 客户端
func NewClient(cfg config.Redis) (*redis.Client, error) {
	// 创建 Redis 客户端实例
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 使用 Ping 命令检查连接是否成功
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
