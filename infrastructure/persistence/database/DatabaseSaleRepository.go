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

func (r *SaleRepository) GetByIds(ctx context.Context, ids []string) ([]sale.Sale, error) {
	if len(ids) == 0 {
		// Postgres rejects "IN ()", so never build that query.
		return []sale.Sale{}, nil
	}

	db := r.db.WithContext(ctx)

	var saleRecords []records.SaleRecord
	// Find on a slice yields no error for zero rows, so there is no not-found case.
	if err := db.Preload("Address").Where("id IN ?", ids).Find(&saleRecords).Error; err != nil {
		return nil, apperror.Internal(err)
	}

	var itemRecords []records.SaleItemRecord
	if err := db.Where("sale_id IN ?", ids).Find(&itemRecords).Error; err != nil {
		return nil, apperror.Internal(err)
	}

	// Grouped in one pass rather than a query per sale.
	itemsBySale := make(map[string][]sale.SaleItem, len(saleRecords))
	for _, rec := range itemRecords {
		itemsBySale[rec.SaleId] = append(itemsBySale[rec.SaleId], *recordToSaleItem(rec))
	}

	sales := make([]sale.Sale, len(saleRecords))
	for i, rec := range saleRecords {
		sales[i] = *recordToSale(rec, itemsBySale[rec.Id])
	}

	return sales, nil
}
