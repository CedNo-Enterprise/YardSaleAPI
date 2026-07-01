package seller

import (
	"time"
)

func CreateSeller(id string, userId string, createdTime time.Time) Seller {
	return Seller{
		id:             id,
		userId:         userId,
		savedAddresses: []SavedAddress{},
		createdAt:      createdTime,
	}
}
