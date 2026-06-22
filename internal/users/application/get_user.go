package application

import (
	"context"

	"go-starter/internal/users/domain"
)

type UserOutput struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Email string  `json:"email"`
	Phone *string `json:"phone"`
	Role  string  `json:"role"`
	Banned bool   `json:"banned"`
}

type GetUser struct {
	userRepo domain.IUserRepository
}

func NewGetUser(userRepo domain.IUserRepository) *GetUser {
	return &GetUser{userRepo: userRepo}
}

func (uc *GetUser) Execute(ctx context.Context, id string) (*UserOutput, error) {
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, &domain.UserNotFoundError{ID: id}
	}
	return userToOutput(user), nil
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
