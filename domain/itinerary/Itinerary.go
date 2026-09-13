package itinerary

import "time"

type Itinerary struct {
	id             string
	userId         string
	name           string
	description    string
	date           time.Time
	startLatitude  float64
	startLongitude float64
	stops          []Stop
	createdAt      time.Time
}

func (i *Itinerary) Id() string {
	return i.id
}

func (i *Itinerary) UserId() string {
	return i.userId
}

func (i *Itinerary) Name() string {
	return i.name
}

func (i *Itinerary) Description() string {
	return i.description
}

func (i *Itinerary) Date() time.Time {
	return i.date
}

func (i *Itinerary) StartLatitude() float64 {
	return i.startLatitude
}

func (i *Itinerary) StartLongitude() float64 {
	return i.startLongitude
}

// Start is the point the route is optimized from.
func (i *Itinerary) Start() GeoPoint {
	return GeoPoint{Latitude: i.startLatitude, Longitude: i.startLongitude}
}

func (i *Itinerary) Stops() []Stop {
	return i.stops
}

func (i *Itinerary) CreatedAt() time.Time {
	return i.createdAt
}

func (i *Itinerary) Rename(name string) {
	i.name = name
}

func (i *Itinerary) Describe(description string) {
	i.description = description
}

func (i *Itinerary) Reschedule(date time.Time) {
	i.date = date
}

func (i *Itinerary) SetStart(latitude float64, longitude float64) {
	i.startLatitude = latitude
	i.startLongitude = longitude
}

func (i *Itinerary) HasStop(saleId string) bool {
	for _, stop := range i.stops {
		if stop.saleId == saleId {
			return true
		}
	}

	return false
}

// AppendStop adds a sale to the end of the route and returns the new stop so the
// caller can persist it. Re-optimizing the route is a separate, explicit action.
func (i *Itinerary) AppendStop(saleId string) Stop {
	stop := CreateStop(i.id, saleId, len(i.stops))
	i.stops = append(i.stops, stop)

	return stop
}

// RemoveStopBySaleId drops a stop and renumbers the survivors so positions stay
// contiguous from zero. It reports whether anything was removed.
func (i *Itinerary) RemoveStopBySaleId(saleId string) bool {
	remaining := make([]Stop, 0, len(i.stops))
	for _, stop := range i.stops {
		if stop.saleId == saleId {
			continue
		}
		stop.position = len(remaining)
		remaining = append(remaining, stop)
	}

	if len(remaining) == len(i.stops) {
		return false
	}

	i.stops = remaining
	return true
}

// SetStopStatus updates one stop's status in place, reporting whether the stop
// is part of this itinerary.
func (i *Itinerary) SetStopStatus(saleId string, status StopStatus) bool {
	for index := range i.stops {
		if i.stops[index].saleId == saleId {
			i.stops[index].status = status
			return true
		}
	}

	return false
}

// ReorderFrom re-sequences the route as a nearest-neighbour walk from start and
// records start as the itinerary's new origin.
func (i *Itinerary) ReorderFrom(start GeoPoint, coordinates map[string]GeoPoint) {
	i.SetStart(start.Latitude, start.Longitude)
	i.stops = OrderStopsByProximity(i.stops, coordinates, start)
}

// Stop has no surrogate id: (itineraryId, saleId) is its identity, and that pair
// is stable across reorders because only position changes.
type Stop struct {
	itineraryId string
	saleId      string
	position    int
	status      StopStatus
}

func (s *Stop) ItineraryId() string {
	return s.itineraryId
}

func (s *Stop) SaleId() string {
	return s.saleId
}

func (s *Stop) Position() int {
	return s.position
}

func (s *Stop) Status() StopStatus {
	return s.status
}

type StopStatus string

const (
	StopStatusPlanned StopStatus = "planned"
	StopStatusVisited StopStatus = "visited"
	StopStatusSkipped StopStatus = "skipped"
)
