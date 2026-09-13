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

	db, err := gorm.Open(postgres.Open(url), &gorm.Config{TranslateError: true})
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return db, nil
}

// widenSellerUserIdToUuid retypes sellers.user_id, which was created as text
// while every other id column is uuid. AutoMigrate cannot do it alone:
// Postgres refuses a text-to-uuid column cast without an explicit USING.
func widenSellerUserIdToUuid(db *gorm.DB) error {
	var dataType string
	err := db.Raw(
		`SELECT data_type FROM information_schema.columns
		 WHERE table_schema = current_schema()
		   AND table_name = 'sellers' AND column_name = 'user_id'`,
	).Scan(&dataType).Error
	if err != nil {
		return err
	}

	// Absent on a fresh database, where AutoMigrate creates it as uuid anyway.
	if dataType != "text" {
		return nil
	}

	return db.Exec(`ALTER TABLE sellers ALTER COLUMN user_id TYPE uuid USING user_id::uuid`).Error
}

func AutoMigrate(db *gorm.DB) error {
	if err := widenSellerUserIdToUuid(db); err != nil {
		return err
	}

	return db.AutoMigrate(
		&records.UserRecord{},
		&records.RevokedTokenRecord{},
		&records.SellerRecord{},
		&records.AddressRecord{},
		&records.BuyerRecord{},
		&records.SavedAddressRecord{},
		&records.InventoryItemRecord{},
		&records.SaleRecord{},
		&records.SaleItemRecord{},
		&records.ItineraryRecord{},
		&records.ItineraryStopRecord{},
	)
}
