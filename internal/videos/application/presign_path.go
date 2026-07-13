package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-starter/internal/videos/domain"
)

type PresignPathInput struct {
	VideoID string
	Path    string
}

type PresignPathOutput struct {
	URL string `json:"url"`
}

type PresignPath struct {
	videoRepo domain.IVideoRepository
	storage   domain.IStorageAdapter
}

func NewPresignPath(
	videoRepo domain.IVideoRepository,
	storage domain.IStorageAdapter,
) *PresignPath {
	return &PresignPath{
		videoRepo: videoRepo,
		storage:   storage,
	}
}

func (uc *PresignPath) Execute(ctx context.Context, input PresignPathInput) (*PresignPathOutput, error) {
	video, err := uc.videoRepo.FindByID(ctx, input.VideoID)
	if err != nil {
		return nil, err
	}
	if video == nil {
		return nil, &domain.VideoNotFoundError{ID: input.VideoID}
	}

	if video.GetStatus() != domain.StatusReady {
		return nil, &domain.VideoNotReadyError{ID: input.VideoID}
	}

	relativePath := strings.TrimPrefix(input.Path, "/")
	if relativePath == "" {
		return nil, fmt.Errorf("path is required")
	}

	key := "videos/" + video.GetID() + "/" + relativePath

	signedURL, err := uc.storage.GeneratePresignedGetUrl(ctx, key, 1*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("generate signed url: %w", err)
	}

	return &PresignPathOutput{
		URL: signedURL,
	}, nil
}
