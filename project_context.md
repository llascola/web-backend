File: ./Dockerfile
```
# --- Stage 1: Builder ---
FROM golang:1.25-alpine AS builder

# Set working directory inside the container
WORKDIR /app

# Copy dependency files first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the binary named 'server'
RUN go build -o server main.go

# --- Stage 2: Runner ---
FROM alpine:latest

WORKDIR /root/

# Copy only the compiled binary from the builder stage
COPY --from=builder /app/server .

# Expose the port
EXPOSE 8080

# Run the app
CMD ["./server"]
```

File: ./internal/adapters/driven/storage/minio_adapter.go
```
package storage

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/llascola/web-backend/internal/app/outports"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOAdapter struct {
	client *minio.Client
	bucket string
	host   string
}

func NewMinIOAdapter(endpoint, accessKey, secretKey, bucket, policyTemplate string) *MinIOAdapter {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalf("MinIO Connection Failed: %v", err)
	}

	// --- Self-Healing: Ensure Bucket & Policy Exists ---
	ctx := context.Background()
	exists, _ := minioClient.BucketExists(ctx, bucket)
	if !exists {
		minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		log.Printf("Created bucket: %s", bucket)
	}

	// Force Public Policy
	policy := fmt.Sprintf(policyTemplate, bucket)
	minioClient.SetBucketPolicy(ctx, bucket, policy)

	return &MinIOAdapter{client: minioClient, bucket: bucket, host: endpoint}
}

func (a *MinIOAdapter) Save(ctx context.Context, file io.Reader, meta outports.FileMetadata) (string, error) {
	_, err := a.client.PutObject(ctx, a.bucket, meta.Name, file, meta.Size, minio.PutObjectOptions{
		ContentType: meta.ContentType,
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("http://%s/%s/%s", a.host, a.bucket, meta.Name), nil
}
```

File: ./internal/adapters/driving/gin/api/router.go
```
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
```

File: ./internal/adapters/driving/gin/controllers/images.go
```
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/llascola/web-backend/internal/app/inports"
	"github.com/llascola/web-backend/internal/app/outports"
)

type ImageController struct {
	imageService inports.ImageService
}

func NewImageController(imageService inports.ImageService) *ImageController {
	return &ImageController{imageService: imageService}
}

func (c *ImageController) UploadImage(ctx *gin.Context) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer file.Close()

	url, err := c.imageService.UploadImage(ctx, file, outports.FileMetadata{
		Name:        fileHeader.Filename,
		Size:        fileHeader.Size,
		ContentType: fileHeader.Header.Get("Content-Type"),
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"url": url})
}
```

File: ./internal/adapters/driving/gin/middleware/logger.go
```
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ZapLogger is a middleware that logs requests using zap
func ZapLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		if len(c.Errors) > 0 {
			for _, e := range c.Errors.Errors() {
				logger.Error(e)
			}
		} else {
			logger.Info(path,
				zap.Int("status", c.Writer.Status()),
				zap.String("method", c.Request.Method),
				zap.String("path", path),
				zap.String("query", query),
				zap.String("ip", c.ClientIP()),
				zap.String("user-agent", c.Request.UserAgent()),
				zap.Duration("latency", latency),
			)
		}
	}
}
```

File: ./internal/app/app.go
```
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
```

File: ./internal/app/domain/image.go
```
package domain

import (
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrImageTooLarge = errors.New("image size exceeds maximum limit")
	ErrInvalidFormat = errors.New("only jpeg and png formats are allowed")
)

type Image struct {
	ID           uuid.UUID
	OriginalName string
	StoredName   string
	ContentType  string
	Size         int64
	CreatedAt    time.Time
}

func NewImage(originalName string, contentType string, size int64) (*Image, error) {
	if size > 5*1024*1024 { // 5MB Limit
		return nil, ErrImageTooLarge
	}

	validTypes := map[string]bool{"image/jpeg": true, "image/png": true, "image/webp": true}
	if !validTypes[contentType] {
		return nil, ErrInvalidFormat
	}

	ext := filepath.Ext(originalName)
	if ext == "" {
		ext = ".jpg"
	}

	id := uuid.New()
	// StoredName sanitizes the filename to prevent path traversal
	storedName := id.String() + strings.ToLower(ext)

	return &Image{
		ID:           id,
		OriginalName: originalName,
		StoredName:   storedName,
		ContentType:  contentType,
		Size:         size,
		CreatedAt:    time.Now(),
	}, nil
}
```

File: ./internal/app/inports/image_service.go
```
package inports

import (
	"context"
	"io"

	"github.com/llascola/web-backend/internal/app/outports"
)

type ImageService interface {
	UploadImage(ctx context.Context, file io.Reader, meta outports.FileMetadata) (string, error)
}
```

File: ./internal/app/outports/storage.go
```
package outports

import (
	"context"
	"io"
)

type FileMetadata struct {
	Name        string
	Size        int64
	ContentType string
}

type FileStorageRepository interface {
	// Returns the public URL of the uploaded file
	Save(ctx context.Context, file io.Reader, meta FileMetadata) (string, error)
}
```

File: ./internal/app/services/image_service.go
```
package services

import (
	"context"
	"io"

	"github.com/llascola/web-backend/internal/app/domain"
	"github.com/llascola/web-backend/internal/app/outports"
)

type ImageServiceImpl struct {
	repo outports.FileStorageRepository
}

// NewImageService is the constructor
func NewImageService(repo outports.FileStorageRepository) *ImageServiceImpl {
	return &ImageServiceImpl{repo: repo}
}

func (s *ImageServiceImpl) UploadImage(ctx context.Context, file io.Reader, meta outports.FileMetadata) (string, error) {

	img, err := domain.NewImage(meta.Name, meta.ContentType, meta.Size)
	if err != nil {
		return "", err // Returns "image size exceeds..." or "only jpeg..."
	}

	repoMeta := outports.FileMetadata{
		Name:        img.StoredName,
		Size:        img.Size,
		ContentType: img.ContentType,
	}

	uploadURL, err := s.repo.Save(ctx, file, repoMeta)
	if err != nil {
		return "", err
	}

	return uploadURL, nil
}
```

File: ./internal/config/config.go
```
package config

import (
	"os"
)

const defaultPublicPolicy = `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {"AWS": ["*"]},
            "Action": ["s3:GetObject"],
            "Resource": ["arn:aws:s3:::%s/*"]
        }
    ]
}`

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	Policy    string
}

type Config struct {
	MinIO MinIOConfig
}

func Load() *Config {
	policy := os.Getenv("MINIO_POLICY")
	if policy == "" {
		policy = defaultPublicPolicy
	}

	return &Config{
		MinIO: MinIOConfig{
			Endpoint:  os.Getenv("MINIO_ENDPOINT"),
			AccessKey: os.Getenv("MINIO_ACCESS_KEY"),
			SecretKey: os.Getenv("MINIO_SECRET_KEY"),
			Bucket:    os.Getenv("MINIO_BUCKET"),
			Policy:    policy,
		},
	}
}
```

File: ./main.go
```
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
```

File: ./server.go
```
package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	srv *http.Server
}

func NewServer(router *gin.Engine) *Server {
	return &Server{
		srv: &http.Server{
			Addr:              ":8001",
			Handler:           router,
			ReadHeaderTimeout: 10 * time.Second,
		},
	}
}

func (s *Server) Run(ctx context.Context) error {
	g, errCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		err := s.srv.ListenAndServe()
		if err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-errCtx.Done()
		if err := s.Shutdown(context.Background()); err != nil {
			return err
		}
		return nil
	})

	return g.Wait()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
```

