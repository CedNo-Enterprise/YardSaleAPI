package seller

import (
	"GarageSaleAPI/domain/address"
	"time"
)

type Seller struct {
	id             string
	userId         string
	savedAddresses []SavedAddress
	inventory      []InventoryItem
	createdAt      time.Time
}

func (s *Seller) Id() string {
	return s.id
}

func (s *Seller) UserId() string {
	return s.id
}

type SavedAddress struct {
	id        string
	label     string
	address   address.Address
	isDefault bool
}

type InventoryItem struct {
	id          string
	name        string
	description string
	price       float64
	status      ItemStatus
}

type ItemStatus string

const (
	ItemStatusAvailable ItemStatus = "available"
	ItemStatusListed    ItemStatus = "listed"
	ItemStatusSold      ItemStatus = "sold"
)
