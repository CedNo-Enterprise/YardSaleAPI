package sale

import (
	"GarageSaleAPI/domain/address"
	"time"
)

func CreateSale(id string, sellerId string, name string, address address.Address, date time.Time, description string, creationTime time.Time) Sale {
	status := StatusScheduled
	if date.Before(time.Now()) {
		status = StatusActive
	}

	return Sale{
		id:          id,
		sellerId:    id,
		name:        name,
		address:     address,
		date:        date,
		description: description,
		items:       []SaleItem{},
		status:      status,
		createdAt:   creationTime,
	}
}
