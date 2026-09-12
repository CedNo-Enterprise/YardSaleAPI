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

func (s *Sale) Id() string {
	return s.id
}

func (s *Sale) SellerId() string {
	return s.sellerId
}

func (s *Sale) Name() string {
	return s.name
}

func (s *Sale) Address() address.Address {
	return s.address
}

func (s *Sale) Date() time.Time {
	return s.date
}

func (s *Sale) Description() string {
	return s.description
}

func (s *Sale) Items() []SaleItem {
	return s.items
}

func (s *Sale) Status() Status {
	return s.status
}

func (s *Sale) CreatedAt() time.Time {
	return s.createdAt
}

type Status string

const (
	StatusScheduled Status = "scheduled"
	StatusActive    Status = "active"
	StatusCompleted Status = "completed"
	StatusCancelled Status = "cancelled"
)

type SaleItem struct {
	id              int64
	saleId          string
	inventoryItemId int64
	name            string
	price           float64
	status          SaleItemStatus
}

func (s *SaleItem) Id() int64 {
	return s.id
}

func (s *SaleItem) SaleId() string {
	return s.saleId
}

func (s *SaleItem) InventoryItemId() int64 {
	return s.inventoryItemId
}

func (s *SaleItem) Name() string {
	return s.name
}

func (s *SaleItem) Price() float64 {
	return s.price
}

func (s *SaleItem) Status() SaleItemStatus {
	return s.status
}

type SaleItemStatus string

const (
	SaleItemStatusAvailable SaleItemStatus = "available"
	SaleItemStatusSold      SaleItemStatus = "sold"
	SaleItemStatusReserved  SaleItemStatus = "reserved"
)
