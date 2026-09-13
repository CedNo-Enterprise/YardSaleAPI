package buyer

import (
	"GarageSaleAPI/domain/address"
	"time"
)

func CreateBuyer(id string, userId string, displayName string, createdTime time.Time) *Buyer {
	return &Buyer{
		id:          id,
		userId:      userId,
		displayName: displayName,
		homeAddress: nil,
		createdAt:   createdTime,
	}
}

func HydrateBuyer(
	id string, userId string, displayName string,
	homeAddress *address.Address, createdTime time.Time,
) *Buyer {
	return &Buyer{
		id:          id,
		userId:      userId,
		displayName: displayName,
		homeAddress: homeAddress,
		createdAt:   createdTime,
	}
}
