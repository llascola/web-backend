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
