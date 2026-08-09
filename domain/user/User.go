package user

import (
	"time"
)

type User struct {
	id        string
	username  string
	password  string
	email     string
	createdAt time.Time
	updatedAt time.Time
}

func (u User) Id() string {
	return u.id
}

func (u User) Username() string {
	return u.username
}

func (u User) Password() string {
	return u.password
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
