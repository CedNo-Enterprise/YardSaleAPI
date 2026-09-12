package database

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/sale"
	"GarageSaleAPI/infrastructure/persistence/database/records"
	"context"
	"errors"

	"gorm.io/gorm"
)

type SaleRepository struct {
	db *gorm.DB
}

func NewSaleRepository(db *gorm.DB) *SaleRepository {
	return &SaleRepository{db: db}
}

func (r *SaleRepository) Create(ctx context.Context, s *sale.Sale) error {
	db := r.db.WithContext(ctx)

	addrRecord := addressToRecord(s.Address())
	if err := db.Create(&addrRecord).Error; err != nil {
		return apperror.Internal(err)
	}

	saleRecord := saleToRecord(s)
	saleRecord.AddressId = addrRecord.Id
	if err := db.Create(&saleRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return apperror.Conflict("sale already exists", err)
		}
		return apperror.Internal(err)
	}

	return nil
}

func (r *SaleRepository) GetById(ctx context.Context, id string) (*sale.Sale, error) {
	db := r.db.WithContext(ctx)

	var saleRecord records.SaleRecord
	if err := db.Preload("Address").First(&saleRecord, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("sale not found", err)
		}
		return nil, apperror.Internal(err)
	}

	var itemRecords []records.SaleItemRecord
	if err := db.Where("sale_id = ?", id).Find(&itemRecords).Error; err != nil {
		return nil, apperror.Internal(err)
	}

	items := make([]sale.SaleItem, len(itemRecords))
	for i, rec := range itemRecords {
		items[i] = *recordToSaleItem(rec)
	}

	return recordToSale(saleRecord, items), nil
}
