package inports

import (
	"context"
	"io"

	"github.com/llascola/web-backend/internal/app/outports"
)

type ImageService interface {
	UploadImage(ctx context.Context, file io.Reader, meta outports.FileMetadata) (string, error)
}
