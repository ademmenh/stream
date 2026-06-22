package infrastructure

import (
	"go-starter/internal/auth/domain"
	"golang.org/x/crypto/bcrypt"
)

type PasswordAdapter struct{}

func NewPasswordAdapter() domain.IPasswordAdapter {
	return &PasswordAdapter{}
}

func (a *PasswordAdapter) Hash(plain string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (a *PasswordAdapter) Compare(plain, hashed string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}
