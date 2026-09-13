package requests

import "time"

type ItineraryRequest struct {
	Name           string    `json:"name"           validate:"required,max=64"`
	Description    string    `json:"description"    validate:"max=500"`
	Date           time.Time `json:"date"           validate:"required"`
	StartLatitude  *float64  `json:"startLatitude"  validate:"omitempty,latitude"`
	StartLongitude *float64  `json:"startLongitude" validate:"omitempty,longitude"`
	SaleIds        []string  `json:"saleIds"        validate:"omitempty,max=50,unique,dive,required"`
}

// UpdateItineraryRequest uses pointers so a PATCH can tell "field absent" from
// "field set to empty" — clearing a description needs to be expressible. An
// explicit JSON null is indistinguishable from absent here; both mean "leave
// alone", which is the pragmatic reading.
type UpdateItineraryRequest struct {
	Name           *string    `json:"name"           validate:"omitempty,max=64"`
	Description    *string    `json:"description"    validate:"omitempty,max=500"`
	Date           *time.Time `json:"date"`
	StartLatitude  *float64   `json:"startLatitude"  validate:"omitempty,latitude"`
	StartLongitude *float64   `json:"startLongitude" validate:"omitempty,longitude"`
}

type AddStopRequest struct {
	SaleId string `json:"saleId" validate:"required"`
}

// ReorderStopsRequest carries the point to optimize from. Omitted coordinates
// reuse the start already stored on the itinerary.
type ReorderStopsRequest struct {
	StartLatitude  *float64 `json:"startLatitude"  validate:"omitempty,latitude"`
	StartLongitude *float64 `json:"startLongitude" validate:"omitempty,longitude"`
}

type UpdateStopStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=planned visited skipped"`
}
