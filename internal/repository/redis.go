package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisCacheRepo struct {
	client *redis.Client
}

func NewRedisCacheRepository(client *redis.Client) CacheRepository {
	return &redisCacheRepo{client: client}
}

func (r *redisCacheRepo) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	} else if err != nil {
		return "", fmt.Errorf("redis get failed: %w", err)
	}
	return val, nil
}

func (r *redisCacheRepo) Set(ctx context.Context, key string, value string, ttlSeconds int) error {
	err := r.client.Set(ctx, key, value, time.Duration(ttlSeconds)*time.Second).Err()
	if err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}
	return nil
}

func (r *redisCacheRepo) Incr(ctx context.Context, key string, ttlSeconds int) (int64, error) {
	pipe := r.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, time.Duration(ttlSeconds)*time.Second)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("redis incr pipeline failed: %w", err)
	}

	return incr.Val(), nil
}
