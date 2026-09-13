package database

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/itinerary"
	"GarageSaleAPI/infrastructure/persistence/database/records"
	"context"
	"errors"

	"gorm.io/gorm"
)

type ItineraryRepository struct {
	db *gorm.DB
}

func NewItineraryRepository(db *gorm.DB) *ItineraryRepository {
	return &ItineraryRepository{db: db}
}

func (r *ItineraryRepository) Create(ctx context.Context, i *itinerary.Itinerary) error {
	// WithContext before Transaction so a cancelled request aborts the transaction.
	db := r.db.WithContext(ctx)

	return db.Transaction(func(tx *gorm.DB) error {
		record := itineraryToRecord(i)
		if err := tx.Create(&record).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return apperror.Conflict("itinerary already exists", err)
			}
			return apperror.Internal(err)
		}

		stops := i.Stops()
		if len(stops) == 0 {
			return nil
		}

		stopRecords := make([]records.ItineraryStopRecord, len(stops))
		for index, stop := range stops {
			stopRecords[index] = itineraryStopToRecord(stop)
		}

		// One multi-row insert rather than a statement per stop.
		if err := tx.Create(&stopRecords).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return apperror.Conflict("duplicate sale in itinerary", err)
			}
			return apperror.Internal(err)
		}

		return nil
	})
}

func (r *ItineraryRepository) GetById(ctx context.Context, id string) (*itinerary.Itinerary, error) {
	db := r.db.WithContext(ctx)

	var record records.ItineraryRecord
	if err := db.First(&record, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("itinerary not found", err)
		}
		return nil, apperror.Internal(err)
	}

	var stopRecords []records.ItineraryStopRecord
	if err := db.Where("itinerary_id = ?", id).
		Order(`"position" ASC`).Find(&stopRecords).Error; err != nil {
		return nil, apperror.Internal(err)
	}

	stops := make([]itinerary.Stop, len(stopRecords))
	for index, rec := range stopRecords {
		stops[index] = recordToItineraryStop(rec)
	}

	return recordToItinerary(record, stops), nil
}

func (r *ItineraryRepository) GetByUserId(ctx context.Context, userId string) ([]itinerary.Itinerary, error) {
	db := r.db.WithContext(ctx)

	var itineraryRecords []records.ItineraryRecord
	if err := db.Where("user_id = ?", userId).
		Order("date ASC").Find(&itineraryRecords).Error; err != nil {
		return nil, apperror.Internal(err)
	}

	if len(itineraryRecords) == 0 {
		return []itinerary.Itinerary{}, nil
	}

	ids := make([]string, len(itineraryRecords))
	for index, rec := range itineraryRecords {
		ids[index] = rec.Id
	}

	// One query for every route's stops rather than one per itinerary.
	var stopRecords []records.ItineraryStopRecord
	if err := db.Where("itinerary_id IN ?", ids).
		Order(`itinerary_id, "position" ASC`).Find(&stopRecords).Error; err != nil {
		return nil, apperror.Internal(err)
	}

	stopsByItinerary := make(map[string][]itinerary.Stop, len(itineraryRecords))
	for _, rec := range stopRecords {
		stopsByItinerary[rec.ItineraryId] = append(
			stopsByItinerary[rec.ItineraryId], recordToItineraryStop(rec),
		)
	}

	itineraries := make([]itinerary.Itinerary, len(itineraryRecords))
	for index, rec := range itineraryRecords {
		stops := stopsByItinerary[rec.Id]
		if stops == nil {
			stops = []itinerary.Stop{}
		}
		itineraries[index] = *recordToItinerary(rec, stops)
	}

	return itineraries, nil
}

func (r *ItineraryRepository) Update(ctx context.Context, i *itinerary.Itinerary) error {
	db := r.db.WithContext(ctx)

	// Scalar columns only: stop rows belong to the stop operations.
	result := db.Model(&records.ItineraryRecord{}).Where("id = ?", i.Id()).
		Updates(map[string]any{
			"name":            i.Name(),
			"description":     i.Description(),
			"date":            i.Date(),
			"start_latitude":  i.StartLatitude(),
			"start_longitude": i.StartLongitude(),
		})
	if result.Error != nil {
		return apperror.Internal(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperror.NotFound("itinerary not found", nil)
	}

	return nil
}

func (r *ItineraryRepository) Delete(ctx context.Context, id string) error {
	db := r.db.WithContext(ctx)

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("itinerary_id = ?", id).
			Delete(&records.ItineraryStopRecord{}).Error; err != nil {
			return apperror.Internal(err)
		}

		// GORM does not report ErrRecordNotFound for deletes, so check the count.
		result := tx.Delete(&records.ItineraryRecord{}, "id = ?", id)
		if result.Error != nil {
			return apperror.Internal(result.Error)
		}
		if result.RowsAffected == 0 {
			return apperror.NotFound("itinerary not found", nil)
		}

		return nil
	})
}

func (r *ItineraryRepository) AddStop(ctx context.Context, s *itinerary.Stop) error {
	db := r.db.WithContext(ctx)

	record := itineraryStopToRecord(*s)
	if err := db.Create(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return apperror.Conflict("sale already in itinerary", err)
		}
		return apperror.Internal(err)
	}

	return nil
}

func (r *ItineraryRepository) RemoveStop(ctx context.Context, itineraryId string, saleId string) error {
	db := r.db.WithContext(ctx)

	return db.Transaction(func(tx *gorm.DB) error {
		var removed records.ItineraryStopRecord
		if err := tx.First(&removed, "itinerary_id = ? AND sale_id = ?", itineraryId, saleId).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperror.NotFound("stop not found", err)
			}
			return apperror.Internal(err)
		}

		if err := tx.Where("itinerary_id = ? AND sale_id = ?", itineraryId, saleId).
			Delete(&records.ItineraryStopRecord{}).Error; err != nil {
			return apperror.Internal(err)
		}

		// Close the gap so positions stay contiguous. Safe as a single statement
		// precisely because no unique index constrains position.
		if err := tx.Model(&records.ItineraryStopRecord{}).
			Where(`itinerary_id = ? AND "position" > ?`, itineraryId, removed.Position).
			Update("position", gorm.Expr(`"position" - 1`)).Error; err != nil {
			return apperror.Internal(err)
		}

		return nil
	})
}

func (r *ItineraryRepository) ReorderStops(ctx context.Context, i *itinerary.Itinerary) error {
	db := r.db.WithContext(ctx)

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&records.ItineraryRecord{}).Where("id = ?", i.Id()).
			Updates(map[string]any{
				"start_latitude":  i.StartLatitude(),
				"start_longitude": i.StartLongitude(),
			}).Error; err != nil {
			return apperror.Internal(err)
		}

		// Positions are updated in place so stop rows survive the reorder, and
		// with them any per-stop state such as status.
		for _, stop := range i.Stops() {
			if err := tx.Model(&records.ItineraryStopRecord{}).
				Where("itinerary_id = ? AND sale_id = ?", stop.ItineraryId(), stop.SaleId()).
				Update("position", stop.Position()).Error; err != nil {
				return apperror.Internal(err)
			}
		}

		return nil
	})
}

func (r *ItineraryRepository) UpdateStopStatus(
	ctx context.Context, itineraryId string, saleId string, status itinerary.StopStatus,
) error {
	db := r.db.WithContext(ctx)

	result := db.Model(&records.ItineraryStopRecord{}).
		Where("itinerary_id = ? AND sale_id = ?", itineraryId, saleId).
		Update("status", string(status))
	if result.Error != nil {
		return apperror.Internal(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperror.NotFound("stop not found", nil)
	}

	return nil
}
