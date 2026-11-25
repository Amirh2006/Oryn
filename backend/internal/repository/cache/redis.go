package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/Amirh2006/Oryn/backend/internal/config"
	"github.com/redis/go-redis/v9"
)

// RedisCache provides structured JSON caching using Redis
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new RedisCache instance
func NewRedisCache(cfg config.RedisConfig) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})

	return &RedisCache{client: client}
}

// CheckReadiness implements handlers.ReadinessChecker
func (r *RedisCache) CheckReadiness(ctx context.Context) error {
	if r.client == nil {
		return fmt.Errorf("redis client is uninitialized")
	}
	return r.client.Ping(ctx).Err()
}

// Get fetches and deserializes a cached value into target
func (r *RedisCache) Get(ctx context.Context, key string, target interface{}) bool {
	if r.client == nil {
		return false
	}

	val, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err != redis.Nil {
			slog.Debug("Redis cache get miss/error", slog.String("key", key), slog.String("error", err.Error()))
		}
		return false
	}

	if err := json.Unmarshal(val, target); err != nil {
		slog.Warn("Failed to unmarshal cached JSON", slog.String("key", key), slog.String("error", err.Error()))
		return false
	}

	return true
}

// Set serializes and stores a value with TTL
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) {
	if r.client == nil {
		return
	}

	bytes, err := json.Marshal(value)
	if err != nil {
		slog.Warn("Failed to marshal value for caching", slog.String("key", key), slog.String("error", err.Error()))
		return
	}

	if err := r.client.Set(ctx, key, bytes, ttl).Err(); err != nil {
		slog.Warn("Failed to set cache in Redis", slog.String("key", key), slog.String("error", err.Error()))
	}
}

// Close closes the underlying Redis client connection
func (r *RedisCache) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}
