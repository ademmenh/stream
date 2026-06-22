package authtests

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/lib/pq"

	"go-starter/internal/shared/infrastructure/ent/generated"

	authapp "go-starter/internal/auth/application"
	authdomain "go-starter/internal/auth/domain"
	authinfra "go-starter/internal/auth/infrastructure"
	configinfra "go-starter/internal/config/infrastructure"
	shareddomain "go-starter/internal/shared/domain"
	usersdomain "go-starter/internal/users/domain"
	usersinfra "go-starter/internal/users/infrastructure"
)

var globalClient *generated.Client

func init() {
	loadDotEnv()
}

func loadDotEnv() {
	dir, err := os.Getwd()
	if err != nil {
		return
	}
	for {
		envPath := filepath.Join(dir, ".env.test")
		data, err := os.ReadFile(envPath)
		if err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					os.Setenv(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
				}
			}
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return
		}
		dir = parent
	}
}

func buildDSN() string {
	host := os.Getenv("DB_HOST")
	if host == "db" || host == "" {
		host = "localhost"
	}
	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"), sslmode)
}

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		os.Exit(m.Run())
	}

	dsn := buildDSN()
	client, err := generated.Open("postgres", dsn)
	if err == nil {
		defer client.Close()
		ctx := context.Background()
		if err := client.Schema.Create(ctx); err == nil {
			client.UserSchema.Delete().ExecX(ctx)
			globalClient = client
		}
	}
	os.Exit(m.Run())
}

func setupAuthIntegration(t *testing.T) (*usersinfra.UserRepository, authdomain.IJwtAdapter, authdomain.IPasswordAdapter, *usersinfra.IDGen) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	if globalClient == nil {
		t.Fatalf("postgres not available — create .env.test at project root with DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, DB_SSLMODE")
	}

	globalClient.UserSchema.Delete().ExecX(context.Background())
	t.Cleanup(func() { globalClient.UserSchema.Delete().ExecX(context.Background()) })

	cfg := configinfra.NewConfigAdapter()
	return usersinfra.NewUserRepository(globalClient),
		authinfra.NewJwtAdapter(cfg),
		authinfra.NewPasswordAdapter(),
		usersinfra.NewIDGenerator()
}

func TestLogin_Integration(t *testing.T) {
	userRepo, jwtAdapter, passwordAdapter, idGen := setupAuthIntegration(t)

	email, _ := shareddomain.NewEmail("alice@test.com")
	hashed, _ := passwordAdapter.Hash("password123")
	_, err := userRepo.Create(context.Background(), &usersdomain.User{
		ID:           shareddomain.IdFromStr(idGen.Generate()),
		Name:         "Alice",
		Email:        email,
		PasswordHash: hashed,
		Role:         "client",
	})
	require.NoError(t, err)

	uc := authapp.NewLogin(userRepo, passwordAdapter, jwtAdapter)
	result, err := uc.Execute(context.Background(), authapp.LoginInput{
		Email:    "alice@test.com",
		Password: "password123",
	})

	require.NoError(t, err)
	assert.Equal(t, "Alice", result.User.Name)
	assert.NotEmpty(t, result.Tokens.AccessToken)
	assert.NotEmpty(t, result.Tokens.RefreshToken)
}

func TestLogin_WrongPassword_Integration(t *testing.T) {
	userRepo, jwtAdapter, passwordAdapter, idGen := setupAuthIntegration(t)

	email, _ := shareddomain.NewEmail("alice@test.com")
	hashed, _ := passwordAdapter.Hash("password123")
	_, err := userRepo.Create(context.Background(), &usersdomain.User{
		ID:           shareddomain.IdFromStr(idGen.Generate()),
		Name:         "Alice",
		Email:        email,
		PasswordHash: hashed,
		Role:         "client",
	})
	require.NoError(t, err)

	uc := authapp.NewLogin(userRepo, passwordAdapter, jwtAdapter)
	_, err = uc.Execute(context.Background(), authapp.LoginInput{
		Email:    "alice@test.com",
		Password: "wrongpassword",
	})

	assert.Error(t, err)
	var invalidErr *authdomain.InvalidCredentialsError
	assert.ErrorAs(t, err, &invalidErr)
}

func TestRegister_Integration(t *testing.T) {
	userRepo, jwtAdapter, passwordAdapter, idGen := setupAuthIntegration(t)

	uc := authapp.NewRegister(userRepo, passwordAdapter, idGen, jwtAdapter)
	result, err := uc.Execute(context.Background(), authapp.RegisterInput{
		Name:     "Bob",
		Email:    "bob@test.com",
		Password: "password123",
	})

	require.NoError(t, err)
	assert.Equal(t, "Bob", result.User.Name)
	assert.NotEmpty(t, result.Tokens.AccessToken)
	assert.NotEmpty(t, result.Tokens.RefreshToken)
}

func TestRegister_DuplicateEmail_Integration(t *testing.T) {
	userRepo, jwtAdapter, passwordAdapter, idGen := setupAuthIntegration(t)

	uc := authapp.NewRegister(userRepo, passwordAdapter, idGen, jwtAdapter)
	_, err := uc.Execute(context.Background(), authapp.RegisterInput{
		Name:     "Bob",
		Email:    "bob@test.com",
		Password: "password123",
	})
	require.NoError(t, err)

	_, err = uc.Execute(context.Background(), authapp.RegisterInput{
		Name:     "Bob Again",
		Email:    "bob@test.com",
		Password: "password456",
	})
	assert.Error(t, err)
	var dupErr *usersdomain.UserEmailAlreadyExistsError
	assert.ErrorAs(t, err, &dupErr)
}

func TestRefreshToken_Integration(t *testing.T) {
	userRepo, jwtAdapter, passwordAdapter, idGen := setupAuthIntegration(t)

	email, _ := shareddomain.NewEmail("alice@test.com")
	hashed, _ := passwordAdapter.Hash("password123")
	user, err := userRepo.Create(context.Background(), &usersdomain.User{
		ID:           shareddomain.IdFromStr(idGen.Generate()),
		Name:         "Alice",
		Email:        email,
		PasswordHash: hashed,
		Role:         "client",
	})
	require.NoError(t, err)

	payload := authdomain.TokenPayload{
		Sub:   user.ID.String(),
		Email: user.Email.String(),
		Role:  user.Role,
	}
	refreshToken, err := jwtAdapter.SignRefresh(payload)
	require.NoError(t, err)

	uc := authapp.NewRefreshToken(userRepo, jwtAdapter)
	result, err := uc.Execute(context.Background(), authapp.RefreshTokenInput{
		RefreshToken: refreshToken,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
}

func TestRefreshToken_InvalidToken_Integration(t *testing.T) {
	userRepo, jwtAdapter, _, _ := setupAuthIntegration(t)

	uc := authapp.NewRefreshToken(userRepo, jwtAdapter)
	_, err := uc.Execute(context.Background(), authapp.RefreshTokenInput{
		RefreshToken: "invalid-token",
	})

	assert.Error(t, err)
	var refreshErr *authdomain.RefreshTokenInvalidError
	assert.ErrorAs(t, err, &refreshErr)
}
