package application

import (
	"context"

	"go-starter/internal/users/domain"
)

type DeleteUser struct {
	userRepo domain.IUserRepository
}

func NewDeleteUser(userRepo domain.IUserRepository) *DeleteUser {
	return &DeleteUser{userRepo: userRepo}
}

func (uc *DeleteUser) Execute(ctx context.Context, id string) error {
	return uc.userRepo.Delete(ctx, id)
}
