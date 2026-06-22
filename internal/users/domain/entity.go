package domain

import shareddomain "go-starter/internal/shared/domain"

type User struct {
	ID           shareddomain.Id
	Name         string
	Email        shareddomain.Email
	Phone        *shareddomain.Phone
	PasswordHash string
	Role         string
	Banned       bool
}

func NewUser(id shareddomain.Id, name string, email shareddomain.Email, passwordHash, role string, phone *shareddomain.Phone) *User {
	return &User{
		ID:           id,
		Name:         name,
		Email:        email,
		Phone:        phone,
		PasswordHash: passwordHash,
		Role:         role,
	}
}

func (e *User) GetID() string                 { return e.ID.String() }
func (e *User) GetName() string               { return e.Name }
func (e *User) GetEmail() string              { return e.Email.String() }
func (e *User) GetPhone() *shareddomain.Phone  { return e.Phone }
func (e *User) GetPasswordHash() string       { return e.PasswordHash }
func (e *User) GetRole() string               { return e.Role }
func (e *User) GetBanned() bool               { return e.Banned }

func (e *User) IsAdmin() bool  { return e.Role == "admin" }
func (e *User) IsClient() bool { return e.Role == "client" }
