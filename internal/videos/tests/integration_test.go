package videotests

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/lib/pq"

	shareddomain "go-starter/internal/shared/domain"
	"go-starter/internal/shared/infrastructure/ent/generated"
	sharedinfra "go-starter/internal/shared/infrastructure"
	sharedtests "go-starter/internal/shared/tests"
	configinfra "go-starter/internal/config/infrastructure"
	"go-starter/internal/videos/domain"
	"go-starter/internal/videos/infrastructure"
)

var (
	globalClient    *generated.Client
	globalS3Adapter *sharedinfra.S3Adapter
)

func init() {
	sharedtests.LoadDotEnv()
}

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		os.Exit(m.Run())
	}

	dsn := sharedtests.BuildDSN()
	client, err := generated.Open("postgres", dsn)
	if err == nil {
		defer client.Close()
		ctx := context.Background()
		if err := client.Schema.Create(ctx); err == nil {
			client.Video.Delete().ExecX(ctx)
			globalClient = client
		}
	}

	origBucket := os.Getenv("S3_BUCKET")
	testBucket := origBucket + "-test"
	os.Setenv("S3_BUCKET", testBucket)
	defer os.Setenv("S3_BUCKET", origBucket)

	s3Adapter, err := sharedinfra.NewS3Adapter(configinfra.NewConfigAdapter())
	if err == nil {
		globalS3Adapter = s3Adapter
	}

	if globalS3Adapter != nil {
		globalS3Adapter.RemoveBucket(context.Background())
		globalS3Adapter.Init(context.Background())
	}

	code := m.Run()

	if globalS3Adapter != nil {
		globalS3Adapter.RemoveBucket(context.Background())
	}

	os.Exit(code)
}

func setupIntegrationRepo(t *testing.T) *infrastructure.VideoRepository {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	if globalClient == nil {
		t.Fatal("postgres not available")
	}

	globalClient.Video.Delete().ExecX(context.Background())
	t.Cleanup(func() { globalClient.Video.Delete().ExecX(context.Background()) })

	return infrastructure.NewVideoRepository(globalClient)
}

func setupIntegrationStorage(t *testing.T) *infrastructure.VideoStorageAdapter {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	if globalS3Adapter == nil {
		t.Skip("s3 not available — set S3_HOST, S3_PORT, S3_ACCESS_KEY, S3_SECRET_KEY, S3_BUCKET in .env.test")
	}

	return infrastructure.NewVideoStorageAdapter(globalS3Adapter)
}

func newTestVideo(title string) *domain.Video {
	return &domain.Video{
		ID:          shareddomain.NewId(),
		Title:       title,
		Description: "test description",
		Type:        domain.VideoTypeMovie,
		Status:      domain.StatusPendingUpload,
		Qualities:   []domain.VideoQuality{},
	}
}

func TestVideoRepository_Create_Integration(t *testing.T) {
	repo := setupIntegrationRepo(t)
	v := newTestVideo("Test Movie")

	created, err := repo.Create(context.Background(), v)
	require.NoError(t, err)

	assert.Equal(t, "Test Movie", created.Title)
	assert.Equal(t, domain.VideoTypeMovie, created.Type)
	assert.Equal(t, domain.StatusPendingUpload, created.Status)
	assert.NotEmpty(t, created.GetID())
}

func TestVideoRepository_FindByID_Integration(t *testing.T) {
	repo := setupIntegrationRepo(t)
	v := newTestVideo("Find Me")
	created, err := repo.Create(context.Background(), v)
	require.NoError(t, err)

	found, err := repo.FindByID(context.Background(), created.GetID())
	require.NoError(t, err)
	assert.Equal(t, "Find Me", found.Title)
}

func TestVideoRepository_FindByID_NotFound_Integration(t *testing.T) {
	repo := setupIntegrationRepo(t)

	found, err := repo.FindByID(context.Background(), "00000000-0000-0000-0000-000000000000")
	assert.NoError(t, err)
	assert.Nil(t, found)
}

func TestVideoRepository_List_Integration(t *testing.T) {
	repo := setupIntegrationRepo(t)
	repo.Create(context.Background(), newTestVideo("Alpha"))
	repo.Create(context.Background(), newTestVideo("Beta"))
	repo.Create(context.Background(), newTestVideo("Gamma"))

	filter := domain.VideoListFilter{Page: 1, Limit: 10, SortBy: "title", Order: "asc"}
	videos, total, err := repo.List(context.Background(), filter)

	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, videos, 3)
	assert.Equal(t, "Alpha", videos[0].Title)
}

func TestVideoRepository_List_FilterByType_Integration(t *testing.T) {
	repo := setupIntegrationRepo(t)
	movie := newTestVideo("Movie")
	movie.Type = domain.VideoTypeMovie
	series := newTestVideo("Series")
	series.Type = domain.VideoTypeSeries
	repo.Create(context.Background(), movie)
	repo.Create(context.Background(), series)

	filterType := domain.VideoTypeMovie
	filter := domain.VideoListFilter{Page: 1, Limit: 10, Type: &filterType}
	videos, total, err := repo.List(context.Background(), filter)

	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, "Movie", videos[0].Title)
}

func TestVideoRepository_Update_Integration(t *testing.T) {
	repo := setupIntegrationRepo(t)
	v := newTestVideo("Original")
	created, err := repo.Create(context.Background(), v)
	require.NoError(t, err)

	created.Title = "Updated"
	created.Description = "updated desc"
	updated, err := repo.Update(context.Background(), created)

	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Title)
	assert.Equal(t, "updated desc", updated.Description)
}

func TestVideoRepository_Delete_Integration(t *testing.T) {
	repo := setupIntegrationRepo(t)
	v := newTestVideo("Delete Me")
	created, err := repo.Create(context.Background(), v)
	require.NoError(t, err)

	err = repo.Delete(context.Background(), created.GetID())
	assert.NoError(t, err)

	found, err := repo.FindByID(context.Background(), created.GetID())
	assert.NoError(t, err)
	assert.Nil(t, found)
}

func TestVideoStorageAdapter_UploadAndGetSignedUrl_Integration(t *testing.T) {
	storage := setupIntegrationStorage(t)
	ctx := context.Background()

	key := "raws/test-video.mp4"
	body := []byte("fake-video-content")

	_, err := storage.UploadFile(ctx, key, body, "application/octet-stream")
	require.NoError(t, err)

	signedUrl, err := storage.GeneratePresignedGetUrl(ctx, key, 0)
	require.NoError(t, err)
	assert.Contains(t, signedUrl, "test-video.mp4")
}

func TestVideoStorageAdapter_ObjectExists_Integration(t *testing.T) {
	storage := setupIntegrationStorage(t)
	ctx := context.Background()

	key := "raws/exists-test.mp4"
	_, err := storage.UploadFile(ctx, key, []byte("data"), "application/octet-stream")
	require.NoError(t, err)

	exists, err := storage.ObjectExists(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)

	missing, err := storage.ObjectExists(ctx, "raws/nonexistent.mp4")
	require.NoError(t, err)
	assert.False(t, missing)
}

func TestVideoStorageAdapter_DeletePrefixes_Integration(t *testing.T) {
	storage := setupIntegrationStorage(t)
	ctx := context.Background()

	storage.UploadFile(ctx, "videos/test-id/480p/seg-1.ts", []byte("ts"), "video/mp2t")
	storage.UploadFile(ctx, "videos/test-id/master.m3u8", []byte("m3u8"), "application/vnd.apple.mpegurl")
	storage.UploadFile(ctx, "photos/test-id/thumbnail.jpg", []byte("jpg"), "image/jpeg")
	storage.UploadFile(ctx, "raws/test-id.mp4", []byte("mp4"), "application/octet-stream")

	err := storage.DeletePrefixes(ctx, []string{
		"videos/test-id/480p/seg-1.ts",
		"videos/test-id/master.m3u8",
		"photos/test-id/thumbnail.jpg",
		"raws/test-id.mp4",
	})
	assert.NoError(t, err)

	exists, err := storage.ObjectExists(ctx, "raws/test-id.mp4")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestQueueAdapter_EnqueueDequeue_Integration(t *testing.T) {
	queue := infrastructure.NewInMemoryQueueAdapter()
	ctx := context.Background()

	job := domain.VideoProcessingJob{
		VideoID:            "test-video",
		RequestedQualities: []string{"480p", "1080p"},
		IsAppend:           false,
	}

	err := queue.Enqueue(ctx, job)
	require.NoError(t, err)

	dequeued, err := queue.Dequeue(ctx)
	require.NoError(t, err)
	require.NotNil(t, dequeued)
	assert.Equal(t, "test-video", dequeued.VideoID)
	assert.Equal(t, []string{"480p", "1080p"}, dequeued.RequestedQualities)
	assert.False(t, dequeued.IsAppend)

	dequeued, err = queue.Dequeue(ctx)
	require.NoError(t, err)
	assert.Nil(t, dequeued)
}
