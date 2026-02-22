package postgres_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/llascola/web-backend/internal/adapters/driven/repository/ent"
	"github.com/llascola/web-backend/internal/adapters/driven/repository/postgres"
	"github.com/llascola/web-backend/internal/app/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	testcontainerpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupPostgresContainer(ctx context.Context) (*testcontainerpostgres.PostgresContainer, *ent.Client, error) {
	pgContainer, err := testcontainerpostgres.Run(ctx,
		"docker.io/postgres:15-alpine",
		testcontainerpostgres.WithDatabase("testdb"),
		testcontainerpostgres.WithUsername("testuser"),
		testcontainerpostgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start container: %w", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	client, err := ent.Open("postgres", connStr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed opening connection to postgres: %w", err)
	}

	if err := client.Schema.Create(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed creating schema resources: %w", err)
	}

	return pgContainer, client, nil
}

func TestPostgresUserRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	container, client, err := setupPostgresContainer(ctx)
	require.NoError(t, err)

	// Clean up the container after the tests run
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("failed to close ent client: %s", err)
		}
		if err := container.Terminate(ctx); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	repo := postgres.NewUserRepository(client)

	t.Run("Save and Find User", func(t *testing.T) {
		user, err := domain.NewUser("int-test@example.com", "Password123!", "member")
		require.NoError(t, err)

		err = repo.Save(ctx, user)
		require.NoError(t, err)

		// Find by ID
		foundUser, err := repo.FindByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.Email, foundUser.Email)
		assert.Equal(t, user.Role, foundUser.Role)

		// Find by Email
		foundEmailUser, err := repo.FindByEmail(ctx, user.Email)
		require.NoError(t, err)
		assert.Equal(t, user.ID, foundEmailUser.ID)
	})

	t.Run("Update User", func(t *testing.T) {
		user, err := domain.NewUser("update-test@example.com", "Password123!", "member")
		require.NoError(t, err)

		err = repo.Save(ctx, user)
		require.NoError(t, err)

		// Set a refresh token
		err = user.SetRefreshToken("test-refresh-token", time.Now().Add(24*time.Hour))
		require.NoError(t, err)

		// Update
		err = repo.Update(ctx, user)
		require.NoError(t, err)

		// Verify update
		updatedUser, err := repo.FindByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.RefreshTokenHash, updatedUser.RefreshTokenHash)
	})

	t.Run("Delete User", func(t *testing.T) {
		user, err := domain.NewUser("delete-test@example.com", "Password123!", "member")
		require.NoError(t, err)

		err = repo.Save(ctx, user)
		require.NoError(t, err)

		// Delete
		err = repo.Delete(ctx, user.ID)
		require.NoError(t, err)

		// Verify deleted
		_, err = repo.FindByID(ctx, user.ID)
		assert.ErrorContains(t, err, "user not found")
	})
}
