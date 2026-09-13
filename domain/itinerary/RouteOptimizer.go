package itinerary

import "math"

const earthRadiusKm = 6371.0

type GeoPoint struct {
	Latitude  float64
	Longitude float64
}

// Locatable reports whether the point carries real coordinates. Addresses that
// have never been geocoded are persisted as (0, 0); Null Island is not a garage
// sale, so that pair is read as "location unknown". See AddLatLong in
// domain/address, which is what will eventually populate them.
func (p GeoPoint) Locatable() bool {
	return p.Latitude != 0 || p.Longitude != 0
}

// HaversineKm returns the great-circle distance in kilometres between two points.
// Straight-line distance is a deliberate simplification: we are ranking
// candidates, not producing turn-by-turn directions.
func HaversineKm(a GeoPoint, b GeoPoint) float64 {
	latitude1 := a.Latitude * math.Pi / 180
	latitude2 := b.Latitude * math.Pi / 180
	deltaLatitude := (b.Latitude - a.Latitude) * math.Pi / 180
	deltaLongitude := (b.Longitude - a.Longitude) * math.Pi / 180

	h := math.Sin(deltaLatitude/2)*math.Sin(deltaLatitude/2) +
		math.Cos(latitude1)*math.Cos(latitude2)*
			math.Sin(deltaLongitude/2)*math.Sin(deltaLongitude/2)

	// Clamped so floating point drift on antipodal points cannot produce NaN.
	return 2 * earthRadiusKm * math.Asin(math.Min(1, math.Sqrt(h)))
}

// OrderStopsByProximity returns stops re-sequenced as a nearest-neighbour walk
// from start, with positions renumbered from zero. Stops whose sale has no known
// coordinates keep their submitted relative order and are appended last. The
// input slice is not modified.
func OrderStopsByProximity(stops []Stop, coordinates map[string]GeoPoint, start GeoPoint) []Stop {
	// Without a start point there is no basis for "nearest", so leave the route
	// as submitted rather than inventing an order.
	if !start.Locatable() {
		return renumber(stops)
	}

	locatable := make([]Stop, 0, len(stops))
	unlocatable := make([]Stop, 0, len(stops))
	for _, stop := range stops {
		if coordinates[stop.saleId].Locatable() {
			locatable = append(locatable, stop)
			continue
		}
		unlocatable = append(unlocatable, stop)
	}

	ordered := make([]Stop, 0, len(stops))
	current := start
	for len(locatable) > 0 {
		nearest := 0
		nearestDistance := HaversineKm(current, coordinates[locatable[0].saleId])
		for index := 1; index < len(locatable); index++ {
			distance := HaversineKm(current, coordinates[locatable[index].saleId])
			// Strict comparison keeps the earlier stop on a tie, so the walk is
			// deterministic for equidistant sales.
			if distance < nearestDistance {
				nearest = index
				nearestDistance = distance
			}
		}

		current = coordinates[locatable[nearest].saleId]
		ordered = append(ordered, locatable[nearest])
		locatable = append(locatable[:nearest], locatable[nearest+1:]...)
	}

	return renumber(append(ordered, unlocatable...))
}

func renumber(stops []Stop) []Stop {
	ordered := make([]Stop, len(stops))
	for index, stop := range stops {
		stop.position = index
		ordered[index] = stop
	}

	return ordered
}
