package requests

import "time"

type SaleRequest struct {
	SellerId    string         `json:"sellerId" validate:"required"`
	Name        string         `json:"name"       validate:"required,max=64"`
	Address     AddressRequest `json:"address"    validate:"required"`
	Date        time.Time      `json:"date"       validate:"required"`
	Description string         `json:"description"`
}
