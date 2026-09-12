package database

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/token"
	"GarageSaleAPI/infrastructure/persistence/database/records"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type RevokedTokenRepository struct {
	db *gorm.DB
}

func NewRevokedTokenRepository(db *gorm.DB) *RevokedTokenRepository {
	return &RevokedTokenRepository{db: db}
}

func (r *RevokedTokenRepository) Revoke(ctx context.Context, t *token.RevokedToken) error {
	db := r.db.WithContext(ctx)

	record := new(revokedTokenToRecord(t))
	if err := db.Create(record).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil
		}
		return apperror.Internal(err)
	}
	return nil
}

func (r *RevokedTokenRepository) IsRevoked(ctx context.Context, jti string) (bool, error) {
	db := r.db.WithContext(ctx)

	var record records.RevokedTokenRecord
	if err := db.First(&record, "jti = ?", jti).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, apperror.Internal(err)
	}
	return true, nil
}

func (r *RevokedTokenRepository) DeleteExpired(ctx context.Context, cutoff time.Time) error {
	db := r.db.WithContext(ctx)

	if err := db.Where("expires_at < ?", cutoff).Delete(&records.RevokedTokenRecord{}).Error; err != nil {
		return apperror.Internal(err)
	}
	return nil
}
