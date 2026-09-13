package records

import "time"

type BuyerRecord struct {
	Id          string    `gorm:"column:id;type:uuid;primaryKey"`
	UserId      string    `gorm:"column:user_id;type:uuid;not null;uniqueIndex"`
	DisplayName string    `gorm:"column:display_name;not null"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:now()"`

	// Null until the buyer saves a home address; loaded via .Preload("HomeAddress").
	HomeAddressId *int64         `gorm:"column:home_address_id"`
	HomeAddress   *AddressRecord `gorm:"foreignKey:HomeAddressId;references:Id"`
}

func (BuyerRecord) TableName() string { return "buyers" }
