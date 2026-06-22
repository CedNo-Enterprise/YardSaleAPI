package responses

import (
	"GarageSaleAPI/domain/address"
)

type AddressResponse struct {
	Line1      string  `json:"line1"`
	Line2      string  `json:"line2,omitempty"`
	City       string  `json:"city"`
	State      string  `json:"state"`
	PostalCode string  `json:"postal_code"`
	Country    string  `json:"country"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
}

func NewAddressResponse(address address.Address) *AddressResponse {
	return &AddressResponse{
		Line1:      address.Line1(),
		Line2:      address.Line2(),
		City:       address.City(),
		State:      address.State(),
		PostalCode: address.PostalCode(),
		Country:    address.Country(),
		Latitude:   address.Latitude(),
		Longitude:  address.Longitude(),
	}
}
