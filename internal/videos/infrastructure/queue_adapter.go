package infrastructure

import (
	"context"
	"sync"

	"go-starter/internal/videos/domain"
)

type InMemoryQueueAdapter struct {
	mu   sync.Mutex
	jobs []domain.VideoProcessingJob
}

func NewInMemoryQueueAdapter() *InMemoryQueueAdapter {
	return &InMemoryQueueAdapter{
		jobs: make([]domain.VideoProcessingJob, 0),
	}
}

func (q *InMemoryQueueAdapter) Enqueue(ctx context.Context, job domain.VideoProcessingJob) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.jobs = append(q.jobs, job)
	return nil
}

func (q *InMemoryQueueAdapter) Dequeue(ctx context.Context) (*domain.VideoProcessingJob, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.jobs) == 0 {
		return nil, nil
	}
	job := q.jobs[0]
	q.jobs = q.jobs[1:]
	return &job, nil
}
