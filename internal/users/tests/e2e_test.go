package userstests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-starter/internal/auth"
	authinfra "go-starter/internal/auth/infrastructure"
	authpresentation "go-starter/internal/auth/presentation"
	configinfra "go-starter/internal/config/infrastructure"
	presentation "go-starter/internal/shared/presentation"
	"go-starter/internal/users"
	usersinfra "go-starter/internal/users/infrastructure"
	userspresentation "go-starter/internal/users/presentation"
)

func setupUsersE2EServer(t *testing.T) *httptest.Server {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping e2e test")
	}
	if globalClient == nil {
		t.Fatal("postgres not available — create .env.test at project root with DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, DB_SSLMODE")
	}

	globalClient.UserSchema.Delete().ExecX(context.Background())
	t.Cleanup(func() { globalClient.UserSchema.Delete().ExecX(context.Background()) })

	cfg := configinfra.NewConfigAdapter()
	jwtAdapter := authinfra.NewJwtAdapter(cfg)
	passwordAdapter := authinfra.NewPasswordAdapter()
	userRepo := usersinfra.NewUserRepository(globalClient)
	idGen := usersinfra.NewIDGenerator()

	e := echo.New()
	e.HTTPErrorHandler = presentation.CustomHTTPErrorHandler
	presentation.AuthErrorHandler = authpresentation.AuthErrorHandler
	presentation.UserErrorHandler = userspresentation.UserErrorHandler

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

	return httptest.NewServer(e)
}

type registerResponse struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
	Data       struct {
		Id    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
		Role  string `json:"role"`
	} `json:"data"`
	Tokens struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	} `json:"tokens"`
}

type apiResponse struct {
	Message    string          `json:"message"`
	StatusCode int             `json:"statusCode"`
	Data       json.RawMessage `json:"data"`
}

func TestUsersE2E_GetCurrentUser(t *testing.T) {
	ts := setupUsersE2EServer(t)
	defer ts.Close()

	regBody := `{"name":"Alice","email":"alice@e2e.com","password":"password123"}`
	resp, err := http.Post(ts.URL+"/api/v1/auth/register", "application/json", bytes.NewBufferString(regBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	var regResp registerResponse
	err = json.NewDecoder(resp.Body).Decode(&regResp)
	require.NoError(t, err)
	require.Equal(t, 201, regResp.StatusCode)

	req, _ := http.NewRequest("GET", ts.URL+"/api/v1/users/@me", nil)
	req.Header.Set("Authorization", "Bearer "+regResp.Tokens.AccessToken)

	client := &http.Client{}
	meResp, err := client.Do(req)
	require.NoError(t, err)
	defer meResp.Body.Close()

	var me apiResponse
	err = json.NewDecoder(meResp.Body).Decode(&me)
	require.NoError(t, err)
	assert.Equal(t, 200, me.StatusCode)
}

func TestUsersE2E_UpdateCurrentUser(t *testing.T) {
	ts := setupUsersE2EServer(t)
	defer ts.Close()

	regBody := `{"name":"Bob","email":"bob@e2e.com","password":"password123"}`
	resp, err := http.Post(ts.URL+"/api/v1/auth/register", "application/json", bytes.NewBufferString(regBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	var regResp registerResponse
	err = json.NewDecoder(resp.Body).Decode(&regResp)
	require.NoError(t, err)

	updateBody := `{"name":"Bob Updated"}`
	req, _ := http.NewRequest("PUT", ts.URL+"/api/v1/users/@me", bytes.NewBufferString(updateBody))
	req.Header.Set("Authorization", "Bearer "+regResp.Tokens.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	updResp, err := client.Do(req)
	require.NoError(t, err)
	defer updResp.Body.Close()

	var upd apiResponse
	err = json.NewDecoder(updResp.Body).Decode(&upd)
	require.NoError(t, err)
	assert.Equal(t, 200, upd.StatusCode)
}

func TestUsersE2E_DeleteCurrentUser(t *testing.T) {
	ts := setupUsersE2EServer(t)
	defer ts.Close()

	regBody := `{"name":"Charlie","email":"charlie@e2e.com","password":"password123"}`
	resp, err := http.Post(ts.URL+"/api/v1/auth/register", "application/json", bytes.NewBufferString(regBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	var regResp registerResponse
	err = json.NewDecoder(resp.Body).Decode(&regResp)
	require.NoError(t, err)

	req, _ := http.NewRequest("DELETE", ts.URL+"/api/v1/users/@me", nil)
	req.Header.Set("Authorization", "Bearer "+regResp.Tokens.AccessToken)

	client := &http.Client{}
	delResp, err := client.Do(req)
	require.NoError(t, err)
	defer delResp.Body.Close()

	var del apiResponse
	err = json.NewDecoder(delResp.Body).Decode(&del)
	require.NoError(t, err)
	assert.Equal(t, 200, del.StatusCode)
}
