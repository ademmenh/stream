package application

import (
	"context"

	"go-starter/internal/users/domain"
)

type UnbanUser struct {
	userRepo domain.IUserRepository
}

func NewUnbanUser(userRepo domain.IUserRepository) *UnbanUser {
	return &UnbanUser{userRepo: userRepo}
}

func (uc *UnbanUser) Execute(ctx context.Context, id string) (*UserOutput, error) {
	user, err := uc.userRepo.Unban(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, &domain.UserNotFoundError{ID: id}
	}
	return userToOutput(user), nil
}
