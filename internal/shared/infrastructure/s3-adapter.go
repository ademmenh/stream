package infrastructure

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	config "go-starter/internal/config/domain"
)

type S3Adapter struct {
	client           *minio.Client
	clientExternal   *minio.Client
	bucket           string
	presignedExpiry  time.Duration
	publicEndpoint   string
	publicPathPrefix string
}

func NewS3Adapter(cfg config.IConfig) (*S3Adapter, error) {
	endpoint := fmt.Sprintf("%s:%s", cfg.S3Host(), cfg.S3Port())
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3AccessKey(), cfg.S3SecretKey(), ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}

	rawPublicEndpoint := cfg.S3PublicEndpoint()
	parsedURL, err := url.Parse(rawPublicEndpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid S3_PUBLIC_ENDPOINT: %w", err)
	}
	publicEndpoint := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	publicPathPrefix := strings.TrimRight(parsedURL.Path, "/")

	clientExternal, err := minio.New(parsedURL.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3AccessKey(), cfg.S3SecretKey(), ""),
		Secure: parsedURL.Scheme == "https",
	})
	if err != nil {
		return nil, err
	}

	return &S3Adapter{
		client:           client,
		clientExternal:   clientExternal,
		bucket:           cfg.S3Bucket(),
		presignedExpiry: 7 * 24 * time.Hour,
		publicEndpoint:   publicEndpoint,
		publicPathPrefix: publicPathPrefix,
	}, nil
}

func (a *S3Adapter) Init(ctx context.Context) error {
	exists, _ := a.client.BucketExists(ctx, a.bucket)
	if !exists {
		if err := a.client.MakeBucket(ctx, a.bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	if err := a.setBucketPolicy(ctx); err != nil {
		return fmt.Errorf("failed to set bucket policy: %w", err)
	}

	slog.Info("S3 adapter initialized", "bucket", a.bucket)
	return nil
}

func (a *S3Adapter) setBucketPolicy(ctx context.Context) error {
	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": "*",
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}
		]
	}`, a.bucket)

	return a.client.SetBucketPolicy(ctx, a.bucket, policy)
}

func (a *S3Adapter) UploadFile(ctx context.Context, key string, body []byte, contentType string) (string, error) {
	opts := minio.PutObjectOptions{ContentType: contentType}
	_, err := a.client.PutObject(ctx, a.bucket, key, strings.NewReader(string(body)), int64(len(body)), opts)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}
	return key, nil
}

func (a *S3Adapter) injectPrefix(rawURL string) string {
	return strings.Replace(rawURL, a.publicEndpoint, a.publicEndpoint+a.publicPathPrefix, 1)
}

func (a *S3Adapter) GetSignedUrl(ctx context.Context, key string) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := a.clientExternal.PresignedGetObject(ctx, a.bucket, key, a.presignedExpiry, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}
	return a.injectPrefix(presignedURL.String()), nil
}

func (a *S3Adapter) GetSignedUploadUrl(ctx context.Context, key string, contentType string) (string, error) {
	presignedURL, err := a.clientExternal.PresignedPutObject(ctx, a.bucket, key, a.presignedExpiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed upload URL: %w", err)
	}
	return a.injectPrefix(presignedURL.String()), nil
}

func (a *S3Adapter) GetPublicUrl(key string) string {
	return fmt.Sprintf("%s%s/%s/%s", a.publicEndpoint, a.publicPathPrefix, a.bucket, key)
}

func (a *S3Adapter) DeleteFile(ctx context.Context, key string) error {
	err := a.client.RemoveObject(ctx, a.bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}
