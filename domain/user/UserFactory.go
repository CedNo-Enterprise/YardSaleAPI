package user

import (
	"time"
)

func CreateUser(username string, password string, email string, createdTime time.Time) User {
	return User{
		Username:  username,
		Password:  password,
		Email:     email,
		CreatedAt: createdTime,
		UpdatedAt: createdTime,
	}
}
