package responses

import (
	"GarageSaleAPI/app/domain/user"
	"time"
)

type UserResponse struct {
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewUserResponse(u *user.User) UserResponse {
	return UserResponse{
		Username:  u.Username,
		Email:     u.Email.Address,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
