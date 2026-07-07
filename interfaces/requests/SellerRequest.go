package requests

type SellerRequest struct {
	Username string `json:"username" validate:"required"`
}
