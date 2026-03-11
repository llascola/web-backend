package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/llascola/web-backend/internal/adapters/driven/repository/memory"
	"github.com/llascola/web-backend/internal/adapters/driven/security"
	"github.com/llascola/web-backend/internal/adapters/driving/rest"
	"github.com/llascola/web-backend/internal/adapters/driving/rest/openapi"
	"github.com/llascola/web-backend/internal/app"
	"github.com/llascola/web-backend/internal/app/outports"
	"github.com/llascola/web-backend/internal/app/services"
	"github.com/llascola/web-backend/internal/config"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRouter creates a fully wired Gin engine using *only* In-Memory adapters.
// This allows us to test the entire HTTP lifecycle, JSON binding, OpenAPI validation,
// and Security Middlewares without touching Postgres or Redis.
func setupTestRouter() (*gin.Engine, *memory.InMemoryUserRepository, outports.TokenBlocklist) {
	// 1. Setup Config
	cfg := &config.Config{
		JWTKeys: map[string]config.JWTKey{
			"test-key": {
				Secret:    []byte("test-access-secret-32-bytes-long!"),
				Algorithm: "HS256",
			},
		},
		ActiveKeyID: "test-key",
	}

	// 2. Setup In-Memory Adapters
	userRepo := memory.NewUserRepository()
	blocklist := memory.NewInMemoryTokenBlocklist()
	tokenGen := security.NewJWTGenerator(cfg.JWTKeys, cfg.ActiveKeyID)

	// Optional: we can pass nil for storage since auth flow doesn't use it,
	// but setting up a dummy one prevents nil panics if routes leak.
	var storage outports.FileStorageRepository = nil

	// 3. Setup Services
	authService := services.NewAuthService(userRepo, tokenGen, blocklist)
	userService := services.NewUserService(userRepo)
	imageService := services.NewImageService(storage)

	// 4. Setup Application Container
	application := &app.Application{
		Service: &app.Service{
			AuthService:    authService,
			UserService:    userService,
			ImageService:   imageService,
			TokenBlocklist: blocklist,
		},
	}

	// 5. Build Router
	router := rest.NewRouter(application, cfg)

	return router, userRepo, blocklist
}

func TestE2EAuthFlow(t *testing.T) {
	router, repo, _ := setupTestRouter()

	email := "e2e@example.com"
	password := "SecurePass123!"
	var accessToken string

	t.Run("1. Register User", func(t *testing.T) {
		reqBody := openapi.AuthRequest{
			Email:    openapi_types.Email(email),
			Password: password,
		}
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		// Verify it actually hit the in-memory database
		user, err := repo.FindByEmail(context.Background(), email)
		require.NoError(t, err)
		assert.Equal(t, email, user.Email)
	})

	t.Run("1.5 Prevent Duplicate Registration (409 Conflict)", func(t *testing.T) {
		reqBody := openapi.AuthRequest{
			Email:    openapi_types.Email(email),
			Password: password,
		}
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert that the domain.ErrConflict correctly translates to HTTP 409 Conflict
		assert.Equal(t, http.StatusConflict, w.Code)

		var resp map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "user already exists", resp["error"])
	})

	t.Run("2. Login User", func(t *testing.T) {
		reqBody := openapi.AuthRequest{
			Email:    openapi_types.Email(email),
			Password: password,
		}
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp openapi.LoginResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.NotEmpty(t, resp.AccessToken)
		assert.NotEmpty(t, resp.RefreshToken)

		// Save access token for subsequent requests
		accessToken = resp.AccessToken
	})

	t.Run("3. Access Protected Profile Route", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/profile", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp openapi.UserProfile
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, openapi_types.Email(email), resp.Email)
		assert.NotEqual(t, openapi_types.UUID{}, resp.Id)
	})

	t.Run("4. Logout User", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Successfully logged out")
	})

	t.Run("5. Deny Access to Profile After Logout", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/profile", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// The SecurityMiddleware should catch the blocklisted token and reject it
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Token has been revoked")
	})
}
