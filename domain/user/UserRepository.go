package user

import "context"

type UserRepository interface {
	Save(context.Context, User) error
	GetByUsername(context.Context, string) (*User, error)
}
