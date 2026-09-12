package records

import "time"

type SaleRecord struct {
	Id          string    `gorm:"column:id;type:uuid;primaryKey"`
	SellerId    string    `gorm:"column:seller_id;type:uuid;not null;index"`
	Name        string    `gorm:"column:name"`
	AddressId   int64     `gorm:"column:address_id;not null"`
	Date        time.Time `gorm:"column:date;not null"`
	Description string    `gorm:"column:description"`
	Status      string    `gorm:"column:status;default:scheduled"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:now()"`

	// Loaded via .Preload("Address") when needed — addresses is its own table.
	Address AddressRecord `gorm:"foreignKey:AddressID;references:ID"`
}

func (SaleRecord) TableName() string { return "sales" }
