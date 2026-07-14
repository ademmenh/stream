package application

import (
	"context"

	shareddomain "go-starter/internal/shared/domain"
	"go-starter/internal/users/domain"
)

type UpdateCurrentUserInput struct {
	ID    string
	Name  *string
	Email *string
	Phone *string
}

type UpdateCurrentUser struct {
	userRepo domain.IUserRepository
}

func NewUpdateCurrentUser(userRepo domain.IUserRepository) *UpdateCurrentUser {
	return &UpdateCurrentUser{userRepo: userRepo}
}

func (uc *UpdateCurrentUser) Execute(ctx context.Context, input UpdateCurrentUserInput) (*UserOutput, error) {
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

	updated, err := uc.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, &domain.UserNotFoundError{ID: input.ID}
	}

	return userToOutput(updated), nil
}
