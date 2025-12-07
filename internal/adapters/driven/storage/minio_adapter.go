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
