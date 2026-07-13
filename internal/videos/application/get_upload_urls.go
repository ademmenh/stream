package application

import (
	"context"
	"time"

	"go-starter/internal/videos/domain"
)

type GetUploadUrlsInput struct {
	VideoID string
}

type GetUploadUrlsOutput struct {
	RawUploadUrl   string `json:"raw_upload_url"`
	PhotoUploadUrl string `json:"photo_upload_url"`
}

type GetUploadUrls struct {
	videoRepo  domain.IVideoRepository
	storage    domain.IStorageAdapter
	rawExpiry  int
	photoExpiry int
}

func NewGetUploadUrls(
	videoRepo domain.IVideoRepository,
	storage domain.IStorageAdapter,
	rawExpiry int,
	photoExpiry int,
) *GetUploadUrls {
	return &GetUploadUrls{
		videoRepo:  videoRepo,
		storage:    storage,
		rawExpiry:  rawExpiry,
		photoExpiry: photoExpiry,
	}
}

func (uc *GetUploadUrls) Execute(ctx context.Context, input GetUploadUrlsInput) (*GetUploadUrlsOutput, error) {
	video, err := uc.videoRepo.FindByID(ctx, input.VideoID)
	if err != nil {
		return nil, err
	}
	if video == nil {
		return nil, &domain.VideoNotFoundError{ID: input.VideoID}
	}

	rawUploadUrl, err := uc.storage.GeneratePresignedUploadUrl(ctx, video.GetRawPath(), durationFromMinutes(uc.rawExpiry))
	if err != nil {
		return nil, err
	}

	photoUploadUrl, err := uc.storage.GeneratePresignedUploadUrl(ctx, video.GetPhotoPath(), durationFromMinutes(uc.photoExpiry))
	if err != nil {
		return nil, err
	}

	return &GetUploadUrlsOutput{
		RawUploadUrl:   rawUploadUrl,
		PhotoUploadUrl: photoUploadUrl,
	}, nil
}

func durationFromMinutes(minutes int) time.Duration {
	return time.Duration(minutes) * time.Minute
}
