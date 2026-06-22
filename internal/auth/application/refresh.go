package application

import (
	"context"

	domain "go-starter/internal/auth/domain"
)

type RefreshTokenInput struct {
	RefreshToken string
}

type RefreshToken struct {
	userRepo   AuthUserRepo
	jwtAdapter JwtAdapter
}

func NewRefreshToken(userRepo AuthUserRepo, jwtAdapter JwtAdapter) *RefreshToken {
	return &RefreshToken{
		userRepo:   userRepo,
		jwtAdapter: jwtAdapter,
	}
}

func (r *RefreshToken) Execute(ctx context.Context, input RefreshTokenInput) (*TokensOutput, error) {
	payload, err := r.jwtAdapter.VerifyRefresh(input.RefreshToken)
	if err != nil {
		return nil, &domain.RefreshTokenInvalidError{}
	}

	user, err := r.userRepo.FindByID(ctx, payload.Sub)
	if err != nil || user == nil {
		return nil, &domain.RefreshTokenInvalidError{}
	}

	newPayload := TokenPayload{
		Sub:   user.GetID(),
		Email: user.GetEmail(),
		Role:  user.Role,
	}

	accessToken, err := r.jwtAdapter.Sign(newPayload)
	if err != nil {
		return nil, err
	}
	refreshToken, err := r.jwtAdapter.SignRefresh(newPayload)
	if err != nil {
		return nil, err
	}

	return &TokensOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}


