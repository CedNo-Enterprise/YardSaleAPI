package records

import "time"

type SellerRecord struct {
	Id        string    `gorm:"column:id;type:uuid;primaryKey"`
	UserId    string    `gorm:"column:user_id;not null"`
	Name      string    `gorm:"column:name;not null"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:now()"`
}

func (SellerRecord) TableName() string { return "sellers" }
