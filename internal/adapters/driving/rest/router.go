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
	"github.com/llascola/web-backend/internal/app/domain"
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

	// Handler
	handler := handlers.NewHandler(app)
	wrapper := openapi.ServerInterfaceWrapper{
		Handler: handler,
	}

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

	r.GET("/health", wrapper.HealthCheck)

	// Swagger UI
	r.StaticFile("/openapi.yml", "./openapi/openapi.yml")
	r.Static("/docs", "./docs")

	// Public Routes
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", wrapper.Login)
		authGroup.POST("/register", wrapper.Register) // Anyone can register
	}

	// Protected Routes (Must be logged in)
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(cfg.JWTKeys))

	// 1. Member Routes (Any logged in user)
	api.GET("/profile", wrapper.GetProfile)

	// 2. Admin Routes (Only Admins)
	admin := api.Group("/admin")
	admin.Use(middleware.RequireRole(domain.RoleAdmin)) // <--- Blocks non-admins
	{
		admin.DELETE("/users/:id", wrapper.DeleteUser)
		admin.GET("/hola-mundo", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Hola Mundo"})
		})
		// admin.POST("/upload-system-config", imageController.UploadConfig) // Removed
		admin.POST("/upload-image", wrapper.UploadImage)
	}

	return r
}
