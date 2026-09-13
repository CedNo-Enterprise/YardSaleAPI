package database

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/buyer"
	"GarageSaleAPI/infrastructure/persistence/database/records"
	"context"
	"errors"

	"gorm.io/gorm"
)

type BuyerRepository struct {
	db *gorm.DB
}

func NewBuyerRepository(db *gorm.DB) *BuyerRepository {
	return &BuyerRepository{db: db}
}

func (r *BuyerRepository) Create(ctx context.Context, b *buyer.Buyer) error {
	db := r.db.WithContext(ctx)

	// The address row has to exist before the buyer can point at it, and a buyer
	// left behind by a failed address write would be worse than no buyer at all.
	err := db.Transaction(func(tx *gorm.DB) error {
		record := buyerToRecord(b)

		if home := b.HomeAddress(); home != nil {
			addressRecord := addressToRecord(*home)
			if err := tx.Create(&addressRecord).Error; err != nil {
				return apperror.Internal(err)
			}
			record.HomeAddressId = &addressRecord.Id
		}

		if err := tx.Create(&record).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return apperror.Conflict("buyer already exists", err)
			}
			return apperror.Internal(err)
		}

		return nil
	})

	return err
}

func (r *BuyerRepository) GetByUserId(ctx context.Context, userId string) (*buyer.Buyer, error) {
	db := r.db.WithContext(ctx)

	var buyerRecord records.BuyerRecord
	if err := db.Preload("HomeAddress").First(&buyerRecord, "user_id = ?", userId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("buyer not found", err)
		}
		return nil, apperror.Internal(err)
	}

	return recordToBuyer(buyerRecord), nil
}

func (r *BuyerRepository) Update(ctx context.Context, b *buyer.Buyer) error {
	db := r.db.WithContext(ctx)

	return db.Transaction(func(tx *gorm.DB) error {
		var stored records.BuyerRecord
		if err := tx.First(&stored, "id = ?", b.Id()).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperror.NotFound("buyer not found", err)
			}
			return apperror.Internal(err)
		}

		homeAddressId, err := saveHomeAddress(tx, b, stored.HomeAddressId)
		if err != nil {
			return err
		}

		updates := map[string]any{
			"display_name":    b.DisplayName(),
			"home_address_id": homeAddressId,
		}
		if err := tx.Model(&records.BuyerRecord{}).Where("id = ?", b.Id()).Updates(updates).Error; err != nil {
			return apperror.Internal(err)
		}

		return nil
	})
}

// saveHomeAddress writes the buyer's home address and reports the id the buyer
// row should carry. A buyer that already had an address keeps the same row, so
// an edit does not leave the old address orphaned behind it.
func saveHomeAddress(tx *gorm.DB, b *buyer.Buyer, storedAddressId *int64) (*int64, error) {
	home := b.HomeAddress()
	if home == nil {
		return storedAddressId, nil
	}

	record := addressToRecord(*home)

	if storedAddressId != nil {
		record.Id = *storedAddressId
		if err := tx.Save(&record).Error; err != nil {
			return nil, apperror.Internal(err)
		}
		return storedAddressId, nil
	}

	record.Id = 0
	if err := tx.Create(&record).Error; err != nil {
		return nil, apperror.Internal(err)
	}

	return &record.Id, nil
}
