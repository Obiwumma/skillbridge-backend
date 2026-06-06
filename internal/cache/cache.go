// Package cache manages the memory-tier caching cluster powered by Redis.
package cache

import (
	"context"
	"skillbridge-backend/pkg/config"
	"skillbridge-backend/pkg/logger"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient is the active Redis caching broker instance.
var RedisClient *redis.Client

// Init establishes connection registers and Pings the caching cluster.
func Init() {
	cfg := config.AppConfig

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		logger.Error("Failed to connect to Redis", err)
		if cfg.Env == "production" {
			panic(err)
		}
	} else {
		logger.Info("Successfully connected to Redis")
	}
}

// Set writes a key-value record in the caching broker with strict expiration limits.
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return RedisClient.Set(ctx, key, value, expiration).Err()
}

// Get queries and extracts the string value associated with a key from the caching broker.
func Get(ctx context.Context, key string) (string, error) {
	return RedisClient.Get(ctx, key).Result()
}

// Delete removes a key-value record from the caching broker instantly.
func Delete(ctx context.Context, key string) error {
	return RedisClient.Del(ctx, key).Err()
}

// Publish transmits real-time messages over Redis distributed pub/sub channels.
func Publish(ctx context.Context, channel string, message interface{}) error {
	return RedisClient.Publish(ctx, channel, message).Err()
}
