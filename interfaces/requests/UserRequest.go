package requests

type UserRequest struct {
	Username string `validate:"required,min=3,max=15,alphanum"`
	Password string `validate:"required,min=12,max=64"`
	Email    string `validate:"required,email"`
}
