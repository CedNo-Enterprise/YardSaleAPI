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

func (r *SaleRepository) Search(ctx context.Context, criteria sale.SearchCriteria) ([]sale.Sale, error) {
	criteria = criteria.Normalize()

	db := r.db.WithContext(ctx)

	statuses := make([]string, 0, len(criteria.Statuses))
	for _, status := range criteria.Statuses {
		statuses = append(statuses, string(status))
	}

	query := db.Model(&records.SaleRecord{}).Preload("Address").Where("status IN ?", statuses)
	if criteria.DateFrom != nil {
		query = query.Where("date >= ?", *criteria.DateFrom)
	}
	if criteria.DateTo != nil {
		query = query.Where("date <= ?", *criteria.DateTo)
	}

	var saleRecords []records.SaleRecord
	err := query.
		Order(orderClause(criteria.Sort)).
		Limit(criteria.FetchLimit()).
		Offset(criteria.Offset).
		Find(&saleRecords).Error
	if err != nil {
		return nil, apperror.Internal(err)
	}

	// Items are left empty on purpose: no browse response shows them, so loading
	// them would be a second query per page for data nobody reads.
	sales := make([]sale.Sale, len(saleRecords))
	for i, rec := range saleRecords {
		sales[i] = *recordToSale(rec, []sale.SaleItem{})
	}

	return sales, nil
}

// orderClause spells a sort order as SQL. Every order breaks ties on id so two
// requests for the same page cannot disagree about where the boundary falls.
func orderClause(order sale.SortOrder) string {
	switch order {
	case sale.SortDateReverse:
		return "date DESC, id ASC"
	case sale.SortCreated:
		return "created_at DESC, id ASC"
	default:
		return "date ASC, id ASC"
	}
}
