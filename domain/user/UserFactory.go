package user

import (
	"time"
)

func CreateUser(username string, password string, email string, createdTime time.Time) User {
	return User{
		username:  username,
		password:  password,
		email:     email,
		createdAt: createdTime,
		updatedAt: createdTime,
	}
}
