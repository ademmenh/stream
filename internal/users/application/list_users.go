package application

import (
	"context"

	"go-starter/internal/users/domain"
)

type ListUsersInput struct {
	Search string
	Page   int
	Limit  int
	Role   string
}

type ListUsersOutput struct {
	Users []UserOutput `json:"users"`
	Total int          `json:"total"`
	Page  int          `json:"page"`
	Limit int          `json:"limit"`
}

type ListUsers struct {
	userRepo domain.IUserRepository
}

func NewListUsers(userRepo domain.IUserRepository) *ListUsers {
	return &ListUsers{userRepo: userRepo}
}

func (uc *ListUsers) Execute(ctx context.Context, input ListUsersInput) (*ListUsersOutput, error) {
	filter := domain.UserListFilter{
		Search: input.Search,
		Page:   input.Page,
		Limit:  input.Limit,
		Role:   input.Role,
	}

	users, total, err := uc.userRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	output := make([]UserOutput, len(users))
	for i, u := range users {
		output[i] = *userToOutput(u)
	}

	return &ListUsersOutput{
		Users: output,
		Total: total,
		Page:  input.Page,
		Limit: input.Limit,
	}, nil
}
