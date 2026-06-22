package application

import (
	"context"

	"go-starter/internal/videos/domain"
)

type DeleteVideo struct {
	videoRepo domain.IVideoRepository
	storage   domain.IStorageAdapter
}

func NewDeleteVideo(
	videoRepo domain.IVideoRepository,
	storage domain.IStorageAdapter,
) *DeleteVideo {
	return &DeleteVideo{
		videoRepo: videoRepo,
		storage:   storage,
	}
}

func (uc *DeleteVideo) Execute(ctx context.Context, id string) error {
	if err := uc.videoRepo.Delete(ctx, id); err != nil {
		return err
	}

	prefixes := []string{
		"videos/" + id + "/",
		"photos/" + id + "/",
		"raws/" + id + ".mp4",
	}

	_ = uc.storage.DeletePrefixes(ctx, prefixes)

	return nil
}
