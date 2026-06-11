package sale

func CreateSale(id string, name string, address string) Sale {
	return Sale{
		id:      id,
		name:    name,
		address: address,
	}
}
