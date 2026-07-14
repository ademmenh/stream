package userstests

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
	sharedtests "go-starter/internal/shared/tests"
	usersapp "go-starter/internal/users/application"
	usersdomain "go-starter/internal/users/domain"
	"go-starter/internal/users/infrastructure"
)

var globalClient *generated.Client

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

func (m *integrationPWAdapter) Compare(plain, hashed string) bool {
	return plain == hashed[len("hashed-"):]
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

	uc := usersapp.NewGetUser(repo, &mockStorageAdapter{})
	result, err := uc.Execute(context.Background(), created.ID.String())

	require.NoError(t, err)
	assert.Equal(t, "Alice", result.Name)
	assert.Equal(t, "alice@test.com", result.Email)
}

func TestGetUser_NotFound_Integration(t *testing.T) {
	repo := setupIntegrationRepo(t)
	uc := usersapp.NewGetUser(repo, &mockStorageAdapter{})
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
	uc := usersapp.NewUpdateUser(repo)
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
