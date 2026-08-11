package database

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/seller"
	"GarageSaleAPI/infrastructure/persistence/database/records"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type SellerRepository struct {
	db *gorm.DB
}

func NewSellerRepository(db *gorm.DB) *SellerRepository {
	return &SellerRepository{db: db}
}

func (r *SellerRepository) Create(ctx context.Context, s *seller.Seller) error {
	db := r.db.WithContext(ctx)

	record := sellerToRecord(s)
	if err := db.Create(record).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return apperror.Conflict("seller already exists", err)
		}
		return apperror.Internal(err)
	}
	return nil
}

// todo: Change errors to apperror types
func (r *SellerRepository) GetById(ctx context.Context, id string) (*seller.Seller, error) {
	db := r.db.WithContext(ctx)

	var sellerRecord records.SellerRecord
	if err := db.First(&sellerRecord, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("seller not found", err)
		}
		return nil, apperror.Internal(err)
	}

	var savedAddressRecords []records.SavedAddressRecord
	if err := db.Preload("Address").Where("sellerId = ?", id).Find(&savedAddressRecords).Error; err != nil {
		return nil, fmt.Errorf("get saved addresses: %w", err)
	}

	var inventoryRecords []records.InventoryItemRecord
	if err := db.Where("sellerId = ?", id).Find(&inventoryRecords).Error; err != nil {
		return nil, fmt.Errorf("get inventory: %w", err)
	}

	savedAddresses := make([]seller.SavedAddress, len(savedAddressRecords))
	for i, rec := range savedAddressRecords {
		savedAddresses[i] = *recordToSavedAddress(rec)
	}

	inventory := make([]seller.InventoryItem, len(inventoryRecords))
	for i, rec := range inventoryRecords {
		inventory[i] = *recordToInventoryItem(rec)
	}

	return seller.HydrateSeller(
		sellerRecord.Id,
		sellerRecord.UserId,
		sellerRecord.Name,
		savedAddresses,
		inventory,
		sellerRecord.CreatedAt,
	), nil
}

// todo: Change errors to apperror types
func (r *SellerRepository) GetByUserId(ctx context.Context, userId string) (*seller.Seller, error) {
	db := r.db.WithContext(ctx)

	var sellerRecord records.SellerRecord
	if err := db.First(&sellerRecord, "userId = ?", userId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("seller not found", err)
		}
		return nil, apperror.Internal(err)
	}

	var savedAddressRecords []records.SavedAddressRecord
	if err := db.Preload("Address").Where("sellerId = ?", sellerRecord.Id).Find(&savedAddressRecords).Error; err != nil {
		return nil, fmt.Errorf("get saved addresses: %w", err)
	}

	var inventoryRecords []records.InventoryItemRecord
	if err := db.Where("sellerId = ?", sellerRecord.Id).Find(&inventoryRecords).Error; err != nil {
		return nil, fmt.Errorf("get inventory: %w", err)
	}

	savedAddresses := make([]seller.SavedAddress, len(savedAddressRecords))
	for i, rec := range savedAddressRecords {
		savedAddresses[i] = *recordToSavedAddress(rec)
	}

	inventory := make([]seller.InventoryItem, len(inventoryRecords))
	for i, rec := range inventoryRecords {
		inventory[i] = *recordToInventoryItem(rec)
	}

	return seller.HydrateSeller(
		sellerRecord.Id,
		sellerRecord.UserId,
		sellerRecord.Name,
		savedAddresses,
		inventory,
		sellerRecord.CreatedAt,
	), nil
}
