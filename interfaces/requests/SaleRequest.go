package requests

type SaleRequest struct {
	Name    string `validate:"required,max=64"`
	Address string `validate:"required"`
}
