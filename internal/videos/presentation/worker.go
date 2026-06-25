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
	slog.Info("video queue consumer started")
	for {
		select {
		case <-ctx.Done():
			slog.Info("video queue consumer stopped")
			return
		default:
			job, err := w.queue.Dequeue(ctx)
			if err != nil {
				slog.Error("consumer dequeue error", "error", err)
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
				slog.Error("process video error", "video_id", job.VideoID, "error", err)
			}
		}
	}
}
