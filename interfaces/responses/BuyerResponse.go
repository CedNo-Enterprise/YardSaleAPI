package responses

import (
	"GarageSaleAPI/domain/buyer"
)

type BuyerResponse struct {
	Id          string           `json:"id"`
	DisplayName string           `json:"display_name"`
	HomeAddress *AddressResponse `json:"home_address,omitempty"`
}

func NewBuyerResponse(b *buyer.Buyer) BuyerResponse {
	response := BuyerResponse{
		Id:          b.Id(),
		DisplayName: b.DisplayName(),
	}

	if home := b.HomeAddress(); home != nil {
		response.HomeAddress = NewAddressResponse(*home)
	}

	return response
}
