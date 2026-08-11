package database

import (
	"GarageSaleAPI/infrastructure/persistence/database/records"
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewGormDB() (*gorm.DB, error) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(url), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&records.UserRecord{},
		&records.SellerRecord{},
		&records.AddressRecord{},
		&records.SavedAddressRecord{},
		&records.InventoryItemRecord{},
		&records.SaleRecord{},
		&records.SaleItemRecord{},
	)
}
