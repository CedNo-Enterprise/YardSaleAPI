package user

import (
	"net/mail"
	"time"
)

type User struct {
	Id        int64
	Username  string
	Password  string
	Email     mail.Address
	CreatedAt time.Time
	UpdatedAt time.Time
}
