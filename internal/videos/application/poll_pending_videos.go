package application

import (
	"context"
	"log/slog"

	"go-starter/internal/videos/domain"
)

const pollChunkSize = 5

type PollPendingVideosResult struct {
	ProcessedIDs []string
	FailedIDs    []string
}

type PollPendingVideos struct {
	videoRepo domain.IVideoRepository
	queue     domain.IMessageQueue
}

func NewPollPendingVideos(
	videoRepo domain.IVideoRepository,
	queue domain.IMessageQueue,
) *PollPendingVideos {
	return &PollPendingVideos{
		videoRepo: videoRepo,
		queue:     queue,
	}
}

func (uc *PollPendingVideos) Execute(ctx context.Context) (*PollPendingVideosResult, error) {
	statusPending := domain.StatusPendingUpload
	videos, _, err := uc.videoRepo.List(ctx, domain.VideoListFilter{
		Status: &statusPending,
		Page:   1,
		Limit:  pollChunkSize,
		SortBy: "uploaded_at",
		Order:  "asc",
	})
	if err != nil {
		return nil, err
	}

	result := &PollPendingVideosResult{}

	for _, video := range videos {
		videoID := video.GetID()

		if err := video.TransitionTo(domain.StatusProcessing); err != nil {
			slog.Error("poller: failed to transition video", "video_id", videoID, "error", err)
			result.FailedIDs = append(result.FailedIDs, videoID)
			continue
		}

		if _, err := uc.videoRepo.Update(ctx, video); err != nil {
			slog.Error("poller: failed to update video status", "video_id", videoID, "error", err)
			result.FailedIDs = append(result.FailedIDs, videoID)
			continue
		}

		if err := uc.queue.Enqueue(ctx, domain.VideoProcessingJob{
			VideoID:            videoID,
			RequestedQualities: []string{"480p", "720p", "1080p", "4k"},
			IsAppend:           false,
		}); err != nil {
			slog.Error("poller: failed to enqueue video", "video_id", videoID, "error", err)
			result.FailedIDs = append(result.FailedIDs, videoID)
			continue
		}

		result.ProcessedIDs = append(result.ProcessedIDs, videoID)
	}

	return result, nil
}
