package database

import (
	"GarageSaleAPI/domain/address"
	"GarageSaleAPI/domain/sale"
	"GarageSaleAPI/domain/seller"
	"GarageSaleAPI/domain/user"
	"GarageSaleAPI/infrastructure/persistence/database/records"
)

func userToRecord(u *user.User) records.UserRecord {
	return records.UserRecord{
		Id:        u.Id(),
		Username:  u.Username(),
		Password:  u.Password(),
		Email:     u.Email(),
		CreatedAt: u.CreatedAt(),
		UpdatedAt: u.UpdatedAt(),
	}
}

func recordToUser(r records.UserRecord) *user.User {
	return user.HydrateUser(r.Id, r.Username, r.Password, r.Email, r.CreatedAt, r.UpdatedAt)
}

func sellerToRecord(s *seller.Seller) records.SellerRecord {
	return records.SellerRecord{
		Id:        s.Id(),
		UserId:    s.UserId(),
		CreatedAt: s.CreatedAt(),
	}
}

func addressToRecord(a address.Address) records.AddressRecord {
	return records.AddressRecord{
		Id:         a.Id(),
		Line1:      a.Line1(),
		Line2:      a.Line2(),
		City:       a.City(),
		State:      a.State(),
		PostalCode: a.PostalCode(),
		Country:    a.Country(),
		Latitude:   a.Latitude(),
		Longitude:  a.Longitude(),
	}
}

func recordToAddress(r records.AddressRecord) *address.Address {
	return address.HydrateAddress(r.Id, r.Line1, r.Line2, r.City, r.State, r.PostalCode, r.Country, r.Latitude, r.Longitude)
}

func savedAddressToRecord(sa *seller.SavedAddress) records.SavedAddressRecord {
	a := sa.Address()
	return records.SavedAddressRecord{
		Id:        sa.Id(),
		SellerId:  sa.SellerId(),
		Label:     sa.Label(),
		AddressId: a.Id(),
		IsDefault: sa.IsDefault(),
	}
}

func recordToSavedAddress(r records.SavedAddressRecord) *seller.SavedAddress {
	addr := recordToAddress(r.Address)
	return seller.HydrateSavedAddress(r.Id, r.SellerId, r.Label, *addr, r.IsDefault)
}

func inventoryItemToRecord(i *seller.InventoryItem) records.InventoryItemRecord {
	return records.InventoryItemRecord{
		Id:          i.Id(),
		SellerId:    i.SellerId(),
		Name:        i.Name(),
		Description: i.Description(),
		Price:       i.Price(),
		Status:      string(i.Status()),
	}
}

func recordToInventoryItem(r records.InventoryItemRecord) *seller.InventoryItem {
	return seller.HydrateInventoryItem(r.Id, r.SellerId, r.Name, r.Description, r.Price, seller.ItemStatus(r.Status))
}

func saleToRecord(s *sale.Sale) records.SaleRecord {
	a := s.Address()
	return records.SaleRecord{
		Id:          s.Id(),
		SellerId:    s.SellerId(),
		Name:        s.Name(),
		AddressId:   a.Id(),
		Date:        s.Date(),
		Description: s.Description(),
		Status:      string(s.Status()),
		CreatedAt:   s.CreatedAt(),
	}
}

func recordToSale(r records.SaleRecord, items []sale.SaleItem) *sale.Sale {
	addr := recordToAddress(r.Address)
	return sale.HydrateSale(r.Id, r.SellerId, r.Name, *addr, r.Date, r.Description, items, sale.Status(r.Status), r.CreatedAt)
}

func saleItemToRecord(saleID string, i sale.SaleItem) records.SaleItemRecord {
	return records.SaleItemRecord{
		Id:              i.Id(),
		SaleId:          saleID,
		InventoryItemId: i.InventoryItemId(),
		Name:            i.Name(),
		Price:           i.Price(),
		Status:          string(i.Status()),
	}
}

func recordToSaleItem(r records.SaleItemRecord) *sale.SaleItem {
	return sale.HydrateSaleItem(r.Id, r.SaleId, r.InventoryItemId, r.Name, r.Price, sale.SaleItemStatus(r.Status))
}
