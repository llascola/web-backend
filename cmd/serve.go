package cmd

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	ginadapter "github.com/llascola/web-backend/internal/adapters/driving/gin"
	"github.com/llascola/web-backend/internal/adapters/driving/http"
	"github.com/llascola/web-backend/internal/app"
	"github.com/llascola/web-backend/internal/config"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the server",
	Run: func(cmd *cobra.Command, args []string) {
		// Load .env file
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found")
		}

		// Load Config
		cfg := config.Load()

		// Initialize Application
		application := app.NewApplication(cfg)

		// Initialize Router from ginadapter package
		r := ginadapter.NewGinRouter(application, cfg)

		// Run Server
		server := http.NewServer(r)
		server.Run(context.Background())
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
