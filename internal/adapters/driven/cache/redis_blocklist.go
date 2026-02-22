package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/llascola/web-backend/internal/app/outports"
	"github.com/redis/go-redis/v9"
)

// RedisTokenBlocklist implements the TokenBlocklist port using Redis.
// It leverages Redis TTL features to automatically evict tokens when they naturally expire.
type RedisTokenBlocklist struct {
	client *redis.Client
}

var _ outports.TokenBlocklist = (*RedisTokenBlocklist)(nil)

// NewRedisTokenBlocklist creates a new Redis-backed blocklist.
func NewRedisTokenBlocklist(client *redis.Client) *RedisTokenBlocklist {
	return &RedisTokenBlocklist{
		client: client,
	}
}

// Add revokes a token by its JTI by storing it in Redis until its natural expiration.
// If the token is already expired, it won't be added to the cache.
func (b *RedisTokenBlocklist) Add(ctx context.Context, jti string, expiration time.Time) error {
	ttl := time.Until(expiration)
	if ttl <= 0 {
		return nil // Token is already expired naturally
	}

	key := fmt.Sprintf("blocked_jti:%s", jti)

	// Value doesn't matter, we just check existence
	err := b.client.Set(ctx, key, "1", ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to block token in redis: %w", err)
	}

	return nil
}

// IsBlocked checks if a token's JTI exists in the Redis blocklist.
func (b *RedisTokenBlocklist) IsBlocked(ctx context.Context, jti string) (bool, error) {
	key := fmt.Sprintf("blocked_jti:%s", jti)

	val, err := b.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil // Key does not exist, token is not blocked
	}
	if err != nil {
		return false, fmt.Errorf("error checking redis blocklist: %w", err)
	}

	return val == "1", nil
}
