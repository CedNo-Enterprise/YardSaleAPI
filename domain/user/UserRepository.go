package user

import "context"

type UserRepository interface {
	AddUser(ctx context.Context, user User) error
	GetUserByUsername(ctx context.Context, username string) (*User, error)
}
