package idempotency

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const keyPrefix = "idempotency:"

// RedisStore persists idempotency records in Redis.
type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(redisURL string) (*RedisStore, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &RedisStore{client: client}, nil
}

func (s *RedisStore) Get(ctx context.Context, key string) (CachedResponse, error) {
	data, err := s.client.Get(ctx, keyPrefix+key).Bytes()
	if err == redis.Nil {
		return CachedResponse{}, ErrNotFound
	}
	if err != nil {
		return CachedResponse{}, err
	}
	var resp CachedResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return CachedResponse{}, err
	}
	return resp, nil
}

func (s *RedisStore) Set(ctx context.Context, key string, resp CachedResponse, ttl time.Duration) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, keyPrefix+key, data, ttl).Err()
}

func (s *RedisStore) Close() error {
	return s.client.Close()
}