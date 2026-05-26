package user

import (
	"net/mail"
	"time"
)

func CreateUser(id int64, username string, password string, email mail.Address, createdTime time.Time) User {
	return User{
		Id:        id,
		Username:  username,
		Password:  password,
		Email:     email,
		CreatedAt: createdTime,
		UpdatedAt: createdTime,
	}
}
