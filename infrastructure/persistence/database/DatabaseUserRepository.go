package database

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/user"
	"GarageSaleAPI/infrastructure/persistence/database/records"
	"context"
	"errors"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	db := r.db.WithContext(ctx)

	record := new(userToRecord(u))
	if err := db.Create(record).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return apperror.Conflict("user already exists", err)
		}
		return apperror.Internal(err)
	}
	return nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	db := r.db.WithContext(ctx)

	var record records.UserRecord
	if err := db.First(&record, "username = ?", username).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("user not found", err)
		}
	}
	return recordToUser(record), nil
}

func (r *UserRepository) GetById(ctx context.Context, id string) (*user.User, error) {
	db := r.db.WithContext(ctx)

	var record records.UserRecord
	if err := db.First(&record, "id = ?", id).Error; err != nil {
		return nil, apperror.NotFound("user not found", nil)
	}
	return recordToUser(record), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	db := r.db.WithContext(ctx)

	var record records.UserRecord
	if err := db.First(&record, "email = ?", email).Error; err != nil {
		return nil, apperror.NotFound("user not found", nil)
	}
	return recordToUser(record), nil
}
