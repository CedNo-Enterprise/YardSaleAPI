package requests

type UserRequest struct {
	Username string `json:"username"   validate:"required,min=3,max=15,alphanum"`
	Password string `json:"password"   validate:"required,min=12,max=64"`
	Email    string `json:"email"      validate:"required,email"`
}
