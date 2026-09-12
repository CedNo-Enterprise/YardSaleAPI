package seller

import (
	"GarageSaleAPI/domain/address"
	"time"
)

type Seller struct {
	id             string
	userId         string
	name           string
	savedAddresses []SavedAddress
	inventory      []InventoryItem
	createdAt      time.Time
}

func (s *Seller) Id() string {
	return s.id
}

func (s *Seller) UserId() string {
	return s.userId
}

func (s *Seller) Name() string {
	return s.name
}

func (s *Seller) SavedAddresses() []SavedAddress {
	return s.savedAddresses
}

func (s *Seller) Inventory() []InventoryItem {
	return s.inventory
}

func (s *Seller) CreatedAt() time.Time {
	return s.createdAt
}

type SavedAddress struct {
	id        int64
	sellerId  string
	label     string
	address   address.Address
	isDefault bool
}

func (s *SavedAddress) Id() int64 {
	return s.id
}

func (s *SavedAddress) SellerId() string {
	return s.sellerId
}

func (s *SavedAddress) Label() string {
	return s.label
}

func (s *SavedAddress) Address() address.Address {
	return s.address
}

func (s *SavedAddress) IsDefault() bool {
	return s.isDefault
}

type InventoryItem struct {
	id          int64
	sellerId    string
	name        string
	description string
	price       float64
	status      ItemStatus
}

func (i *InventoryItem) Id() int64 {
	return i.id
}

func (i *InventoryItem) SellerId() string {
	return i.sellerId
}

func (i *InventoryItem) Name() string {
	return i.name
}

func (i *InventoryItem) Description() string {
	return i.description
}

func (i *InventoryItem) Price() float64 {
	return i.price
}

func (i *InventoryItem) Status() ItemStatus {
	return i.status
}

type ItemStatus string

const (
	ItemStatusAvailable ItemStatus = "available"
	ItemStatusListed    ItemStatus = "listed"
	ItemStatusSold      ItemStatus = "sold"
)
