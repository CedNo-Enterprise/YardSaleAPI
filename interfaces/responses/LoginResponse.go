package responses

import (
	"GarageSaleAPI/domain/user"
	"time"
)

type LoginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      UserResponse `json:"user"`
}

func NewLoginResponse(token *string, expiresAt *time.Time, u *user.User) LoginResponse {
	return LoginResponse{
		*token,
		*expiresAt,
		NewUserResponse(u),
	}
}
