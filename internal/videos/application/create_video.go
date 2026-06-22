package application

import (
	"context"
	"time"

	shareddomain "go-starter/internal/shared/domain"
	"go-starter/internal/videos/domain"
)

type CreateVideoInput struct {
	Title       string
	Description string
	Type        string
}

type CreateVideoOutput struct {
	ID              string `json:"id"`
	VideoUploadUrl  string `json:"video_upload_url"`
	PhotoUploadUrl  string `json:"photo_upload_url"`
}

type IDGenerator interface {
	Generate() string
}

type CreateVideo struct {
	videoRepo  domain.IVideoRepository
	storage    domain.IStorageAdapter
	idGen      IDGenerator
	rawExpiry  int
	photoExpiry int
}

func NewCreateVideo(
	videoRepo domain.IVideoRepository,
	storage domain.IStorageAdapter,
	idGen IDGenerator,
	rawExpiry int,
	photoExpiry int,
) *CreateVideo {
	return &CreateVideo{
		videoRepo:   videoRepo,
		storage:     storage,
		idGen:       idGen,
		rawExpiry:   rawExpiry,
		photoExpiry: photoExpiry,
	}
}

func (uc *CreateVideo) Execute(ctx context.Context, input CreateVideoInput) (*CreateVideoOutput, error) {
	videoType := domain.VideoType(input.Type)
	if !videoType.IsValid() {
		return nil, &domain.VideoInvalidTypeError{Type: input.Type}
	}

	id := shareddomain.IdFromStr(uc.idGen.Generate())
	video := domain.NewVideo(id, input.Title, input.Description, videoType)

	_, err := uc.videoRepo.Create(ctx, video)
	if err != nil {
		return nil, err
	}

	rawKey := "raws/" + video.GetID() + ".mp4"
	videoUploadUrl, err := uc.storage.GeneratePresignedUploadUrl(ctx, rawKey, durationFromMinutes(uc.rawExpiry))
	if err != nil {
		return nil, err
	}

	photoKey := "photos/" + video.GetID() + "/thumbnail.jpg"
	photoUploadUrl, err := uc.storage.GeneratePresignedUploadUrl(ctx, photoKey, durationFromMinutes(uc.photoExpiry))
	if err != nil {
		return nil, err
	}

	return &CreateVideoOutput{
		ID:             video.GetID(),
		VideoUploadUrl: videoUploadUrl,
		PhotoUploadUrl: photoUploadUrl,
	}, nil
}

func durationFromMinutes(minutes int) time.Duration {
	return time.Duration(minutes) * time.Minute
}
