package rest

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/llascola/web-backend/internal/adapters/driving/rest/handlers"
	"github.com/llascola/web-backend/internal/adapters/driving/rest/middleware"
	"github.com/llascola/web-backend/internal/adapters/driving/rest/openapi"
	"github.com/llascola/web-backend/internal/app"
	"github.com/llascola/web-backend/internal/config"
	"go.uber.org/zap"
)

// NewRouter initializes the Gin engine with middleware and routes
func NewRouter(app *app.Application, cfg *config.Config) *gin.Engine {
	// Initialize Zap Logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.ZapLogger(logger))

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://lucianoscola.com", "https://www.lucianoscola.com", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Static routes (not part of OpenAPI)
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "API Root - Go to /status"})
	})

	r.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "online",
			"system": "Go + Docker Microservice",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Swagger UI
	r.StaticFile("/openapi.yml", "./openapi/openapi.yml")
	r.Static("/docs", "./docs")

	// Register all OpenAPI routes via generated RegisterHandlers.
	// Security is handled per-endpoint via BearerAuthScopes set by the
	// generated wrapper — the SecurityMiddleware reads those scopes and
	// enforces JWT auth + RBAC automatically.
	handler := handlers.NewHandler(app)
	openapi.RegisterHandlersWithOptions(r, handler, openapi.GinServerOptions{
		Middlewares: []openapi.MiddlewareFunc{
			middleware.SecurityMiddleware(cfg.JWTKeys, app.Service.TokenBlocklist),
		},
		ErrorHandler: func(c *gin.Context, err error, statusCode int) {
			c.JSON(statusCode, gin.H{"error": err.Error()})
		},
	})

	return r
}
