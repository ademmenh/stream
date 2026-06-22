package userstests

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

	shareddomain "go-starter/internal/shared/domain"
	"go-starter/internal/shared/infrastructure/ent/generated"
	usersapp "go-starter/internal/users/application"
	usersdomain "go-starter/internal/users/domain"
	"go-starter/internal/users/infrastructure"
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

type integrationPWAdapter struct{}

func (m *integrationPWAdapter) Hash(plain string) (string, error) {
	return "hashed-" + plain, nil
}

func setupIntegrationRepo(t *testing.T) *infrastructure.UserRepository {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	if globalClient == nil {
		t.Fatal("postgres not available — create .env.test at project root with DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, DB_SSLMODE")
	}

	globalClient.UserSchema.Delete().ExecX(context.Background())
	t.Cleanup(func() { globalClient.UserSchema.Delete().ExecX(context.Background()) })

	return infrastructure.NewUserRepository(globalClient)
}

func newUserForIntegration(name, email string) *usersdomain.User {
	e, _ := shareddomain.NewEmail(email)
	return &usersdomain.User{
		ID:           shareddomain.NewId(),
		Name:         name,
		Email:        e,
		PasswordHash: "hashed-password",
		Role:         "client",
	}
}

func TestGetUser_Integration(t *testing.T) {
	repo := setupIntegrationRepo(t)
	user := newUserForIntegration("Alice", "alice@test.com")
	created, err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	uc := usersapp.NewGetUser(repo)
	result, err := uc.Execute(context.Background(), created.ID.String())

	require.NoError(t, err)
	assert.Equal(t, "Alice", result.Name)
	assert.Equal(t, "alice@test.com", result.Email)
}

func TestGetUser_NotFound_Integration(t *testing.T) {
	repo := setupIntegrationRepo(t)
	uc := usersapp.NewGetUser(repo)
	_, err := uc.Execute(context.Background(), "00000000-0000-0000-0000-000000000000")
	assert.Error(t, err)
	assert.ErrorContains(t, err, "not found")
}

func TestListUsers_Integration(t *testing.T) {
	repo := setupIntegrationRepo(t)
	repo.Create(context.Background(), newUserForIntegration("Bob", "bob@test.com"))
	repo.Create(context.Background(), newUserForIntegration("Alice", "alice@test.com"))

	uc := usersapp.NewListUsers(repo)
	result, err := uc.Execute(context.Background(), usersapp.ListUsersInput{Page: 1, Limit: 20})

	require.NoError(t, err)
	assert.Len(t, result.Users, 2)
	assert.Equal(t, 2, result.Total)
}

func TestListUsers_Search_Integration(t *testing.T) {
	repo := setupIntegrationRepo(t)
	repo.Create(context.Background(), newUserForIntegration("Alice", "alice@test.com"))
	repo.Create(context.Background(), newUserForIntegration("Bob", "bob@test.com"))

	uc := usersapp.NewListUsers(repo)
	result, err := uc.Execute(context.Background(), usersapp.ListUsersInput{Search: "Alice", Page: 1, Limit: 20})

	require.NoError(t, err)
	assert.Len(t, result.Users, 1)
	assert.Equal(t, "Alice", result.Users[0].Name)
}

func TestUpdateUser_Integration(t *testing.T) {
	repo := setupIntegrationRepo(t)
	created, err := repo.Create(context.Background(), newUserForIntegration("Alice", "alice@test.com"))
	require.NoError(t, err)

	newName := "Alice Updated"
	uc := usersapp.NewUpdateUser(repo, &integrationPWAdapter{})
	result, err := uc.Execute(context.Background(), usersapp.UpdateUserInput{
		ID:   created.ID.String(),
		Name: &newName,
	})

	require.NoError(t, err)
	assert.Equal(t, newName, result.Name)
}

func TestDeleteUser_Integration(t *testing.T) {
	repo := setupIntegrationRepo(t)
	created, err := repo.Create(context.Background(), newUserForIntegration("Alice", "alice@test.com"))
	require.NoError(t, err)

	uc := usersapp.NewDeleteUser(repo)
	err = uc.Execute(context.Background(), created.ID.String())
	assert.NoError(t, err)

	found, err := repo.FindByID(context.Background(), created.ID.String())
	assert.NoError(t, err)
	assert.Nil(t, found)
}

func TestBanUnbanUser_Integration(t *testing.T) {
	repo := setupIntegrationRepo(t)
	created, err := repo.Create(context.Background(), newUserForIntegration("Alice", "alice@test.com"))
	require.NoError(t, err)

	id := created.ID.String()

	banUC := usersapp.NewBanUser(repo)
	banned, err := banUC.Execute(context.Background(), id)
	require.NoError(t, err)
	assert.True(t, banned.Banned)

	unbanUC := usersapp.NewUnbanUser(repo)
	unbanned, err := unbanUC.Execute(context.Background(), id)
	require.NoError(t, err)
	assert.False(t, unbanned.Banned)
}
