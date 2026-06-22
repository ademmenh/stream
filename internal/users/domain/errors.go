package domain

type UserError interface {
	error
	isUserError()
}

type UserEmailAlreadyExistsError struct {
	Email string
}

func (e *UserEmailAlreadyExistsError) Error() string {
	return "User with email already exists: " + e.Email
}
func (e *UserEmailAlreadyExistsError) isUserError() {}

type UserNotFoundError struct {
	ID string
}

func (e *UserNotFoundError) Error() string { return "User not found: " + e.ID }
func (e *UserNotFoundError) isUserError()  {}
