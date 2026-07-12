package application

import (
	"context"

	"go-starter/internal/auth/domain"
	usersdomain "go-starter/internal/users/domain"
)

type LoginInput struct {
	Email    string
	Password string
}

type TokensOutput struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type LoginUserOutput struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type LoginOutput struct {
	User   LoginUserOutput
	Tokens TokensOutput
}

type PasswordAdapter = domain.IPasswordAdapter
type JwtAdapter = domain.IJwtAdapter
type TokenPayload = domain.TokenPayload
type IDGenerator interface {
	Generate() string
}

type AuthUserRepo interface {
	FindByEmail(ctx context.Context, email string) (*usersdomain.User, error)
	FindByID(ctx context.Context, id string) (*usersdomain.User, error)
	Create(ctx context.Context, user *usersdomain.User) (*usersdomain.User, error)
}

func NewLogin(userRepo AuthUserRepo, passwordAdapter PasswordAdapter, jwtAdapter JwtAdapter) *Login {
	return &Login{
		userRepo:        userRepo,
		passwordAdapter: passwordAdapter,
		jwtAdapter:      jwtAdapter,
	}
}

type Login struct {
	userRepo        AuthUserRepo
	passwordAdapter PasswordAdapter
	jwtAdapter      JwtAdapter
}

func userToAuthOutput(user *usersdomain.User) LoginUserOutput {
	return LoginUserOutput{
		Id:    user.GetID(),
		Name:  user.GetName(),
		Email: user.GetEmail(),
		Role:  user.GetRole(),
	}
}

func (l *Login) Execute(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	user, err := l.userRepo.FindByEmail(ctx, input.Email)
	if err != nil || user == nil {
		return nil, &domain.InvalidCredentialsError{}
	}

	if !l.passwordAdapter.Compare(input.Password, user.PasswordHash) {
		return nil, &domain.InvalidCredentialsError{}
	}

	if user.GetBanned() {
		return nil, &domain.UserBannedError{}
	}

	payload := TokenPayload{
		Sub:   user.GetID(),
		Email: user.GetEmail(),
		Role:  user.Role,
	}

	accessToken, err := l.jwtAdapter.Sign(payload)
	if err != nil {
		return nil, err
	}
	refreshToken, err := l.jwtAdapter.SignRefresh(payload)
	if err != nil {
		return nil, err
	}

	return &LoginOutput{
		User: userToAuthOutput(user),
		Tokens: TokensOutput{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}, nil
}


