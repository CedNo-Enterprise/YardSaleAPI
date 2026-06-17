package user

import (
	"time"
)

type User struct {
	Username  string
	Password  string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
