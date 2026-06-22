package requests

type SaleRequest struct {
	Name    string         `json:"name"       validate:"required,max=64"`
	Address AddressRequest `json:"address"    validate:"required"`
}
