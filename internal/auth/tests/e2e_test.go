package authtests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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
	usersinfra "go-starter/internal/users/infrastructure"
	userspresentation "go-starter/internal/users/presentation"
)

func setupAuthE2EServer(t *testing.T) *httptest.Server {
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

	authModule := auth.NewModule(auth.Dependencies{
		JwtAdapter:      jwtAdapter,
		PasswordAdapter: passwordAdapter,
		IDGenerator:     idGen,
		UserRepo:        userRepo,
		Config:          cfg,
	})
	authModule.RegisterRoutes(e.Group("/api/v1/auth"))

	return httptest.NewServer(e)
}

type authResp struct {
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

type tokensResp struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
	Data       struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	} `json:"data"`
}

func TestAuthE2E_RegisterAndLogin(t *testing.T) {
	ts := setupAuthE2EServer(t)
	defer ts.Close()

	regBody := `{"name":"Alice","email":"alice@e2e.com","password":"password123"}`
	resp, err := http.Post(ts.URL+"/api/v1/auth/register", "application/json", bytes.NewBufferString(regBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	var reg authResp
	err = json.NewDecoder(resp.Body).Decode(&reg)
	require.NoError(t, err)
	assert.Equal(t, 201, reg.StatusCode)
	assert.Equal(t, "Alice", reg.Data.Name)
	assert.NotEmpty(t, reg.Tokens.AccessToken)
	assert.NotEmpty(t, reg.Tokens.RefreshToken)

	loginBody := `{"email":"alice@e2e.com","password":"password123"}`
	resp, err = http.Post(ts.URL+"/api/v1/auth/login", "application/json", bytes.NewBufferString(loginBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	var login authResp
	err = json.NewDecoder(resp.Body).Decode(&login)
	require.NoError(t, err)
	assert.Equal(t, 200, login.StatusCode)
	assert.Equal(t, "Alice", login.Data.Name)
	assert.NotEmpty(t, login.Tokens.AccessToken)
	assert.NotEmpty(t, login.Tokens.RefreshToken)
}

func TestAuthE2E_DuplicateRegister(t *testing.T) {
	ts := setupAuthE2EServer(t)
	defer ts.Close()

	body := `{"name":"Alice","email":"dup@e2e.com","password":"password123"}`
	resp, err := http.Post(ts.URL+"/api/v1/auth/register", "application/json", bytes.NewBufferString(body))
	require.NoError(t, err)
	resp.Body.Close()

	resp, err = http.Post(ts.URL+"/api/v1/auth/register", "application/json", bytes.NewBufferString(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	var errResp struct {
		Message    string `json:"message"`
		StatusCode int    `json:"statusCode"`
		Error      string `json:"error"`
	}
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, 409, errResp.StatusCode)
}

func TestAuthE2E_LoginWrongPassword(t *testing.T) {
	ts := setupAuthE2EServer(t)
	defer ts.Close()

	regBody := `{"name":"Alice","email":"alice@e2e.com","password":"password123"}`
	resp, err := http.Post(ts.URL+"/api/v1/auth/register", "application/json", bytes.NewBufferString(regBody))
	require.NoError(t, err)
	resp.Body.Close()

	loginBody := `{"email":"alice@e2e.com","password":"wrongpassword"}`
	resp, err = http.Post(ts.URL+"/api/v1/auth/login", "application/json", bytes.NewBufferString(loginBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	var errResp struct {
		Message    string `json:"message"`
		StatusCode int    `json:"statusCode"`
		Error      string `json:"error"`
	}
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, 401, errResp.StatusCode)
}

func TestAuthE2E_RefreshToken(t *testing.T) {
	ts := setupAuthE2EServer(t)
	defer ts.Close()

	regBody := `{"name":"Alice","email":"alice@e2e.com","password":"password123"}`
	resp, err := http.Post(ts.URL+"/api/v1/auth/register", "application/json", bytes.NewBufferString(regBody))
	require.NoError(t, err)

	var reg authResp
	err = json.NewDecoder(resp.Body).Decode(&reg)
	require.NoError(t, err)
	resp.Body.Close()

	req, _ := http.NewRequest("POST", ts.URL+"/api/v1/auth/refresh", nil)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: reg.Tokens.RefreshToken})

	client := &http.Client{}
	refreshResp, err := client.Do(req)
	require.NoError(t, err)
	defer refreshResp.Body.Close()

	var refresh tokensResp
	body, err := io.ReadAll(refreshResp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &refresh)
	require.NoError(t, err)
	assert.Equal(t, 200, refresh.StatusCode)
	assert.NotEmpty(t, refresh.Data.AccessToken)
	assert.NotEmpty(t, refresh.Data.RefreshToken)
}
