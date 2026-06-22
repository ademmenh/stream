package application

import (
	"context"

	shareddomain "go-starter/internal/shared/domain"
	"go-starter/internal/users/domain"
)

type UpdateUserInput struct {
	ID       string
	Name     *string
	Email    *string
	Phone    *string
	Password *string
	NewRole  *string
}

type PasswordAdapter interface {
	Hash(plain string) (string, error)
}

type UpdateUser struct {
	userRepo        domain.IUserRepository
	passwordAdapter PasswordAdapter
}

func NewUpdateUser(userRepo domain.IUserRepository, passwordAdapter PasswordAdapter) *UpdateUser {
	return &UpdateUser{userRepo: userRepo, passwordAdapter: passwordAdapter}
}

func (uc *UpdateUser) Execute(ctx context.Context, input UpdateUserInput) (*UserOutput, error) {
	user, err := uc.userRepo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, &domain.UserNotFoundError{ID: input.ID}
	}

	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.Email != nil {
		email, err := shareddomain.NewEmail(*input.Email)
		if err != nil {
			return nil, err
		}
		user.Email = email
	}
	if input.Phone != nil {
		phone, err := shareddomain.NewPhone(*input.Phone)
		if err != nil {
			return nil, err
		}
		user.Phone = &phone
	}
	if input.Password != nil {
		hash, err := uc.passwordAdapter.Hash(*input.Password)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = hash
	}
	if input.NewRole != nil {
		user.Role = *input.NewRole
	}

	updated, err := uc.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, &domain.UserNotFoundError{ID: input.ID}
	}

	return userToOutput(updated), nil
}
