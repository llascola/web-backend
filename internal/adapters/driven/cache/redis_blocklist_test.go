package cache_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/llascola/web-backend/internal/adapters/driven/cache"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	testcontainerredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func setupRedisContainer(ctx context.Context) (*testcontainerredis.RedisContainer, *redis.Client, error) {
	redisContainer, err := testcontainerredis.Run(ctx,
		"docker.io/redis:7-alpine",
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start redis container: %w", err)
	}

	uri, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get redis connection string: %w", err)
	}

	redisOptions, err := redis.ParseURL(uri)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	client := redis.NewClient(redisOptions)
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, nil, fmt.Errorf("failed pinging redis: %w", err)
	}

	return redisContainer, client, nil
}

func TestRedisTokenBlocklist(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	container, client, err := setupRedisContainer(ctx)
	require.NoError(t, err)

	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("failed to close redis client: %s", err)
		}
		if err := container.Terminate(ctx); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	blocklist := cache.NewRedisTokenBlocklist(client)

	t.Run("Add and Verify Blocked Token (TTL Check)", func(t *testing.T) {
		jti := "test-jwt-id-12345"
		// Set execution for 1 second in the future
		expirationTime := time.Now().Add(1 * time.Second)

		// 1. Add token
		err := blocklist.Add(ctx, jti, expirationTime)
		require.NoError(t, err)

		// 2. Token should be blocked immediately after adding
		isBlocked, err := blocklist.IsBlocked(ctx, jti)
		require.NoError(t, err)
		assert.True(t, isBlocked)

		// 3. Wait 1.1 seconds for TTL to expire in Redis
		time.Sleep(1100 * time.Millisecond)

		// 4. Token should no longer be blocked because Redis auto-evicted it
		isBlockedAfterTTL, err := blocklist.IsBlocked(ctx, jti)
		require.NoError(t, err)
		assert.False(t, isBlockedAfterTTL)
	})

	t.Run("Returns False for Unknown Token", func(t *testing.T) {
		isBlocked, err := blocklist.IsBlocked(ctx, "non-existent-jti")
		require.NoError(t, err)
		assert.False(t, isBlocked)
	})
}
