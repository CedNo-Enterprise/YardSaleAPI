package seller

import (
	"GarageSaleAPI/domain/address"
	"time"
)

func CreateSeller(id string, userId string, createdTime time.Time) *Seller {
	return &Seller{
		id:             id,
		userId:         userId,
		savedAddresses: []SavedAddress{},
		inventory:      []InventoryItem{},
		createdAt:      createdTime,
	}
}

func HydrateSeller(
	id string, userId string, name string,
	savedAddresses []SavedAddress, inventoryItems []InventoryItem,
	createdTime time.Time,
) *Seller {
	return &Seller{
		id:             id,
		userId:         userId,
		name:           name,
		savedAddresses: savedAddresses,
		inventory:      inventoryItems,
		createdAt:      createdTime,
	}
}

func HydrateSavedAddress(id int64, sellerId string, label string, address address.Address, isDefault bool) *SavedAddress {
	return &SavedAddress{
		id:        id,
		sellerId:  sellerId,
		label:     label,
		address:   address,
		isDefault: isDefault,
	}
}

func HydrateInventoryItem(id int64, sellerId string, name string, description string, price float64, status ItemStatus) *InventoryItem {
	return &InventoryItem{
		id:          id,
		sellerId:    sellerId,
		name:        name,
		description: description,
		price:       price,
		status:      status,
	}
}
