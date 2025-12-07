package main

import (
	"context"

	"log"

	"github.com/joho/godotenv"
	"github.com/llascola/web-backend/internal/adapters/driving/gin/api"
	"github.com/llascola/web-backend/internal/app"
	"github.com/llascola/web-backend/internal/config"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load Config
	cfg := config.Load()

	// Initialize Application
	application := app.NewApplication(cfg)

	// Initialize Router from api package
	r := api.NewRouter(application)

	// Run Server
	// We run on 0.0.0.0:8001 so Docker can map it correctly
	server := NewServer(r)
	server.Run(context.Background())
}
