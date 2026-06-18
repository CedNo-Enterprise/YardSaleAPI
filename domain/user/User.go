package user

import (
	"time"
)

type User struct {
	username  string
	password  string
	email     string
	createdAt time.Time
	updatedAt time.Time
}

func (u User) Username() string {
	return u.username
}

func (u User) Email() string {
	return u.email
}

func (u User) CreatedAt() time.Time {
	return u.createdAt
}

func (u User) UpdatedAt() time.Time {
	return u.updatedAt
}
