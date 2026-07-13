package presentation

import (
	"context"
	"log/slog"
	"time"

	"go-starter/internal/videos/application"
	"go-starter/internal/videos/domain"
)

type ConsumerQueueWorker struct {
	queue        domain.IMessageQueue
	processVideo *application.ProcessVideo
}

func NewConsumerQueueWorker(
	queue domain.IMessageQueue,
	processVideo *application.ProcessVideo,
) *ConsumerQueueWorker {
	return &ConsumerQueueWorker{
		queue:        queue,
		processVideo: processVideo,
	}
}

func (w *ConsumerQueueWorker) Start(ctx context.Context) {
	slog.Info("[ConsumerQueue] worker started")
	for {
		select {
		case <-ctx.Done():
			slog.Info("[ConsumerQueue] worker stopped")
			return
		default:
			job, err := w.queue.Dequeue(ctx)
			if err != nil {
				slog.Error("[ConsumerQueue] dequeue error", "error", err)
				time.Sleep(1 * time.Second)
				continue
			}
			if job == nil {
				time.Sleep(1 * time.Second)
				continue
			}

			if err := w.processVideo.Execute(ctx, application.ProcessVideoInput{
				VideoID:            job.VideoID,
				RequestedQualities: job.RequestedQualities,
				IsAppend:           job.IsAppend,
			}); err != nil {
				slog.Error("[ConsumerQueue] process video error", "video_id", job.VideoID, "error", err)
			}
		}
	}
}

type ProducerWorker struct {
	pollPendingVideos *application.PollPendingVideos
	isRunning         bool
}

func NewProducerWorker(
	pollPendingVideos *application.PollPendingVideos,
) *ProducerWorker {
	return &ProducerWorker{
		pollPendingVideos: pollPendingVideos,
	}
}

func (w *ProducerWorker) Start(ctx context.Context) {
	slog.Info("[Producer] worker started")
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("[Producer] worker stopped")
			return
		case <-ticker.C:
			if w.isRunning {
				slog.Info("[Producer] previous cycle still running, skipping tick")
				continue
			}
			w.run(ctx)
		}
	}
}

func (w *ProducerWorker) run(ctx context.Context) {
	w.isRunning = true
	defer func() { w.isRunning = false }()

	result, err := w.pollPendingVideos.Execute(ctx)
	if err != nil {
		slog.Error("[Producer] error during poll cycle", "error", err)
		return
	}

	if len(result.ProcessedIDs) > 0 {
		slog.Info("[Producer] enqueued videos", "count", len(result.ProcessedIDs), "ids", result.ProcessedIDs)
	}
	if len(result.FailedIDs) > 0 {
		slog.Error("[Producer] failed to enqueue videos", "count", len(result.FailedIDs), "ids", result.FailedIDs)
	}
}
