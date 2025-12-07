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
