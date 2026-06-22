package infrastructure

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go-starter/internal/auth/domain"
	config "go-starter/internal/config/domain"
)

type JwtAdapter struct {
	config config.IConfig
}

func NewJwtAdapter(cfg config.IConfig) domain.IJwtAdapter {
	return &JwtAdapter{config: cfg}
}

func (a *JwtAdapter) sign(payload domain.TokenPayload, secret string, expirySec int) (string, error) {
	now := time.Now().UTC()
	claims := domain.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   payload.Sub,
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expirySec) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		Email: payload.Email,
		Role:  payload.Role,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func (a *JwtAdapter) verify(tokenStr string, secret string) (*domain.TokenPayload, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &domain.JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*domain.JWTClaims)
	if !ok || !token.Valid {
		return nil, &domain.InvalidCredentialsError{}
	}
	return &domain.TokenPayload{
		Sub:   claims.Subject,
		Email: claims.Email,
		Role:  claims.Role,
	}, nil
}

func (a *JwtAdapter) Sign(payload domain.TokenPayload) (string, error) {
	return a.sign(payload, a.config.JWTAccessTokenSecret(), a.config.JWTAccessTokenExpiry())
}

func (a *JwtAdapter) Verify(token string) (*domain.TokenPayload, error) {
	return a.verify(token, a.config.JWTAccessTokenSecret())
}

func (a *JwtAdapter) SignRefresh(payload domain.TokenPayload) (string, error) {
	return a.sign(payload, a.config.JWTRefreshTokenSecret(), a.config.JWTRefreshTokenExpiry())
}

func (a *JwtAdapter) VerifyRefresh(token string) (*domain.TokenPayload, error) {
	return a.verify(token, a.config.JWTRefreshTokenSecret())
}
