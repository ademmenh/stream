package application

import (
	"context"

	"go-starter/internal/users/domain"
)

type PasswordAdapter interface {
	Hash(plain string) (string, error)
	Compare(plain, hashed string) bool
}

type ResetPasswordInput struct {
	UserID          string
	OldPassword     string
	NewPassword     string
	ConfirmPassword string
}

type ResetPassword struct {
	userRepo        domain.IUserRepository
	passwordAdapter PasswordAdapter
}

func NewResetPassword(userRepo domain.IUserRepository, passwordAdapter PasswordAdapter) *ResetPassword {
	return &ResetPassword{
		userRepo:        userRepo,
		passwordAdapter: passwordAdapter,
	}
}

func (uc *ResetPassword) Execute(ctx context.Context, input ResetPasswordInput) error {
	user, err := uc.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return &domain.UserNotFoundError{ID: input.UserID}
	}

	if !uc.passwordAdapter.Compare(input.OldPassword, user.PasswordHash) {
		return &domain.InvalidOldPasswordError{}
	}

	if input.NewPassword != input.ConfirmPassword {
		return &domain.PasswordMismatchError{}
	}

	hash, err := uc.passwordAdapter.Hash(input.NewPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hash

	_, err = uc.userRepo.Update(ctx, user)
	if err != nil {
		return err
	}

	return nil
}
