package application

import (
	"context"

	"go-starter/internal/videos/domain"
)

type ReplaceVideoInput struct {
	VideoID string
}

type ReplaceVideoOutput struct {
	RawPath string `json:"raw_path"`
}

type ReplaceVideo struct {
	videoRepo domain.IVideoRepository
}

func NewReplaceVideo(
	videoRepo domain.IVideoRepository,
) *ReplaceVideo {
	return &ReplaceVideo{
		videoRepo: videoRepo,
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
	if err := video.TransitionTo(domain.StatusPendingUpload); err != nil {
		return nil, err
	}

	_, err = uc.videoRepo.Update(ctx, video)
	if err != nil {
		return nil, err
	}

	return &ReplaceVideoOutput{
		RawPath: video.GetRawPath(),
	}, nil
}
