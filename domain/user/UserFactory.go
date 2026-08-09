package user

import (
	"time"
)

func CreateUser(id string, username string, password string, email string, createdTime time.Time) User {
	return User{
		id:        id,
		username:  username,
		password:  password,
		email:     email,
		createdAt: createdTime,
		updatedAt: createdTime,
	}
}
