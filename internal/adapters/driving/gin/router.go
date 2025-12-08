package ginadapter

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/llascola/web-backend/internal/adapters/driving/gin/controllers"
	"github.com/llascola/web-backend/internal/adapters/driving/gin/middleware"
	"github.com/llascola/web-backend/internal/app"
	"github.com/llascola/web-backend/internal/app/domain"
	"github.com/llascola/web-backend/internal/config"
	"go.uber.org/zap"
)

// NewRouter initializes the Gin engine with middleware and routes
func NewGinRouter(app *app.Application, cfg *config.Config) *gin.Engine {
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

	// Controllers
	imageController := controllers.NewImageController(app.Service.ImageService)
	userController := controllers.NewUserController(app.Service.UserService)
	authController := controllers.NewAuthController(app.Service.AuthService)

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

	// Public Routes
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", authController.Login)
		authGroup.POST("/register", authController.Register) // Anyone can register
	}

	// Protected Routes (Must be logged in)
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(cfg.JWTKeys))

	// 1. Member Routes (Any logged in user)
	api.GET("/profile", userController.GetProfile)

	// 2. Admin Routes (Only Admins)
	admin := api.Group("/admin")
	admin.Use(middleware.RequireRole(domain.RoleAdmin)) // <--- Blocks non-admins
	{
		admin.DELETE("/users/:id", userController.DeleteUser)
		admin.GET("/hola-mundo", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Hola Mundo"})
		})
		admin.POST("/upload-system-config", imageController.UploadConfig)
		admin.POST("/upload-image", imageController.UploadImage)
	}

	return r
}
