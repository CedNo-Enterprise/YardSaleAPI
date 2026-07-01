package seller

import (
	"GarageSaleAPI/domain/address"
	"time"
)

type Seller struct {
	id             string
	userId         string
	savedAddresses []SavedAddress
	createdAt      time.Time
}

type SavedAddress struct {
	id        string
	label     string
	address   address.Address
	isDefault bool
}

func (s *Seller) Id() string {
	return s.id
}

func (s *Seller) UserId() string {
	return s.id
}
