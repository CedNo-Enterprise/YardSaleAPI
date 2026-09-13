package records

// (itinerary_id, sale_id) is the primary key because it is the stop's real
// identity and the invariant worth enforcing — a sale cannot appear twice in one
// itinerary. Keying on it means reordering cannot churn stop identity, so
// per-stop data added later can reference a stop safely.
//
// Position is deliberately part of no unique index: reordering updates it in
// place, and rows transiently share a position partway through that transaction.
type ItineraryStopRecord struct {
	ItineraryId string `gorm:"column:itinerary_id;type:uuid;primaryKey"`
	SaleId      string `gorm:"column:sale_id;type:uuid;primaryKey"`
	Position    int    `gorm:"column:position;not null"`
	Status      string `gorm:"column:status;not null;default:planned"`
}

func (ItineraryStopRecord) TableName() string { return "itinerary_stops" }
