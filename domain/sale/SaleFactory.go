package sale

import (
	"GarageSaleAPI/domain/address"
	"time"
)

func CreateSale(
	id string, sellerId string, name string, address address.Address,
	date time.Time, description string, creationTime time.Time,
) *Sale {
	status := StatusScheduled
	if date.Before(time.Now()) {
		status = StatusActive
	}

	return &Sale{
		id:          id,
		sellerId:    sellerId,
		name:        name,
		address:     address,
		date:        date,
		description: description,
		items:       []SaleItem{},
		status:      status,
		createdAt:   creationTime,
	}
}

func HydrateSale(
	id, sellerId, name string, addr address.Address, date time.Time,
	description string, items []SaleItem, status Status, createdAt time.Time,
) *Sale {
	return &Sale{
		id:          id,
		sellerId:    sellerId,
		name:        name,
		address:     addr,
		date:        date,
		description: description,
		items:       items,
		status:      status,
		createdAt:   createdAt,
	}
}

func HydrateSaleItem(
	id int64, saleId string, inventoryItemId int64, name string,
	price float64, status SaleItemStatus,
) *SaleItem {
	return &SaleItem{
		id:              id,
		saleId:          saleId,
		inventoryItemId: inventoryItemId,
		name:            name,
		price:           price,
		status:          status,
	}
}
