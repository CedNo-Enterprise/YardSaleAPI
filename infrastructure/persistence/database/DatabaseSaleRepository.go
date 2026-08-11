package database

import (
	"GarageSaleAPI/domain/sale"
	"GarageSaleAPI/infrastructure/persistence/database/records"
	"context"
	"fmt"

	"gorm.io/gorm"
)

type SaleRepository struct {
	db *gorm.DB
}

func NewSaleRepository(db *gorm.DB) *SaleRepository {
	return &SaleRepository{db: db}
}

// todo: Change errors to apperror types
// todo: probably do not need to insert items and addresses since theyre empty on creation
func (r *SaleRepository) Create(ctx context.Context, s *sale.Sale) error {
	db := r.db.WithContext(ctx)

	addrRecord := addressToRecord(s.Address())
	if err := db.Create(&addrRecord).Error; err != nil {
		return fmt.Errorf("create address: %w", err)
	}

	saleRecord := saleToRecord(s)
	saleRecord.AddressId = addrRecord.Id
	if err := db.Create(&saleRecord).Error; err != nil {
		return fmt.Errorf("create sale: %w", err)
	}

	itemRecords := make([]records.SaleItemRecord, len(s.Items()))
	for i, item := range s.Items() {
		itemRecords[i] = saleItemToRecord(saleRecord.Id, item)
	}
	if len(itemRecords) > 0 {
		if err := db.Create(&itemRecords).Error; err != nil {
			return fmt.Errorf("create sale items: %w", err)
		}
	}

	items := make([]sale.SaleItem, len(itemRecords))
	for i, rec := range itemRecords {
		items[i] = *recordToSaleItem(rec)
	}
	saleRecord.Address = addrRecord

	return nil
}

func (r *SaleRepository) GetById(ctx context.Context, id string) (*sale.Sale, error) {
	db := r.db.WithContext(ctx)

	var saleRecord records.SaleRecord
	if err := db.Preload("Address").First(&saleRecord, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("get sale: %w", err)
	}

	var itemRecords []records.SaleItemRecord
	if err := db.Where("sale_id = ?", id).Find(&itemRecords).Error; err != nil {
		return nil, fmt.Errorf("get sale items: %w", err)
	}

	items := make([]sale.SaleItem, len(itemRecords))
	for i, rec := range itemRecords {
		items[i] = *recordToSaleItem(rec)
	}

	return recordToSale(saleRecord, items), nil
}
