package sale

import (
	"GarageSaleAPI/domain/address"
	"time"
)

type Sale struct {
	id          string
	sellerId    string
	name        string
	address     address.Address
	date        time.Time
	description string
	items       []SaleItem
	status      Status
	createdAt   time.Time
}

type Status string

const (
	StatusScheduled Status = "scheduled"
	StatusActive    Status = "active"
	StatusCompleted Status = "completed"
	StatusCancelled Status = "cancelled"
)

type SaleItem struct {
	InventoryItemID string
	Name            string
	Price           float64
	Status          SaleItemStatus
}

type SaleItemStatus string

const (
	SaleItemStatusAvailable SaleItemStatus = "available"
	SaleItemStatusSold      SaleItemStatus = "sold"
	SaleItemStatusReserved  SaleItemStatus = "reserved"
)

func (s *Sale) Id() string {
	return s.id
}

func (s *Sale) Name() string {
	return s.name
}

func (s *Sale) Address() address.Address {
	return s.address
}
