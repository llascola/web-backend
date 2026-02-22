package services_test

import (
	"context"
	"strings"
	"testing"

	"github.com/llascola/web-backend/internal/app/domain"
	"github.com/llascola/web-backend/internal/app/outports"
	"github.com/llascola/web-backend/internal/app/outports/mocks"
	"github.com/llascola/web-backend/internal/app/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestImageService_UploadImage(t *testing.T) {
	ctx := context.Background()

	t.Run("Successfully Uploads Valid Image", func(t *testing.T) {
		mockStorage := mocks.NewMockFileStorageRepository(t)
		service := services.NewImageService(mockStorage)

		filePayload := strings.NewReader("fake image bytes")
		meta := outports.FileMetadata{
			Name:        "profile-pic.jpeg",
			Size:        1024 * 500, // 500KB
			ContentType: "image/jpeg",
		}

		mockStorage.On("Save", ctx, filePayload, mock.AnythingOfType("outports.FileMetadata")).
			Return("https://storage.example.com/images/uuid-profile.jpeg", nil).
			Run(func(args mock.Arguments) {
				savedMeta := args.Get(2).(outports.FileMetadata)
				assert.Equal(t, meta.Size, savedMeta.Size)
				assert.Equal(t, meta.ContentType, savedMeta.ContentType)
				assert.NotEqual(t, meta.Name, savedMeta.Name, "File name should be obfuscated/uniqueified by the domain object")
				assert.True(t, strings.HasSuffix(savedMeta.Name, ".jpeg"))
			})

		url, err := service.UploadImage(ctx, filePayload, meta)

		assert.NoError(t, err)
		assert.Equal(t, "https://storage.example.com/images/uuid-profile.jpeg", url)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Fails Domain Validation (File Too Large)", func(t *testing.T) {
		mockStorage := mocks.NewMockFileStorageRepository(t)
		service := services.NewImageService(mockStorage)

		filePayload := strings.NewReader("huge file bytes")
		meta := outports.FileMetadata{
			Name:        "huge.jpeg",
			Size:        5*1024*1024 + 10,
			ContentType: "image/jpeg",
		}

		// Save should never be called because domain validation throws first
		url, err := service.UploadImage(ctx, filePayload, meta)

		var target *domain.ErrValidation
		require.ErrorAs(t, err, &target)
		assert.Empty(t, url)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Fails Domain Validation (Invalid Content Type)", func(t *testing.T) {
		mockStorage := mocks.NewMockFileStorageRepository(t)
		service := services.NewImageService(mockStorage)

		filePayload := strings.NewReader("malicious exe bytes")
		meta := outports.FileMetadata{
			Name:        "malware.exe",
			Size:        1024,
			ContentType: "application/x-msdownload",
		}

		url, err := service.UploadImage(ctx, filePayload, meta)

		var target *domain.ErrValidation
		require.ErrorAs(t, err, &target)
		assert.Empty(t, url)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Fails if Storage Repository Returns Error", func(t *testing.T) {
		mockStorage := mocks.NewMockFileStorageRepository(t)
		service := services.NewImageService(mockStorage)

		filePayload := strings.NewReader("fake image bytes")
		meta := outports.FileMetadata{
			Name:        "valid.png",
			Size:        2048,
			ContentType: "image/png",
		}

		mockStorage.On("Save", ctx, filePayload, mock.AnythingOfType("outports.FileMetadata")).
			Return("", &domain.ErrInternal{Message: "s3 bucket connection timeout"})

		url, err := service.UploadImage(ctx, filePayload, meta)

		var target *domain.ErrInternal
		require.ErrorAs(t, err, &target)
		assert.Empty(t, url)
		mockStorage.AssertExpectations(t)
	})
}
