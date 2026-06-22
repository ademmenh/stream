package application

import (
	"context"

	"go-starter/internal/videos/domain"
)

type ReplaceVideoInput struct {
	VideoID string
}

type ReplaceVideoOutput struct {
	VideoUploadUrl string `json:"video_upload_url"`
}

type ReplaceVideo struct {
	videoRepo domain.IVideoRepository
	storage   domain.IStorageAdapter
	rawExpiry int
}

func NewReplaceVideo(
	videoRepo domain.IVideoRepository,
	storage domain.IStorageAdapter,
	rawExpiry int,
) *ReplaceVideo {
	return &ReplaceVideo{
		videoRepo: videoRepo,
		storage:   storage,
		rawExpiry: rawExpiry,
	}
}

func (uc *ReplaceVideo) Execute(ctx context.Context, input ReplaceVideoInput) (*ReplaceVideoOutput, error) {
	video, err := uc.videoRepo.FindByID(ctx, input.VideoID)
	if err != nil {
		return nil, err
	}
	if video == nil {
		return nil, &domain.VideoNotFoundError{ID: input.VideoID}
	}

	if err := video.TransitionTo(domain.StatusReplacing); err != nil {
		return nil, err
	}

	_, err = uc.videoRepo.Update(ctx, video)
	if err != nil {
		return nil, err
	}

	rawKey := "raws/" + video.GetID() + ".mp4"
	videoUploadUrl, err := uc.storage.GeneratePresignedUploadUrl(ctx, rawKey, durationFromMinutes(uc.rawExpiry))
	if err != nil {
		return nil, err
	}

	return &ReplaceVideoOutput{
		VideoUploadUrl: videoUploadUrl,
	}, nil
}
