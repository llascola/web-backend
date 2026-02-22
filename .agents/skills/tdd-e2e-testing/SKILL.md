---
name: tdd-e2e-testing
description: Guidelines for the middle of the Test Pyramid. Use this skill to write high-speed, in-memory E2E API Tests evaluating the Gin Router, OpenAPI unmarshaling, JSON payloads, and HTTP Middlewares without a database.
metadata:
  version: "1.0"
---

# TDD: E2E API Tests (The Middle)

The middle of the Test Pyramid proves that the routing, HTTP parsing, `oapi-codegen` generated wrappers, and Gin Middlewares (`SecurityMiddleware`) connect flawlessly to our Services.

## Core Strategy: In-Memory Fake Servers

To keep HTTP lifecycle tests blazing fast and free of external dockers/networking, we construct the REST API application using **Fake Adapters** implementing the Out-Ports inside the `memory` package.

### Step 1: Create Memory Adapters
Fake adapters reside in `internal/adapters/driven/repository/memory/`. Use Go `map` and `sync.RWMutex` to mimic caching and SQL tables.

```go
// Example Memory Adapter
type InMemoryUserRepository struct {
	users map[uuid.UUID]*domain.User
	mu    sync.RWMutex
}

var _ outports.UserRepository = (*InMemoryUserRepository)(nil)
```

### Step 2: Wire the Test Router
Instantiate the `app.Application` bypassing real infrastructure dependencies.

```go
// e2e_auth_flow_test.go
func setupTestRouter() (*gin.Engine, *memory.InMemoryUserRepository) {
	// 1. Fake adapters
	userRepo := memory.NewUserRepository()
	
	// 2. Real Services
	authService := services.NewAuthService(userRepo, ...)

	// 3. Fake Application Container
	application := &app.Application{
		Service: &app.Service{
			AuthService: authService,
		},
	}

	// 4. Real Gin Router
	router := rest.NewRouter(application, &mockConfig)
	return router, userRepo
}
```

### Step 3: Test Real HTTP Calls
Fire real JSON HTTP requests through the endpoints using `httptest` and assert the responses.

```go
func TestE2EFlow(t *testing.T) {
	router, repo := setupTestRouter()

	t.Run("Create Item", func(t *testing.T) {
		reqBody := openapi.CreateRequest{Name: "Test Item"}
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/api/items", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 1. Check HTTP Status
		assert.Equal(t, http.StatusCreated, w.Code)

		// 2. Check direct internal memory state via the exposed Fakes
		found, err := repo.FindByName(context.Background(), "Test Item")
		require.NoError(t, err)
		assert.NotNil(t, found)
	})
}
```

## Rules
- **Never mock `httptest`**: Test the full middleware -> handler cascade.
- **Check exact schemas**: Unmarshal the `w.Body.Bytes()` into `openapi.MyResponseStruct` to prove the exact JSON fields being served match the OpenAPI types.
