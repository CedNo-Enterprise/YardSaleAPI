package sale

import "GarageSaleAPI/domain/address"

func CreateSale(id string, name string, address address.Address) Sale {
	return Sale{
		id:      id,
		name:    name,
		address: address,
	}
}
