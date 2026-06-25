package infrastructure

import (
	"context"
	"fmt"
	"time"

	infra "go-starter/internal/shared/infrastructure"
)

type VideoStorageAdapter struct {
	s3 *infra.S3Adapter
}

func NewVideoStorageAdapter(s3 *infra.S3Adapter) *VideoStorageAdapter {
	return &VideoStorageAdapter{s3: s3}
}

func (a *VideoStorageAdapter) GeneratePresignedUploadUrl(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return a.s3.GetSignedUploadUrl(ctx, key, "application/octet-stream")
}

func (a *VideoStorageAdapter) GeneratePresignedGetUrl(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return a.s3.GetSignedUrl(ctx, key)
}

func (a *VideoStorageAdapter) ObjectExists(ctx context.Context, key string) (bool, error) {
	_, err := a.s3.GetSignedUrl(ctx, key)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (a *VideoStorageAdapter) DeletePrefixes(ctx context.Context, prefixes []string) error {
	for _, prefix := range prefixes {
		err := a.s3.DeleteFile(ctx, prefix)
		if err != nil {
			return fmt.Errorf("delete prefix %s: %w", prefix, err)
		}
	}
	return nil
}

func (a *VideoStorageAdapter) UploadFile(ctx context.Context, key string, body []byte, contentType string) (string, error) {
	return a.s3.UploadFile(ctx, key, body, contentType)
}
