package domain

import "context"

type IUserRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, filter UserListFilter) ([]*User, int, error)
	Update(ctx context.Context, user *User) (*User, error)
	Delete(ctx context.Context, id string) error
	Ban(ctx context.Context, id string) (*User, error)
	Unban(ctx context.Context, id string) (*User, error)
}

type UserListFilter struct {
	Search string
	Page   int
	Limit  int
	Role   string
}
