package services

import (
	"GarageSaleAPI/domain/address"
	"GarageSaleAPI/interfaces/requests"
)

// addressFromRequest builds a domain address from a request DTO. Coordinates are
// deliberately absent: nothing geocodes an address on the way in yet, so every
// address written through the API sits at (0, 0) until that lands.
func addressFromRequest(addressDTO requests.AddressRequest) address.Address {
	return address.CreateAddress(
		addressDTO.Line1, &addressDTO.Line2,
		addressDTO.City, addressDTO.State, addressDTO.PostalCode,
		addressDTO.Country,
	)
}
