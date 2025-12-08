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
	RootUser  string
	RootPass  string
	Bucket    string
	Policy    string
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type Config struct {
	MinIO       MinIOConfig
	Postgres    PostgresConfig
	JWTKeys     map[string]JWTKey
	ActiveKeyID string
}

type JWTKey struct {
	Secret    []byte
	Algorithm string
}

func Load() *Config {
	policy := os.Getenv("MINIO_POLICY")
	if policy == "" {
		policy = defaultPublicPolicy
	}

	keyID := os.Getenv("JWT_KEY_ID")

	return &Config{
		MinIO: MinIOConfig{
			Endpoint: os.Getenv("MINIO_ENDPOINT"),
			RootUser: os.Getenv("MINIO_ROOT_USER"),
			RootPass: os.Getenv("MINIO_ROOT_PASSWORD"),
			Bucket:   os.Getenv("MINIO_BUCKET"),
			Policy:   policy,
		},
		Postgres: PostgresConfig{
			Host:     os.Getenv("POSTGRES_HOST"),
			Port:     os.Getenv("POSTGRES_PORT"),
			User:     os.Getenv("POSTGRES_USER"),
			Password: os.Getenv("POSTGRES_PASSWORD"),
			DBName:   os.Getenv("POSTGRES_DB"),
			SSLMode:  os.Getenv("POSTGRES_SSLMODE"),
		},

		JWTKeys: map[string]JWTKey{
			keyID: {
				Secret:    []byte(os.Getenv("JWT_SECRET")),
				Algorithm: os.Getenv("JWT_ALGORITHM"),
			},
		},
		ActiveKeyID: keyID,
	}
}
