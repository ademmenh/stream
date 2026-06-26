package videotests

	import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-starter/internal/auth"
	authapp "go-starter/internal/auth/application"
	authinfra "go-starter/internal/auth/infrastructure"
	authpresentation "go-starter/internal/auth/presentation"
	configinfra "go-starter/internal/config/infrastructure"
	shareddomain "go-starter/internal/shared/domain"
	"go-starter/internal/shared/infrastructure/ent/generated/video"
	presentation "go-starter/internal/shared/presentation"
	"go-starter/internal/users"
	usersdomain "go-starter/internal/users/domain"
	usersinfra "go-starter/internal/users/infrastructure"
	userspresentation "go-starter/internal/users/presentation"
	"go-starter/internal/videos"
	videosapp "go-starter/internal/videos/application"
	videosinfra "go-starter/internal/videos/infrastructure"
	videospres "go-starter/internal/videos/presentation"
)

type videoResp struct {
	Message    string                      `json:"message"`
	StatusCode int                         `json:"statusCode"`
	Data       *videosapp.CreateVideoOutput `json:"data"`
}

type videoListResp struct {
	Message    string                    `json:"message"`
	StatusCode int                       `json:"statusCode"`
	Data       []videosapp.VideoOutput   `json:"data"`
	Pagination struct {
		Total int `json:"total"`
		Page  int `json:"page"`
		Limit int `json:"limit"`
	} `json:"pagination"`
}

type videoStreamResp struct {
	Message    string                        `json:"message"`
	StatusCode int                           `json:"statusCode"`
	Data       *videosapp.VideoStreamOutput  `json:"data"`
}

type errResp struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
	Error      string `json:"error"`
}

func setupVideoE2EServer(t *testing.T) (*httptest.Server, string) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping e2e test")
	}
	if globalClient == nil {
		t.Fatal("postgres not available")
	}

	globalClient.UserSchema.Delete().ExecX(context.Background())
	globalClient.Video.Delete().ExecX(context.Background())
	t.Cleanup(func() {
		globalClient.UserSchema.Delete().ExecX(context.Background())
		globalClient.Video.Delete().ExecX(context.Background())
	})

	cfg := configinfra.NewConfigAdapter()
	jwtAdapter := authinfra.NewJwtAdapter(cfg)
	passwordAdapter := authinfra.NewPasswordAdapter()
	userRepo := usersinfra.NewUserRepository(globalClient)
	idGen := usersinfra.NewIDGenerator()

	e := echo.New()
	e.HTTPErrorHandler = presentation.CustomHTTPErrorHandler
	presentation.AuthErrorHandler = authpresentation.AuthErrorHandler
	presentation.UserErrorHandler = userspresentation.UserErrorHandler
	presentation.VideoErrorHandler = videospres.VideoErrorHandler

	v1 := e.Group("/api/v1")

	authModule := auth.NewModule(auth.Dependencies{
		JwtAdapter:      jwtAdapter,
		PasswordAdapter: passwordAdapter,
		IDGenerator:     idGen,
		UserRepo:        userRepo,
		Config:          cfg,
	})
	authModule.RegisterRoutes(v1.Group("/auth"))

	usersModule := users.NewModule(users.Dependencies{
		UserRepo:        userRepo,
		PasswordAdapter: passwordAdapter,
	})
	usersModule.RegisterRoutes(v1.Group("/users"), cfg.JWTAccessTokenSecret())

	videoRepo := videosinfra.NewVideoRepository(globalClient)
	videoStorage := videosinfra.NewVideoStorageAdapter(globalS3Adapter)
	videoQueue := videosinfra.NewInMemoryQueueAdapter()

	videosModule := videos.NewModule(videos.Dependencies{
		VideoRepo:    videoRepo,
		Storage:      videoStorage,
		MessageQueue: videoQueue,
		IDGenerator:  idGen,
	})
	videosModule.RegisterRoutes(v1, cfg.JWTAccessTokenSecret())

	ctx, cancel := context.WithCancel(context.Background())
	ffmpegAdapter := videosinfra.NewFfmpegAdapter()
	processVideoUC := videosapp.NewProcessVideo(videoRepo, videoStorage, ffmpegAdapter)
	consumer := videospres.NewConsumerQueueWorker(videoQueue, processVideoUC)
	go consumer.Start(ctx)
	t.Cleanup(func() { cancel() })

	ts := httptest.NewServer(e)

	adminToken := getAdminToken(t, idGen, passwordAdapter, userRepo, jwtAdapter)

	return ts, adminToken
}

func getAdminToken(t *testing.T, idGen *usersinfra.IDGen, passwordAdapter authapp.PasswordAdapter, userRepo *usersinfra.UserRepository, jwtAdapter authapp.JwtAdapter) string {
	t.Helper()

	ctx := context.Background()

	email, _ := shareddomain.NewEmail("admin@test.com")
	hashed, _ := passwordAdapter.Hash("admin123")
	_, err := userRepo.Create(ctx, &usersdomain.User{
		ID:           shareddomain.IdFromStr(idGen.Generate()),
		Name:         "Admin",
		Email:        email,
		PasswordHash: hashed,
		Role:         "admin",
	})
	require.NoError(t, err)

	payload := authapp.TokenPayload{
		Sub:   email.String(),
		Email: email.String(),
		Role:  "admin",
	}
	accessToken, err := jwtAdapter.Sign(payload)
	require.NoError(t, err)

	return accessToken
}

func doRequest(t *testing.T, method, url, token string, body []byte) *http.Response {
	t.Helper()
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewBuffer(body)
	}
	req, err := http.NewRequest(method, url, reqBody)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func TestVideoE2E_CreateAndListAdmin(t *testing.T) {
	ts, adminToken := setupVideoE2EServer(t)
	defer ts.Close()

	createBody := `{"title":"E2E Movie","description":"e2e test","type":"movie"}`
	resp := doRequest(t, http.MethodPost, ts.URL+"/api/v1/admin/videos", adminToken, []byte(createBody))
	defer resp.Body.Close()

	var created videoResp
	err := json.NewDecoder(resp.Body).Decode(&created)
	require.NoError(t, err)
	assert.Equal(t, 201, created.StatusCode)
	assert.NotEmpty(t, created.Data.ID)
	assert.Contains(t, created.Data.VideoUploadUrl, "raws/")
	assert.Contains(t, created.Data.PhotoUploadUrl, "photos/")

	listResp := doRequest(t, http.MethodGet, ts.URL+"/api/v1/admin/videos", adminToken, nil)
	defer listResp.Body.Close()

	var list videoListResp
	err = json.NewDecoder(listResp.Body).Decode(&list)
	require.NoError(t, err)
	assert.Equal(t, 200, list.StatusCode)
	assert.GreaterOrEqual(t, list.Pagination.Total, 1)
	assert.Equal(t, "E2E Movie", list.Data[0].Title)
}

func TestVideoE2E_AdminWithoutToken(t *testing.T) {
	ts, _ := setupVideoE2EServer(t)
	defer ts.Close()

	resp := doRequest(t, http.MethodGet, ts.URL+"/api/v1/admin/videos", "", nil)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestVideoE2E_ProcessFlow(t *testing.T) {
	ts, adminToken := setupVideoE2EServer(t)
	defer ts.Close()

	createBody := `{"title":"Processing Test","description":"test","type":"movie"}`
	resp := doRequest(t, http.MethodPost, ts.URL+"/api/v1/admin/videos", adminToken, []byte(createBody))
	defer resp.Body.Close()

	var created videoResp
	err := json.NewDecoder(resp.Body).Decode(&created)
	require.NoError(t, err)

	processBody := `{"requested_qualities":["480p","1080p"]}`
	resp = doRequest(t, http.MethodPost, ts.URL+"/api/v1/admin/videos/"+created.Data.ID+"/process", adminToken, []byte(processBody))
	defer resp.Body.Close()
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)

	time.Sleep(100 * time.Millisecond)

	listResp := doRequest(t, http.MethodGet, ts.URL+"/api/v1/admin/videos?status=Processing", adminToken, nil)
	defer listResp.Body.Close()

	var list videoListResp
	err = json.NewDecoder(listResp.Body).Decode(&list)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, list.Pagination.Total, 1)
	assert.Equal(t, "Processing Test", list.Data[0].Title)
	assert.Equal(t, "Processing", list.Data[0].Status)
}

func TestVideoE2E_DeleteVideo(t *testing.T) {
	ts, adminToken := setupVideoE2EServer(t)
	defer ts.Close()

	createBody := `{"title":"Delete Me","description":"test","type":"documentary"}`
	resp := doRequest(t, http.MethodPost, ts.URL+"/api/v1/admin/videos", adminToken, []byte(createBody))
	defer resp.Body.Close()

	var created videoResp
	err := json.NewDecoder(resp.Body).Decode(&created)
	require.NoError(t, err)

	resp = doRequest(t, http.MethodDelete, ts.URL+"/api/v1/admin/videos/"+created.Data.ID, adminToken, nil)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp = doRequest(t, http.MethodGet, ts.URL+"/api/v1/admin/videos?status=PendingUpload", adminToken, nil)
	defer resp.Body.Close()
	var list videoListResp
	err = json.NewDecoder(resp.Body).Decode(&list)
	require.NoError(t, err)
	assert.Equal(t, 0, list.Pagination.Total)
}

func TestVideoE2E_UserCatalog(t *testing.T) {
	ts, _ := setupVideoE2EServer(t)
	defer ts.Close()

	resp := doRequest(t, http.MethodGet, ts.URL+"/api/v1/videos", "", nil)
	defer resp.Body.Close()

	var catalog videoListResp
	err := json.NewDecoder(resp.Body).Decode(&catalog)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotNil(t, catalog.Data)
}

func TestVideoE2E_ReplaceVideo(t *testing.T) {
	ts, adminToken := setupVideoE2EServer(t)
	defer ts.Close()

	createBody := `{"title":"Replace Test","description":"test","type":"movie"}`
	resp := doRequest(t, http.MethodPost, ts.URL+"/api/v1/admin/videos", adminToken, []byte(createBody))
	defer resp.Body.Close()

	var created videoResp
	err := json.NewDecoder(resp.Body).Decode(&created)
	require.NoError(t, err)

	_, err = globalClient.Video.UpdateOneID(uuid.MustParse(created.Data.ID)).
		SetStatus(video.StatusReady).
		Save(context.Background())
	require.NoError(t, err)

	resp = doRequest(t, http.MethodPut, ts.URL+"/api/v1/admin/videos/"+created.Data.ID+"/replace", adminToken, nil)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var replace videoResp
	err = json.NewDecoder(resp.Body).Decode(&replace)
	require.NoError(t, err)
	assert.Contains(t, replace.Data.VideoUploadUrl, "raws/"+created.Data.ID+".mp4")
}

func TestVideoE2E_GetVideoStream(t *testing.T) {
	ts, adminToken := setupVideoE2EServer(t)
	defer ts.Close()

	createBody := `{"title":"Stream Test","description":"test","type":"movie"}`
	resp := doRequest(t, http.MethodPost, ts.URL+"/api/v1/admin/videos", adminToken, []byte(createBody))
	defer resp.Body.Close()

	var created videoResp
	err := json.NewDecoder(resp.Body).Decode(&created)
	require.NoError(t, err)

	resp = doRequest(t, http.MethodGet, ts.URL+"/api/v1/videos/"+created.Data.ID, "", nil)
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	var streamErr errResp
	err = json.Unmarshal(bodyBytes, &streamErr)
	require.NoError(t, err)
	assert.Equal(t, 409, streamErr.StatusCode)
}
