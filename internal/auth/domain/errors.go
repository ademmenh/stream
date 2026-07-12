package domain

type AuthError interface {
	error
	isAuthError()
}

type InvalidCredentialsError struct{}

func (e *InvalidCredentialsError) Error() string  { return "Invalid email or password" }
func (e *InvalidCredentialsError) isAuthError()   {}

type RefreshTokenInvalidError struct{}

func (e *RefreshTokenInvalidError) Error() string { return "Invalid or expired refresh token" }
func (e *RefreshTokenInvalidError) isAuthError()  {}

type UserBannedError struct{}

func (e *UserBannedError) Error() string { return "User is banned" }
func (e *UserBannedError) isAuthError()  {}
