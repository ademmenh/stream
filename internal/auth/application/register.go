package application

import (
	"context"

	shareddomain "go-starter/internal/shared/domain"
	usersdomain "go-starter/internal/users/domain"
)

type RegisterInput struct {
	Name     string
	Email    string
	Password string
	Phone    *string
}

type Register struct {
	userRepo        AuthUserRepo
	passwordAdapter PasswordAdapter
	idGenerator     IDGenerator
	jwtAdapter      JwtAdapter
}

func NewRegister(
	userRepo AuthUserRepo,
	passwordAdapter PasswordAdapter,
	idGenerator IDGenerator,
	jwtAdapter JwtAdapter,
) *Register {
	return &Register{
		userRepo:        userRepo,
		passwordAdapter: passwordAdapter,
		idGenerator:     idGenerator,
		jwtAdapter:      jwtAdapter,
	}
}

func (r *Register) Execute(ctx context.Context, input RegisterInput) (*LoginOutput, error) {
	existing, _ := r.userRepo.FindByEmail(ctx, input.Email)
	if existing != nil {
		return nil, &usersdomain.UserEmailAlreadyExistsError{Email: input.Email}
	}

	hashedPassword, err := r.passwordAdapter.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	email, _ := shareddomain.NewEmail(input.Email)
	user := &usersdomain.User{
		ID:           shareddomain.IdFromStr(r.idGenerator.Generate()),
		Name:         input.Name,
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         "client",
	}
	if input.Phone != nil {
		phone, _ := shareddomain.NewPhone(*input.Phone)
		user.Phone = &phone
	}

	savedUser, err := r.userRepo.Create(ctx, user)
	if err != nil {
		existing, lookupErr := r.userRepo.FindByEmail(ctx, input.Email)
		if lookupErr == nil && existing != nil {
			return nil, &usersdomain.UserEmailAlreadyExistsError{Email: input.Email}
		}
		return nil, err
	}

	payload := TokenPayload{
		Sub:   savedUser.GetID(),
		Email: savedUser.GetEmail(),
		Role:  savedUser.Role,
	}

	accessToken, err := r.jwtAdapter.Sign(payload)
	if err != nil {
		return nil, err
	}
	refreshToken, err := r.jwtAdapter.SignRefresh(payload)
	if err != nil {
		return nil, err
	}

	return &LoginOutput{
		User: userToAuthOutput(savedUser),
		Tokens: TokensOutput{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}, nil
}


