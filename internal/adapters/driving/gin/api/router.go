package api

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/llascola/web-backend/internal/adapters/driving/gin/controllers"
	"github.com/llascola/web-backend/internal/adapters/driving/gin/middleware"
	"github.com/llascola/web-backend/internal/app"
	"go.uber.org/zap"
)

// NewRouter initializes the Gin engine with middleware and routes
func NewRouter(app *app.Application) *gin.Engine {
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

	r.POST("/upload", imageController.UploadImage)

	return r
}
