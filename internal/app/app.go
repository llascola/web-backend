package app

import (
	"github.com/llascola/web-backend/internal/adapters/driven/storage"
	"github.com/llascola/web-backend/internal/app/inports"
	"github.com/llascola/web-backend/internal/app/services"
	"github.com/llascola/web-backend/internal/config"
)

type Service struct {
	ImageService inports.ImageService
}

type Application struct {
	Service *Service
}

func NewApplication(cfg *config.Config) *Application {
	fileStorage := storage.NewMinIOAdapter(
		cfg.MinIO.Endpoint,
		cfg.MinIO.AccessKey,
		cfg.MinIO.SecretKey,
		cfg.MinIO.Bucket,
		cfg.MinIO.Policy,
	)

	imageService := services.NewImageService(fileStorage)

	return &Application{
		Service: &Service{
			ImageService: imageService,
		},
	}
}
