package user

import "context"

type UserRepository interface {
	AddUser(context.Context, User) error
	GetUserByUsername(context.Context, string) (*User, error)
}
