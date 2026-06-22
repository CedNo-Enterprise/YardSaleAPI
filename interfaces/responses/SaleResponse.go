package responses

import "GarageSaleAPI/domain/sale"

type SaleResponse struct {
	Name    string          `json:"name"`
	Address AddressResponse `json:"address"`
}

func NewSaleResponse(sale sale.Sale) *SaleResponse {
	return &SaleResponse{
		sale.Name(),
		*NewAddressResponse(sale.Address()),
	}
}
