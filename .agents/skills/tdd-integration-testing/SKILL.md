---
name: tdd-integration-testing
description: Guidelines for the peak of the Test Pyramid. Use this skill when writing tests for actual infrastructure adapters (Postgres, Redis) by spinning up ephemeral remote Docker containers via testcontainers-go.
metadata:
  version: "1.0"
---

# TDD: Integration Testing (The Peak)

The peak of the Test Pyramid verifies that our code talks correctly to real external systems—processing raw SQL schemas, handling connection drivers, asserting DB-level constraints, and reacting to real TTL/Timeouts.

## Core Strategy: Ephemeral Docker Containers

We use `github.com/testcontainers/testcontainers-go` to dynamically spin up clean, temporary instances of Postgres and Redis before the test suite starts, verify the Adapters against them, and terminate the containers.

### Step 1: Skip Short Mode
Always wrap Integration tests to skip if `go test -short` is provided, as these take seconds to spin up.

```go
func TestPostgresAdapter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	// ...
}
```

### Step 2: Spin Up Testcontainers
Initialize the container using the exact versions matched in production.

#### Postgres Example
```go
import (
	"github.com/testcontainers/testcontainers-go"
	testcontainerpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupPostgresContainer(ctx context.Context) (*testcontainerpostgres.PostgresContainer, *ent.Client, error) {
	pgContainer, err := testcontainerpostgres.Run(ctx,
		"docker.io/postgres:15-alpine", // ALWAYS USE .Run(), NEVER .RunContainer() which is deprecated!
		testcontainerpostgres.WithDatabase("testdb"),
		testcontainerpostgres.WithUsername("testuser"),
		testcontainerpostgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	
	// Get standard connection string to pass to our Ent schema builders
	connStr, _ := pgContainer.ConnectionString(ctx, "sslmode=disable")
	client, _ := ent.Open("postgres", connStr)
	client.Schema.Create(ctx) // Auto-Migrate schema

	return pgContainer, client, nil
}
```

#### Redis Example
```go
import testcontainerredis "github.com/testcontainers/testcontainers-go/modules/redis"

func setupRedisContainer(ctx context.Context) (*testcontainerredis.RedisContainer, *redis.Client, error) {
	redisContainer, err := testcontainerredis.Run(ctx,
		"docker.io/redis:7-alpine", // ALWAYS USE .Run(), NEVER .RunContainer()
	)

	uri, _ := redisContainer.ConnectionString(ctx)
	redisOptions, _ := redis.ParseURL(uri)
	client := redis.NewClient(redisOptions)

	return redisContainer, client, nil
}
```

### Step 3: Run Setup and Teardown
```go
func TestRedisCache(t *testing.T) {
	ctx := context.Background()
	container, client, err := setupRedisContainer(ctx)
	require.NoError(t, err)

	// CLEANUP Container when done
	defer func() {
		client.Close()
		container.Terminate(ctx)
	}()

	adapter := cache.NewRedisAdapter(client)

	t.Run("Verifies Cache Set", func(t *testing.T) {
		err := adapter.Set(ctx, "key", "value")
		require.NoError(t, err)

		val, err := adapter.Get(ctx, "key")
		require.NoError(t, err)
		assert.Equal(t, "value", val)
	})
}
```

## Rules
- **No Global Containers**: Containers must be created inside standard Test functions leveraging `t.Run()` for sub-tests over the same container.
- **Terminate Always**: Always `defer container.Terminate(ctx)` immediately to prevent zombie docker instances clogging the host system.
- **Dependency Paths**: Integration tests belong alongside the implementation code in `internal/adapters/driven/repository/postgres/` or `internal/adapters/driven/cache/` (using package `package_name_test`).
