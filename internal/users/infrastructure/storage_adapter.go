package infrastructure

import (
	"context"
	"time"

	infra "go-starter/internal/shared/infrastructure"
)

type UserStorageAdapter struct {
	s3 *infra.S3Adapter
}

func NewUserStorageAdapter(s3 *infra.S3Adapter) *UserStorageAdapter {
	return &UserStorageAdapter{s3: s3}
}

func (a *UserStorageAdapter) GeneratePresignedGetUrl(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return a.s3.GetSignedUrl(ctx, key)
}

func (a *UserStorageAdapter) GeneratePresignedUploadUrl(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return a.s3.GetSignedUploadUrl(ctx, key, "application/octet-stream")
}
