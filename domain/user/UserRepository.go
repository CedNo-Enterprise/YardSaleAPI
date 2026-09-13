package user

import "context"

type UserRepository interface {
	Create(context.Context, *User) error
	GetByEmail(context.Context, string) (*User, error)
	GetById(context.Context, string) (*User, error)
}
