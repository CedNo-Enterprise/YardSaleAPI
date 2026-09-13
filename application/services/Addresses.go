package services

import (
	"GarageSaleAPI/domain/address"
	"GarageSaleAPI/interfaces/requests"
)

func addressFromRequest(addressDTO requests.AddressRequest) address.Address {
	return address.CreateAddress(
		addressDTO.Line1, &addressDTO.Line2,
		addressDTO.City, addressDTO.State, addressDTO.PostalCode,
		addressDTO.Country,
	)
}
