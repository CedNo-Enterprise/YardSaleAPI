package records

import "time"

type ItineraryRecord struct {
	Id             string    `gorm:"column:id;type:uuid;primaryKey"`
	UserId         string    `gorm:"column:user_id;type:uuid;not null;index"`
	Name           string    `gorm:"column:name;not null"`
	Description    string    `gorm:"column:description"`
	Date           time.Time `gorm:"column:date;not null"`
	StartLatitude  float64   `gorm:"column:start_latitude;not null;default:0"`
	StartLongitude float64   `gorm:"column:start_longitude;not null;default:0"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;default:now()"`
}

func (ItineraryRecord) TableName() string { return "itineraries" }
