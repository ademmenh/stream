package infrastructure

import (
	"github.com/google/uuid"

	"go-starter/internal/shared/domain"
	"go-starter/internal/shared/infrastructure/ent/generated"
	usersdomain "go-starter/internal/users/domain"
)

func toDomain(u *generated.UserSchema) *usersdomain.User {
	var phone *domain.Phone
	if u.Phone != nil && *u.Phone != "" {
		p, _ := domain.NewPhone(*u.Phone)
		phone = &p
	}
	email, _ := domain.NewEmail(u.Email)
	return &usersdomain.User{
		ID:           domain.IdFromStr(u.ID.String()),
		Name:         u.Name,
		Email:        email,
		Phone:        phone,
		PasswordHash: u.PasswordHash,
		Role:         u.Role,
		Banned:       u.Banned,
		ProfileImage: u.ProfileImage,
	}
}

type IDGen struct{}

func NewIDGenerator() *IDGen {
	return &IDGen{}
}

func (g *IDGen) Generate() string {
	return uuid.New().String()
}
