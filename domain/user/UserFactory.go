package user

import (
	"net/mail"
	"time"
)

func CreateUser(username string, password string, email mail.Address, createdTime time.Time) User {
	return User{
		Username:  username,
		Password:  password,
		Email:     email,
		CreatedAt: createdTime,
		UpdatedAt: createdTime,
	}
}
