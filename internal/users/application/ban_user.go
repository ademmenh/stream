package application

import (
	"context"

	"go-starter/internal/users/domain"
)

type BanUser struct {
	userRepo domain.IUserRepository
}

func NewBanUser(userRepo domain.IUserRepository) *BanUser {
	return &BanUser{userRepo: userRepo}
}

func (uc *BanUser) Execute(ctx context.Context, id string) (*UserOutput, error) {
	user, err := uc.userRepo.Ban(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, &domain.UserNotFoundError{ID: id}
	}
	return userToOutput(user), nil
}
