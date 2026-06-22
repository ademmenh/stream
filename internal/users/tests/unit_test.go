package userstests

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	shareddomain "go-starter/internal/shared/domain"
	usersapp "go-starter/internal/users/application"
	usersdomain "go-starter/internal/users/domain"
	"go-starter/internal/users/infrastructure"
)

type mockPasswordAdapter struct{}

func (m *mockPasswordAdapter) Hash(plain string) (string, error) {
	return "hashed-" + plain, nil
}

func newTestUser(id, name, email string) *usersdomain.User {
	e, _ := shareddomain.NewEmail(email)
	return &usersdomain.User{
		ID:           shareddomain.IdFromStr(id),
		Name:         name,
		Email:        e,
		PasswordHash: "hashed-password",
		Role:         "client",
	}
}

func seedUser(t *testing.T, repo *infrastructure.InMemoryUserRepository, u *usersdomain.User) {
	t.Helper()
	_, err := repo.Create(context.Background(), u)
	require.NoError(t, err)
}

func TestGetUser_Success(t *testing.T) {
	repo := infrastructure.NewInMemoryUserRepository()
	user := newTestUser("u1", "Alice", "alice@test.com")
	seedUser(t, repo, user)

	uc := usersapp.NewGetUser(repo)
	result, err := uc.Execute(context.Background(), "u1")

	require.NoError(t, err)
	assert.Equal(t, "u1", result.ID)
	assert.Equal(t, "Alice", result.Name)
	assert.Equal(t, "alice@test.com", result.Email)
	assert.Equal(t, "client", result.Role)
}

func TestGetUser_NotFound(t *testing.T) {
	repo := infrastructure.NewInMemoryUserRepository()
	uc := usersapp.NewGetUser(repo)

	_, err := uc.Execute(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestListUsers_Empty(t *testing.T) {
	repo := infrastructure.NewInMemoryUserRepository()
	uc := usersapp.NewListUsers(repo)

	result, err := uc.Execute(context.Background(), usersapp.ListUsersInput{Page: 1, Limit: 20})

	require.NoError(t, err)
	assert.Empty(t, result.Users)
	assert.Equal(t, 0, result.Total)
}

func TestListUsers_WithResults(t *testing.T) {
	repo := infrastructure.NewInMemoryUserRepository()
	seedUser(t, repo, newTestUser("u1", "Bob", "bob@test.com"))
	seedUser(t, repo, newTestUser("u2", "Alice", "alice@test.com"))

	uc := usersapp.NewListUsers(repo)
	result, err := uc.Execute(context.Background(), usersapp.ListUsersInput{Page: 1, Limit: 20})

	require.NoError(t, err)
	assert.Len(t, result.Users, 2)
	assert.Equal(t, 2, result.Total)
	assert.Equal(t, "Alice", result.Users[0].Name)
	assert.Equal(t, "Bob", result.Users[1].Name)
}

func TestListUsers_Search(t *testing.T) {
	repo := infrastructure.NewInMemoryUserRepository()
	seedUser(t, repo, newTestUser("u1", "Alice", "alice@test.com"))
	seedUser(t, repo, newTestUser("u2", "Bob", "bob@test.com"))
	seedUser(t, repo, newTestUser("u3", "Charlie", "charlie@test.com"))

	uc := usersapp.NewListUsers(repo)
	result, err := uc.Execute(context.Background(), usersapp.ListUsersInput{Search: "alice", Page: 1, Limit: 20})

	require.NoError(t, err)
	assert.Len(t, result.Users, 1)
	assert.Equal(t, "Alice", result.Users[0].Name)
}

func TestListUsers_Pagination(t *testing.T) {
	repo := infrastructure.NewInMemoryUserRepository()
	for i := 0; i < 5; i++ {
		id := shareddomain.NewId()
		seedUser(t, repo, newTestUser(id.String(), fmt.Sprintf("User%d", i), fmt.Sprintf("user%d@test.com", i)))
	}

	uc := usersapp.NewListUsers(repo)
	result, err := uc.Execute(context.Background(), usersapp.ListUsersInput{Page: 1, Limit: 2})

	require.NoError(t, err)
	assert.Len(t, result.Users, 2)
	assert.Equal(t, 5, result.Total)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 2, result.Limit)
}

func TestUpdateUser_Success(t *testing.T) {
	repo := infrastructure.NewInMemoryUserRepository()
	user := newTestUser("u1", "Alice", "alice@test.com")
	seedUser(t, repo, user)

	newName := "Alice Updated"
	uc := usersapp.NewUpdateUser(repo, &mockPasswordAdapter{})
	result, err := uc.Execute(context.Background(), usersapp.UpdateUserInput{
		ID:   "u1",
		Name: &newName,
	})

	require.NoError(t, err)
	assert.Equal(t, newName, result.Name)
}

func TestUpdateUser_NotFound(t *testing.T) {
	repo := infrastructure.NewInMemoryUserRepository()
	uc := usersapp.NewUpdateUser(repo, &mockPasswordAdapter{})
	name := "Nobody"
	_, err := uc.Execute(context.Background(), usersapp.UpdateUserInput{
		ID:   "nonexistent",
		Name: &name,
	})
	assert.Error(t, err)
}

func TestDeleteUser_Success(t *testing.T) {
	repo := infrastructure.NewInMemoryUserRepository()
	seedUser(t, repo, newTestUser("u1", "Alice", "alice@test.com"))

	uc := usersapp.NewDeleteUser(repo)
	err := uc.Execute(context.Background(), "u1")

	assert.NoError(t, err)
}

func TestDeleteUser_NotFound(t *testing.T) {
	repo := infrastructure.NewInMemoryUserRepository()
	uc := usersapp.NewDeleteUser(repo)

	err := uc.Execute(context.Background(), "nonexistent")
	assert.NoError(t, err)
}

func TestBanUser_Success(t *testing.T) {
	repo := infrastructure.NewInMemoryUserRepository()
	seedUser(t, repo, newTestUser("u1", "Alice", "alice@test.com"))

	uc := usersapp.NewBanUser(repo)
	result, err := uc.Execute(context.Background(), "u1")

	require.NoError(t, err)
	assert.True(t, result.Banned)
}

func TestUnbanUser_Success(t *testing.T) {
	repo := infrastructure.NewInMemoryUserRepository()
	user := newTestUser("u1", "Alice", "alice@test.com")
	user.Banned = true
	seedUser(t, repo, user)

	uc := usersapp.NewUnbanUser(repo)
	result, err := uc.Execute(context.Background(), "u1")

	require.NoError(t, err)
	assert.False(t, result.Banned)
}
