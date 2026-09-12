package user

import "context"

type UserRepository interface {
	Create(context.Context, *User) error
	GetByUsername(context.Context, string) (*User, error)
	GetByEmail(context.Context, string) (*User, error)
	GetById(context.Context, string) (*User, error)
}
