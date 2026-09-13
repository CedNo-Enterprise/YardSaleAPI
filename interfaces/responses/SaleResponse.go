package responses

import (
	"GarageSaleAPI/domain/sale"
	"time"
)

type SaleResponse struct {
	Id          string             `json:"id"`
	SellerId    string             `json:"seller_id"`
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Date        time.Time          `json:"date"`
	Status      sale.Status        `json:"status"`
	Address     AddressResponse    `json:"address"`
	Items       []SaleItemResponse `json:"items"`
}

type SaleItemResponse struct {
	Id     int64               `json:"id"`
	Name   string              `json:"name"`
	Price  float64             `json:"price"`
	Status sale.SaleItemStatus `json:"status"`
}

// SaleSummaryResponse is what a browse result carries. It leaves out items,
// which search does not load, so the shape cannot promise data that is not there.
type SaleSummaryResponse struct {
	Id          string          `json:"id"`
	SellerId    string          `json:"seller_id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Date        time.Time       `json:"date"`
	Status      sale.Status     `json:"status"`
	Address     AddressResponse `json:"address"`
}

type SaleSearchResponse struct {
	Sales   []SaleSummaryResponse `json:"sales"`
	Limit   int                   `json:"limit"`
	Offset  int                   `json:"offset"`
	HasMore bool                  `json:"has_more"`
}

func NewSaleResponse(s sale.Sale) *SaleResponse {
	return &SaleResponse{
		Id:          s.Id(),
		SellerId:    s.SellerId(),
		Name:        s.Name(),
		Description: s.Description(),
		Date:        s.Date(),
		Status:      s.Status(),
		Address:     *NewAddressResponse(s.Address()),
		Items:       newSaleItemResponses(s.Items()),
	}
}

func NewSaleSummaryResponse(s sale.Sale) SaleSummaryResponse {
	return SaleSummaryResponse{
		Id:          s.Id(),
		SellerId:    s.SellerId(),
		Name:        s.Name(),
		Description: s.Description(),
		Date:        s.Date(),
		Status:      s.Status(),
		Address:     *NewAddressResponse(s.Address()),
	}
}

func NewSaleSearchResponse(sales []sale.Sale, limit int, offset int, hasMore bool) SaleSearchResponse {
	summaries := make([]SaleSummaryResponse, 0, len(sales))
	for _, s := range sales {
		summaries = append(summaries, NewSaleSummaryResponse(s))
	}

	return SaleSearchResponse{
		Sales:   summaries,
		Limit:   limit,
		Offset:  offset,
		HasMore: hasMore,
	}
}

func newSaleItemResponses(items []sale.SaleItem) []SaleItemResponse {
	result := make([]SaleItemResponse, 0, len(items))
	for _, item := range items {
		result = append(result, SaleItemResponse{
			Id:     item.Id(),
			Name:   item.Name(),
			Price:  item.Price(),
			Status: item.Status(),
		})
	}

	return result
}
