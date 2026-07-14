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

type CannotBanAdminError struct{}

func (e *CannotBanAdminError) Error() string { return "Cannot ban an admin user" }
func (e *CannotBanAdminError) isUserError()  {}

type InvalidOldPasswordError struct{}

func (e *InvalidOldPasswordError) Error() string { return "Old password is incorrect" }
func (e *InvalidOldPasswordError) isUserError()  {}

type PasswordMismatchError struct{}

func (e *PasswordMismatchError) Error() string { return "New password and confirmation do not match" }
func (e *PasswordMismatchError) isUserError()  {}
