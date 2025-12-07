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
