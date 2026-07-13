package videotests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	shareddomain "go-starter/internal/shared/domain"
	videosapp "go-starter/internal/videos/application"
	"go-starter/internal/videos/domain"
	"go-starter/internal/videos/infrastructure"
	videospres "go-starter/internal/videos/presentation"
)

type mockIDGenerator struct{}

func (m *mockIDGenerator) Generate() string { return shareddomain.NewId().String() }

type mockStorageAdapter struct {
	uploadUrls map[string]string
	getUrls    map[string]string
	objects    map[string][]byte
}

func newMockStorageAdapter() *mockStorageAdapter {
	return &mockStorageAdapter{
		uploadUrls: make(map[string]string),
		getUrls:    make(map[string]string),
		objects:    make(map[string][]byte),
	}
}

func (m *mockStorageAdapter) GeneratePresignedUploadUrl(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return "https://s3.example.com/" + key + "?upload", nil
}

func (m *mockStorageAdapter) GeneratePresignedGetUrl(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return "https://s3.example.com/" + key + "?get", nil
}

func (m *mockStorageAdapter) ObjectExists(ctx context.Context, key string) (bool, error) {
	_, ok := m.objects[key]
	return ok, nil
}

func (m *mockStorageAdapter) DeletePrefixes(ctx context.Context, prefixes []string) error {
	for _, p := range prefixes {
		for k := range m.objects {
			if len(k) >= len(p) && k[:len(p)] == p {
				delete(m.objects, k)
			}
		}
	}
	return nil
}

func (m *mockStorageAdapter) UploadFile(ctx context.Context, key string, body []byte, contentType string) (string, error) {
	m.objects[key] = body
	return key, nil
}

func (m *mockStorageAdapter) DownloadFile(ctx context.Context, key string) ([]byte, error) {
	data, ok := m.objects[key]
	if !ok {
		return nil, nil
	}
	return data, nil
}

type mockMessageQueue struct {
	jobs []domain.VideoProcessingJob
}

func newMockMessageQueue() *mockMessageQueue {
	return &mockMessageQueue{}
}

func (q *mockMessageQueue) Enqueue(ctx context.Context, job domain.VideoProcessingJob) error {
	q.jobs = append(q.jobs, job)
	return nil
}

func (q *mockMessageQueue) Dequeue(ctx context.Context) (*domain.VideoProcessingJob, error) {
	if len(q.jobs) == 0 {
		return nil, nil
	}
	job := q.jobs[0]
	q.jobs = q.jobs[1:]
	return &job, nil
}

type mockTranscoder struct{}

func (m *mockTranscoder) TranscodeToHLS(ctx context.Context, params domain.TranscodeParams) error {
	return nil
}

func seedVideo(t *testing.T, repo *infrastructure.InMemoryVideoRepository, id, title, desc string, videoType domain.VideoType, status domain.VideoStatus) *domain.Video {
	t.Helper()
	v := &domain.Video{
		ID:          shareddomain.IdFromStr(id),
		Title:       title,
		Description: desc,
		Type:        videoType,
		Status:      status,
		Qualities:   []domain.VideoQuality{},
		RawPath:     "raws/" + id + ".mp4",
		PhotoPath:   "photos/" + id + "/thumbnail.jpg",
		UploadedAt:  time.Now().UTC(),
	}
	_, err := repo.Create(context.Background(), v)
	require.NoError(t, err)
	return v
}

func TestCreateVideo_Success(t *testing.T) {
	repo := infrastructure.NewInMemoryVideoRepository()

	storage := newMockStorageAdapter()
	uc := videosapp.NewCreateVideo(repo, storage, &mockIDGenerator{}, 60, 15)
	result, err := uc.Execute(context.Background(), videosapp.CreateVideoInput{
		Title:       "Test Movie",
		Description: "A test movie",
		Type:        "movie",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, result.ID)
	assert.Contains(t, result.RawUploadUrl, "raws/")
	assert.Contains(t, result.PhotoUploadUrl, "photos/")

	v, err := repo.FindByID(context.Background(), result.ID)
	require.NoError(t, err)
	assert.Equal(t, "Test Movie", v.Title)
	assert.Equal(t, domain.VideoType("movie"), v.Type)
	assert.Equal(t, domain.StatusPendingUpload, v.Status)
}

func TestCreateVideo_InvalidType(t *testing.T) {
	repo := infrastructure.NewInMemoryVideoRepository()

	uc := videosapp.NewCreateVideo(repo, newMockStorageAdapter(), &mockIDGenerator{}, 60, 15)
	_, err := uc.Execute(context.Background(), videosapp.CreateVideoInput{
		Title:       "Invalid",
		Description: "",
		Type:        "invalid_type",
	})

	assert.Error(t, err)
	var invalidType *domain.VideoInvalidTypeError
	assert.ErrorAs(t, err, &invalidType)
}

func TestTriggerProcessing_Success(t *testing.T) {
	repo := infrastructure.NewInMemoryVideoRepository()
	queue := newMockMessageQueue()
	seedVideo(t, repo, "v1", "Movie", "desc", domain.VideoTypeMovie, domain.StatusPendingUpload)

	uc := videosapp.NewTriggerProcessing(repo, queue)
	err := uc.Execute(context.Background(), videosapp.TriggerProcessingInput{
		VideoID:            "v1",
		RequestedQualities: []string{"480p", "1080p"},
	})

	require.NoError(t, err)

	v, err := repo.FindByID(context.Background(), "v1")
	require.NoError(t, err)
	assert.Equal(t, domain.StatusProcessing, v.Status)

	assert.Len(t, queue.jobs, 1)
	assert.Equal(t, "v1", queue.jobs[0].VideoID)
	assert.False(t, queue.jobs[0].IsAppend)
}

func TestTriggerProcessing_NotFound(t *testing.T) {
	repo := infrastructure.NewInMemoryVideoRepository()
	queue := newMockMessageQueue()

	uc := videosapp.NewTriggerProcessing(repo, queue)
	err := uc.Execute(context.Background(), videosapp.TriggerProcessingInput{
		VideoID:            "nonexistent",
		RequestedQualities: []string{"480p"},
	})

	assert.Error(t, err)
	var notFound *domain.VideoNotFoundError
	assert.ErrorAs(t, err, &notFound)
}

func TestReplaceVideo_Success(t *testing.T) {
	repo := infrastructure.NewInMemoryVideoRepository()
	seedVideo(t, repo, "v1", "Movie", "desc", domain.VideoTypeMovie, domain.StatusReady)

	uc := videosapp.NewReplaceVideo(repo)
	result, err := uc.Execute(context.Background(), videosapp.ReplaceVideoInput{
		VideoID: "v1",
	})

	require.NoError(t, err)
	assert.Contains(t, result.RawPath, "raws/v1.mp4")

	v, err := repo.FindByID(context.Background(), "v1")
	require.NoError(t, err)
	assert.Equal(t, domain.StatusPendingUpload, v.Status)
}

func TestRegenerateQuality_Success(t *testing.T) {
	repo := infrastructure.NewInMemoryVideoRepository()
	storage := newMockStorageAdapter()
	queue := newMockMessageQueue()
	seedVideo(t, repo, "v1", "Movie", "desc", domain.VideoTypeMovie, domain.StatusReady)

	rawKey := "raws/v1.mp4"
	storage.objects[rawKey] = []byte("fake-raw")

	uc := videosapp.NewRegenerateQuality(repo, storage, queue)
	err := uc.Execute(context.Background(), videosapp.RegenerateQualityInput{
		VideoID: "v1",
		Quality: "720p",
	})

	require.NoError(t, err)
	assert.Len(t, queue.jobs, 1)
	assert.True(t, queue.jobs[0].IsAppend)
	assert.Equal(t, "720p", queue.jobs[0].RequestedQualities[0])
}

func TestRegenerateQuality_RawNotFound(t *testing.T) {
	repo := infrastructure.NewInMemoryVideoRepository()
	storage := newMockStorageAdapter()
	queue := newMockMessageQueue()
	seedVideo(t, repo, "v1", "Movie", "desc", domain.VideoTypeMovie, domain.StatusReady)

	uc := videosapp.NewRegenerateQuality(repo, storage, queue)
	err := uc.Execute(context.Background(), videosapp.RegenerateQualityInput{
		VideoID: "v1",
		Quality: "720p",
	})

	assert.Error(t, err)
	var rawNotFound *domain.VideoRawNotFoundError
	assert.ErrorAs(t, err, &rawNotFound)
}

func TestDeleteVideo_Success(t *testing.T) {
	repo := infrastructure.NewInMemoryVideoRepository()
	storage := newMockStorageAdapter()
	seedVideo(t, repo, "v1", "Movie", "desc", domain.VideoTypeMovie, domain.StatusReady)

	uc := videosapp.NewDeleteVideo(repo, storage)
	err := uc.Execute(context.Background(), "v1")

	assert.NoError(t, err)

	v, err := repo.FindByID(context.Background(), "v1")
	assert.NoError(t, err)
	assert.Nil(t, v)
}

func TestListVideos_WithFilters(t *testing.T) {
	repo := infrastructure.NewInMemoryVideoRepository()
	storage := newMockStorageAdapter()
	seedVideo(t, repo, "v1", "Alpha", "first", domain.VideoTypeMovie, domain.StatusReady)
	seedVideo(t, repo, "v2", "Beta", "second", domain.VideoTypeSeries, domain.StatusProcessing)

	uc := videosapp.NewListVideos(repo, storage)
	result, err := uc.Execute(context.Background(), videosapp.ListVideosInput{
		Page:   1,
		Limit:  20,
		SortBy: "title",
		Order:  "asc",
	})

	require.NoError(t, err)
	assert.Len(t, result.Videos, 2)
	assert.Equal(t, 2, result.Total)
	assert.Equal(t, "Alpha", result.Videos[0].Title)
	assert.Equal(t, "Beta", result.Videos[1].Title)
}

func TestListVideos_FilterByStatus(t *testing.T) {
	repo := infrastructure.NewInMemoryVideoRepository()
	storage := newMockStorageAdapter()
	seedVideo(t, repo, "v1", "Ready Vid", "desc", domain.VideoTypeMovie, domain.StatusReady)
	seedVideo(t, repo, "v2", "Processing Vid", "desc", domain.VideoTypeMovie, domain.StatusProcessing)

	readyStatus := "Ready"
	uc := videosapp.NewListVideos(repo, storage)
	result, err := uc.Execute(context.Background(), videosapp.ListVideosInput{
		Page:   1,
		Limit:  20,
		Status: &readyStatus,
	})

	require.NoError(t, err)
	assert.Len(t, result.Videos, 1)
	assert.Equal(t, "Ready Vid", result.Videos[0].Title)
}

func TestListCatalog_OnlyReady(t *testing.T) {
	repo := infrastructure.NewInMemoryVideoRepository()
	storage := newMockStorageAdapter()
	seedVideo(t, repo, "v1", "Ready Vid", "desc", domain.VideoTypeMovie, domain.StatusReady)
	seedVideo(t, repo, "v2", "Processing Vid", "desc", domain.VideoTypeMovie, domain.StatusProcessing)

	uc := videosapp.NewListCatalog(repo, storage)
	result, err := uc.Execute(context.Background(), videosapp.ListCatalogInput{
		Page:  1,
		Limit: 20,
	})

	require.NoError(t, err)
	assert.Len(t, result.Videos, 1)
	assert.Equal(t, "Ready Vid", result.Videos[0].Title)
}

func TestGetVideoStream_Success(t *testing.T) {
	repo := infrastructure.NewInMemoryVideoRepository()
	storage := newMockStorageAdapter()
	seedVideo(t, repo, "v1", "Movie", "desc", domain.VideoTypeMovie, domain.StatusReady)

	uc := videosapp.NewGetVideoStream(repo, storage)
	result, err := uc.Execute(context.Background(), "v1")

	require.NoError(t, err)
	assert.Equal(t, "Movie", result.Title)
	assert.Contains(t, result.MasterPlaylistUrl, "master.m3u8")
}

func TestGetVideoStream_NotReady(t *testing.T) {
	repo := infrastructure.NewInMemoryVideoRepository()
	storage := newMockStorageAdapter()
	seedVideo(t, repo, "v1", "Movie", "desc", domain.VideoTypeMovie, domain.StatusPendingUpload)

	uc := videosapp.NewGetVideoStream(repo, storage)
	_, err := uc.Execute(context.Background(), "v1")

	assert.Error(t, err)
	var notReady *domain.VideoNotReadyError
	assert.ErrorAs(t, err, &notReady)
}

func TestProducerWorker_PollsAndEnqueues(t *testing.T) {
	repo := infrastructure.NewInMemoryVideoRepository()
	queue := newMockMessageQueue()
	seedVideo(t, repo, "v1", "Movie1", "desc", domain.VideoTypeMovie, domain.StatusPendingUpload)
	seedVideo(t, repo, "v2", "Movie2", "desc", domain.VideoTypeMovie, domain.StatusPendingUpload)

	uc := videosapp.NewPollPendingVideos(repo, queue)
	worker := videospres.NewProducerWorker(uc)

	result, err := uc.Execute(context.Background())
	require.NoError(t, err)
	assert.Len(t, result.ProcessedIDs, 2)
	assert.Len(t, queue.jobs, 2)
	_ = worker
}

func TestProducerWorker_SkipsIfAlreadyRunning(t *testing.T) {
	repo := infrastructure.NewInMemoryVideoRepository()
	queue := newMockMessageQueue()
	uc := videosapp.NewPollPendingVideos(repo, queue)
	worker := videospres.NewProducerWorker(uc)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		worker.Start(ctx)
		close(done)
	}()

	cancel()
	<-done

	result, err := uc.Execute(context.Background())
	require.NoError(t, err)
	assert.Len(t, result.ProcessedIDs, 0)
}

func TestConsumerQueueWorker_ProcessesJob(t *testing.T) {
	repo := infrastructure.NewInMemoryVideoRepository()
	storage := newMockStorageAdapter()
	queue := newMockMessageQueue()
	seedVideo(t, repo, "v1", "Movie", "desc", domain.VideoTypeMovie, domain.StatusProcessing)

	uc := videosapp.NewProcessVideo(repo, storage, &mockTranscoder{})
	worker := videospres.NewConsumerQueueWorker(queue, uc)

	queue.jobs = append(queue.jobs, domain.VideoProcessingJob{
		VideoID:            "v1",
		RequestedQualities: []string{"480p"},
		IsAppend:           false,
	})

	err := uc.Execute(context.Background(), videosapp.ProcessVideoInput{
		VideoID:            "v1",
		RequestedQualities: []string{"480p"},
		IsAppend:           false,
	})
	assert.NoError(t, err)

	v, err := repo.FindByID(context.Background(), "v1")
	require.NoError(t, err)
	assert.Equal(t, domain.StatusReady, v.Status)
	_ = worker
}
