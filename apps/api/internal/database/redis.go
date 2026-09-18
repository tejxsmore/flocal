package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"flocal/internal/config"
)

func ConnectRedis(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("redis: parse REDIS_URL: %w", err)
	}

	client := redis.NewClient(opts)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("redis: ping failed: %w", err)
	}

	return client, nil
}
