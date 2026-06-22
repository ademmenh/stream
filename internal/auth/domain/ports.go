package domain

import "github.com/golang-jwt/jwt/v5"

type TokenPayload struct {
	Sub   string
	Email string
	Role  string
}

type JWTClaims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
	Role  string `json:"role"`
}

type IJwtAdapter interface {
	Sign(payload TokenPayload) (string, error)
	Verify(token string) (*TokenPayload, error)
	SignRefresh(payload TokenPayload) (string, error)
	VerifyRefresh(token string) (*TokenPayload, error)
}

type IPasswordAdapter interface {
	Hash(plain string) (string, error)
	Compare(plain, hashed string) bool
}
