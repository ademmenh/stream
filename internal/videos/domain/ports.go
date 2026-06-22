package domain

import (
	"context"
	"time"
)

type IVideoRepository interface {
	Create(ctx context.Context, video *Video) (*Video, error)
	FindByID(ctx context.Context, id string) (*Video, error)
	List(ctx context.Context, filter VideoListFilter) ([]*Video, int, error)
	Update(ctx context.Context, video *Video) (*Video, error)
	Delete(ctx context.Context, id string) error
}

type VideoListFilter struct {
	Search string
	Type   *VideoType
	Status *VideoStatus
	Page   int
	Limit  int
	SortBy string
	Order  string
}

type IStorageAdapter interface {
	GeneratePresignedUploadUrl(ctx context.Context, key string, expiry time.Duration) (string, error)
	GeneratePresignedGetUrl(ctx context.Context, key string, expiry time.Duration) (string, error)
	ObjectExists(ctx context.Context, key string) (bool, error)
	DeletePrefixes(ctx context.Context, prefixes []string) error
}

type IMessageQueue interface {
	Enqueue(ctx context.Context, job VideoProcessingJob) error
}

type VideoProcessingJob struct {
	VideoID            string   `json:"video_id"`
	RequestedQualities []string `json:"requested_qualities"`
	IsAppend           bool     `json:"is_append"`
}
