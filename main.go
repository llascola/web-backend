package main

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Initialize Gin
	r := gin.Default()

	// 2. Configure CORS (Crucial for separating Frontend/Backend)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://lucianoscola.com", "https://www.lucianoscola.com", "http://localhost:5173"}, // Added localhost for dev
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 3. Define Routes
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

	// 4. Run Server
	// We run on 0.0.0.0:8001 so Docker can map it correctly
	r.Run(":8001")
}
