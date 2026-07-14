package application

import (
	"context"
	"time"

	"go-starter/internal/users/domain"
)

type UserStorageAdapter interface {
	GeneratePresignedGetUrl(ctx context.Context, key string, expiry time.Duration) (string, error)
	GeneratePresignedUploadUrl(ctx context.Context, key string, expiry time.Duration) (string, error)
}

type UserOutput struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Email        string  `json:"email"`
	Phone        *string `json:"phone"`
	Role         string  `json:"role"`
	Banned       bool    `json:"banned"`
	ProfileImage *string `json:"profile_image"`
}

type GetUser struct {
	userRepo domain.IUserRepository
	storage  UserStorageAdapter
}

func NewGetUser(userRepo domain.IUserRepository, storage UserStorageAdapter) *GetUser {
	return &GetUser{userRepo: userRepo, storage: storage}
}

func (uc *GetUser) Execute(ctx context.Context, id string) (*UserOutput, error) {
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, &domain.UserNotFoundError{ID: id}
	}
	return uc.userToOutput(ctx, user), nil
}

func (uc *GetUser) userToOutput(ctx context.Context, user *domain.User) *UserOutput {
	var phonePtr *string
	if user.GetPhone() != nil {
		p := user.GetPhone().String()
		phonePtr = &p
	}

	var profileImageURL *string
	if user.GetProfileImage() != nil {
		url, err := uc.storage.GeneratePresignedGetUrl(ctx, *user.GetProfileImage(), 1*time.Hour)
		if err == nil {
			profileImageURL = &url
		}
	}

	return &UserOutput{
		ID:           user.GetID(),
		Name:         user.GetName(),
		Email:        user.GetEmail(),
		Phone:        phonePtr,
		Role:         user.GetRole(),
		Banned:       user.GetBanned(),
		ProfileImage: profileImageURL,
	}
}

func userToOutput(user *domain.User) *UserOutput {
	var phonePtr *string
	if user.GetPhone() != nil {
		p := user.GetPhone().String()
		phonePtr = &p
	}
	return &UserOutput{
		ID:     user.GetID(),
		Name:   user.GetName(),
		Email:  user.GetEmail(),
		Phone:  phonePtr,
		Role:   user.GetRole(),
		Banned: user.GetBanned(),
	}
}
