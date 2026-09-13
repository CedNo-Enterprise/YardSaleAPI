package itinerary

import "context"

type ItineraryRepository interface {
	Create(context.Context, *Itinerary) error
	GetById(context.Context, string) (*Itinerary, error)
	GetByUserId(context.Context, string) ([]Itinerary, error)
	Update(context.Context, *Itinerary) error
	Delete(context.Context, string) error
	AddStop(context.Context, *Stop) error
	RemoveStop(context.Context, string, string) error
	ReorderStops(context.Context, *Itinerary) error
	UpdateStopStatus(context.Context, string, string, StopStatus) error
}
