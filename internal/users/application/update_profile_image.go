package application

import (
	"context"

	"go-starter/internal/users/domain"
)

type UpdateProfileImageInput struct {
	UserID string
	Key    string
}

type UpdateProfileImage struct {
	userRepo domain.IUserRepository
}

func NewUpdateProfileImage(userRepo domain.IUserRepository) *UpdateProfileImage {
	return &UpdateProfileImage{userRepo: userRepo}
}

func (uc *UpdateProfileImage) Execute(ctx context.Context, input UpdateProfileImageInput) (*UserOutput, error) {
	user, err := uc.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, &domain.UserNotFoundError{ID: input.UserID}
	}

	user.ProfileImage = &input.Key

	updated, err := uc.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, &domain.UserNotFoundError{ID: input.UserID}
	}

	return userToOutput(updated), nil
}
