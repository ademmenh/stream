package authtests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authapp "go-starter/internal/auth/application"
	authdomain "go-starter/internal/auth/domain"
	shareddomain "go-starter/internal/shared/domain"
	usersdomain "go-starter/internal/users/domain"
	"go-starter/internal/users/infrastructure"
)

type mockJwtAdapter struct{}

func (m *mockJwtAdapter) Sign(payload authdomain.TokenPayload) (string, error) {
	return "access-token", nil
}

func (m *mockJwtAdapter) Verify(token string) (*authdomain.TokenPayload, error) {
	return &authdomain.TokenPayload{Sub: "test-id", Email: "test@test.com", Role: "client"}, nil
}

func (m *mockJwtAdapter) SignRefresh(payload authdomain.TokenPayload) (string, error) {
	return "refresh-token", nil
}

func (m *mockJwtAdapter) VerifyRefresh(token string) (*authdomain.TokenPayload, error) {
	return &authdomain.TokenPayload{Sub: "test-id", Email: "test@test.com", Role: "client"}, nil
}

type mockPasswordAdapter struct{}

func (m *mockPasswordAdapter) Hash(plain string) (string, error) {
	return "hashed-" + plain, nil
}

func (m *mockPasswordAdapter) Compare(plain, hashed string) bool {
	return hashed == "hashed-"+plain
}

type mockIDGenerator struct{}

func (m *mockIDGenerator) Generate() string {
	return "gen-id-1"
}

func seedAuthUser(t *testing.T, repo *infrastructure.InMemoryUserRepository, id, name, email, password string) {
	t.Helper()
	e, _ := shareddomain.NewEmail(email)
	_, err := repo.Create(context.Background(), &usersdomain.User{
		ID:           shareddomain.IdFromStr(id),
		Name:         name,
		Email:        e,
		PasswordHash: "hashed-" + password,
		Role:         "client",
	})
	require.NoError(t, err)
}

func TestLogin_Success(t *testing.T) {
	repo := infrastructure.NewInMemoryUserRepository()
	seedAuthUser(t, repo, "u1", "Alice", "alice@test.com", "password123")

	uc := authapp.NewLogin(repo, &mockPasswordAdapter{}, &mockJwtAdapter{})
	result, err := uc.Execute(context.Background(), authapp.LoginInput{
		Email:    "alice@test.com",
		Password: "password123",
	})

	require.NoError(t, err)
	assert.Equal(t, "Alice", result.User.Name)
	assert.Equal(t, "alice@test.com", result.User.Email)
	assert.NotEmpty(t, result.Tokens.AccessToken)
	assert.NotEmpty(t, result.Tokens.RefreshToken)
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := infrastructure.NewInMemoryUserRepository()
	seedAuthUser(t, repo, "u1", "Alice", "alice@test.com", "password123")

	uc := authapp.NewLogin(repo, &mockPasswordAdapter{}, &mockJwtAdapter{})
	_, err := uc.Execute(context.Background(), authapp.LoginInput{
		Email:    "alice@test.com",
		Password: "wrongpassword",
	})

	assert.Error(t, err)
	var invalidErr *authdomain.InvalidCredentialsError
	assert.ErrorAs(t, err, &invalidErr)
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := infrastructure.NewInMemoryUserRepository()

	uc := authapp.NewLogin(repo, &mockPasswordAdapter{}, &mockJwtAdapter{})
	_, err := uc.Execute(context.Background(), authapp.LoginInput{
		Email:    "nonexistent@test.com",
		Password: "password123",
	})

	assert.Error(t, err)
	var invalidErr *authdomain.InvalidCredentialsError
	assert.ErrorAs(t, err, &invalidErr)
}

func TestRegister_Success(t *testing.T) {
	repo := infrastructure.NewInMemoryUserRepository()

	phone := "+1234567890"
	uc := authapp.NewRegister(repo, &mockPasswordAdapter{}, &mockIDGenerator{}, &mockJwtAdapter{})
	result, err := uc.Execute(context.Background(), authapp.RegisterInput{
		Name:     "Bob",
		Email:    "bob@test.com",
		Password: "password123",
		Phone:    &phone,
	})

	require.NoError(t, err)
	assert.Equal(t, "Bob", result.User.Name)
	assert.Equal(t, "bob@test.com", result.User.Email)
	assert.NotEmpty(t, result.Tokens.AccessToken)
	assert.NotEmpty(t, result.Tokens.RefreshToken)

	created, err := repo.FindByEmail(context.Background(), "bob@test.com")
	require.NoError(t, err)
	assert.Equal(t, "Bob", created.Name)
}

func TestRegister_DuplicateEmail(t *testing.T) {
	repo := infrastructure.NewInMemoryUserRepository()
	seedAuthUser(t, repo, "u1", "Alice", "alice@test.com", "password123")

	uc := authapp.NewRegister(repo, &mockPasswordAdapter{}, &mockIDGenerator{}, &mockJwtAdapter{})
	_, err := uc.Execute(context.Background(), authapp.RegisterInput{
		Name:     "Alice Duplicate",
		Email:    "alice@test.com",
		Password: "password456",
	})

	assert.Error(t, err)
	var dupErr *usersdomain.UserEmailAlreadyExistsError
	assert.ErrorAs(t, err, &dupErr)
	assert.Contains(t, dupErr.Error(), "alice@test.com")
}

func TestRefreshToken_Success(t *testing.T) {
	repo := infrastructure.NewInMemoryUserRepository()
	seedAuthUser(t, repo, "test-id", "Alice", "test@test.com", "password123")

	uc := authapp.NewRefreshToken(repo, &mockJwtAdapter{})
	result, err := uc.Execute(context.Background(), authapp.RefreshTokenInput{
		RefreshToken: "valid-refresh-token",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
}

type failingJwtAdapter struct {
	mockJwtAdapter
}

func (f *failingJwtAdapter) VerifyRefresh(token string) (*authdomain.TokenPayload, error) {
	return nil, assert.AnError
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	seedAuthUser(t, infrastructure.NewInMemoryUserRepository(), "test-id", "Alice", "test@test.com", "password123")

	uc := authapp.NewRefreshToken(infrastructure.NewInMemoryUserRepository(), &failingJwtAdapter{})
	_, err := uc.Execute(context.Background(), authapp.RefreshTokenInput{
		RefreshToken: "invalid",
	})

	assert.Error(t, err)
	var refreshErr *authdomain.RefreshTokenInvalidError
	assert.ErrorAs(t, err, &refreshErr)
}
