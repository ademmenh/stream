package application

import (
	"context"
	"time"

	"go-starter/internal/videos/domain"
)

type VideoStreamOutput struct {
	ID                string `json:"id"`
	Title             string `json:"title"`
	Description       string `json:"description"`
	Type              string `json:"type"`
	Qualities         []string `json:"qualities"`
	MasterPlaylistUrl string `json:"master_playlist_url"`
	UploadedAt        string `json:"uploaded_at"`
}

type GetVideoStream struct {
	videoRepo domain.IVideoRepository
	storage   domain.IStorageAdapter
}

func NewGetVideoStream(
	videoRepo domain.IVideoRepository,
	storage domain.IStorageAdapter,
) *GetVideoStream {
	return &GetVideoStream{
		videoRepo: videoRepo,
		storage:   storage,
	}
}

func (uc *GetVideoStream) Execute(ctx context.Context, id string) (*VideoStreamOutput, error) {
	video, err := uc.videoRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if video == nil {
		return nil, &domain.VideoNotFoundError{ID: id}
	}

	if video.GetStatus() != domain.StatusReady {
		return nil, &domain.VideoNotReadyError{ID: id}
	}

	masterKey := "videos/" + video.GetID() + "/master.m3u8"
	masterPlaylistUrl, err := uc.storage.GeneratePresignedGetUrl(ctx, masterKey, 1*time.Hour)
	if err != nil {
		return nil, err
	}

	return &VideoStreamOutput{
		ID:                video.GetID(),
		Title:             video.GetTitle(),
		Description:       video.GetDescription(),
		Type:              video.GetType().String(),
		Qualities:         qualitiesToStrings(video.GetQualities()),
		MasterPlaylistUrl: masterPlaylistUrl,
		UploadedAt:        video.GetUploadedAt().Format(time.RFC3339),
	}, nil
}
