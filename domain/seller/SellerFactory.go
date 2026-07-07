package seller

import (
	"time"
)

func CreateSeller(id string, username string, createdTime time.Time) Seller {
	return Seller{
		id:             id,
		username:       username,
		savedAddresses: []SavedAddress{},
		inventory:      []InventoryItem{},
		createdAt:      createdTime,
	}
}
