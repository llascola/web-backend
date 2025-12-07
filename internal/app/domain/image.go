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
