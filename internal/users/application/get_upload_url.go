package application

import (
	"context"
	"fmt"
	"time"

	"go-starter/internal/users/domain"
)

type IDGenerator interface {
	Generate() string
}

type GetProfileImageUploadUrlInput struct {
	UserID string
}

type GetProfileImageUploadUrlOutput struct {
	UploadURL string `json:"upload_url"`
	Key       string `json:"key"`
}

type GetProfileImageUploadUrl struct {
	userRepo domain.IUserRepository
	storage  UserStorageAdapter
	idGen    IDGenerator
}

func NewGetProfileImageUploadUrl(
	userRepo domain.IUserRepository,
	storage UserStorageAdapter,
	idGen IDGenerator,
) *GetProfileImageUploadUrl {
	return &GetProfileImageUploadUrl{
		userRepo: userRepo,
		storage:  storage,
		idGen:    idGen,
	}
}

func (uc *GetProfileImageUploadUrl) Execute(ctx context.Context, input GetProfileImageUploadUrlInput) (*GetProfileImageUploadUrlOutput, error) {
	user, err := uc.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, &domain.UserNotFoundError{ID: input.UserID}
	}

	key := fmt.Sprintf("images/%s/profile", input.UserID)

	uploadURL, err := uc.storage.GeneratePresignedUploadUrl(ctx, key, 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("generate upload url: %w", err)
	}

	return &GetProfileImageUploadUrlOutput{
		UploadURL: uploadURL,
		Key:       key,
	}, nil
}
