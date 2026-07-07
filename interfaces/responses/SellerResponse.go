package responses

import (
	"GarageSaleAPI/domain/seller"
)

type SellerResponse struct {
	Id             string                  `json:"id"`
	Username       string                  `json:"username"`
	SavedAddresses []SavedAddressResponse  `json:"saved_addresses"`
	Inventory      []InventoryItemResponse `json:"inventory"`
}

type SavedAddressResponse struct {
	Id        string          `json:"id"`
	Label     string          `json:"label"`
	Address   AddressResponse `json:"address"`
	IsDefault bool            `json:"is_default"`
}

type InventoryItemResponse struct {
	Id          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Price       float64           `json:"price"`
	Status      seller.ItemStatus `json:"status"`
}

func NewSellerResponse(s *seller.Seller) SellerResponse {
	return SellerResponse{
		Id:             s.Id(),
		Username:       s.Username(),
		SavedAddresses: newSavedAddressResponses(s.SavedAddresses()),
		Inventory:      newInventoryItemResponses(s.Inventory()),
	}
}

func newSavedAddressResponses(addresses []seller.SavedAddress) []SavedAddressResponse {
	result := make([]SavedAddressResponse, 0, len(addresses))
	for _, a := range addresses {
		result = append(result, newSavedAddressResponse(a))
	}
	return result
}

func newSavedAddressResponse(a seller.SavedAddress) SavedAddressResponse {
	return SavedAddressResponse{
		Id:        a.Id(),
		Label:     a.Label(),
		Address:   *NewAddressResponse(a.Address()),
		IsDefault: a.IsDefault(),
	}
}

func newInventoryItemResponses(items []seller.InventoryItem) []InventoryItemResponse {
	result := make([]InventoryItemResponse, 0, len(items))
	for _, item := range items {
		result = append(result, newInventoryItemResponse(item))
	}
	return result
}

func newInventoryItemResponse(item seller.InventoryItem) InventoryItemResponse {
	return InventoryItemResponse{
		Id:          item.Id(),
		Name:        item.Name(),
		Description: item.Description(),
		Price:       item.Price(),
		Status:      item.Status(),
	}
}
