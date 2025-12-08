package app

import (
	"context"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/llascola/web-backend/internal/adapters/driven/repository/ent"
	"github.com/llascola/web-backend/internal/adapters/driven/repository/postgres"
	"github.com/llascola/web-backend/internal/adapters/driven/storage"
	"github.com/llascola/web-backend/internal/app/inports"
	"github.com/llascola/web-backend/internal/app/services"
	"github.com/llascola/web-backend/internal/config"
)

type Service struct {
	ImageService inports.ImageService
	UserService  inports.UserService
	AuthService  inports.AuthService
}

type Application struct {
	Service *Service
}

func NewApplication(cfg *config.Config) *Application {
	fileStorage := storage.NewMinIOAdapter(
		cfg.MinIO.Endpoint,
		cfg.MinIO.RootUser,
		cfg.MinIO.RootPass,
		cfg.MinIO.Bucket,
		cfg.MinIO.Policy,
	)

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.DBName,
		cfg.Postgres.SSLMode,
	)

	client, err := ent.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}
	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	userRepo := postgres.NewUserRepository(client)

	imageService := services.NewImageService(fileStorage)
	userService := services.NewUserService(userRepo)
	authService := services.NewAuthService(userRepo, cfg.JWTKeys, cfg.ActiveKeyID)

	return &Application{
		Service: &Service{
			ImageService: imageService,
			UserService:  userService,
			AuthService:  authService,
		},
	}
}
