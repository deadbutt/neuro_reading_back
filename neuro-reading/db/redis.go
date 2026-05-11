package db

import (
	"context"
	"fmt"
	"neuro-reading/config"
	"time"

	"github.com/redis/go-redis/v9"
)

var Redis *redis.Client

func InitRedis(cfg *config.Config) error {
	Redis = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Redis.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis connection failed: %w", err)
	}

	return nil
}
