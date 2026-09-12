package database

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/seller"
	"GarageSaleAPI/infrastructure/persistence/database/records"
	"context"
	"errors"

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
	if err := db.Create(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return apperror.Conflict("seller already exists", err)
		}
		return apperror.Internal(err)
	}
	return nil
}

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
	if err := db.Preload("Address").Where("seller_id = ?", id).Find(&savedAddressRecords).Error; err != nil {
		return nil, apperror.Internal(err)
	}

	var inventoryRecords []records.InventoryItemRecord
	if err := db.Where("seller_id = ?", id).Find(&inventoryRecords).Error; err != nil {
		return nil, apperror.Internal(err)
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

func (r *SellerRepository) GetByUserId(ctx context.Context, userId string) (*seller.Seller, error) {
	db := r.db.WithContext(ctx)

	var sellerRecord records.SellerRecord
	if err := db.First(&sellerRecord, "user_id = ?", userId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("seller not found", err)
		}
		return nil, apperror.Internal(err)
	}

	var savedAddressRecords []records.SavedAddressRecord
	if err := db.Preload("Address").Where("seller_id = ?", sellerRecord.Id).Find(&savedAddressRecords).Error; err != nil {
		return nil, apperror.Internal(err)
	}

	var inventoryRecords []records.InventoryItemRecord
	if err := db.Where("seller_id = ?", sellerRecord.Id).Find(&inventoryRecords).Error; err != nil {
		return nil, apperror.Internal(err)
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
