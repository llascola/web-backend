package cmd

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"github.com/llascola/web-backend/internal/app"
	"github.com/llascola/web-backend/internal/config"
	"github.com/spf13/cobra"
)

var (
	adminEmail    string
	adminPassword string
)

var createAdminCmd = &cobra.Command{
	Use:   "create-admin",
	Short: "Create an admin user",
	Run: func(cmd *cobra.Command, args []string) {
		// Load .env file
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found")
		}

		// Load Config
		cfg := config.Load()

		// Initialize Application
		application := app.NewApplication(cfg)

		// Create Admin User
		err := application.Service.AuthService.RegisterAdmin(context.Background(), adminEmail, adminPassword)
		if err != nil {
			log.Fatalf("Failed to create admin user: %v", err)
		}

		log.Printf("Admin user created successfully: %s", adminEmail)
	},
}

func init() {
	createAdminCmd.Flags().StringVarP(&adminEmail, "email", "e", "", "Admin email")
	createAdminCmd.Flags().StringVarP(&adminPassword, "password", "p", "", "Admin password")
	createAdminCmd.MarkFlagRequired("email")
	createAdminCmd.MarkFlagRequired("password")

	rootCmd.AddCommand(createAdminCmd)
}
