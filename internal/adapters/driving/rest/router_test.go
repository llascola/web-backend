package rest_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/llascola/web-backend/internal/adapters/driving/rest"
	"github.com/llascola/web-backend/internal/app"
	"github.com/llascola/web-backend/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	// Setup
	cfg := &config.Config{
		JWTKeys: map[string]config.JWTKey{},
	}
	application := &app.Application{
		Service: &app.Service{}, // Empty services are fine for HealthCheck
	}
	router := rest.NewRouter(application, cfg)

	// Request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}

func TestStatus(t *testing.T) {
	// Setup
	cfg := &config.Config{
		JWTKeys: map[string]config.JWTKey{},
	}
	application := &app.Application{
		Service: &app.Service{},
	}
	router := rest.NewRouter(application, cfg)

	// Request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/status", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "online")
}
