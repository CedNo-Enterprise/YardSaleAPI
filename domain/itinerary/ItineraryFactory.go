package itinerary

import "time"

func CreateItinerary(
	id string, userId string, name string, description string, date time.Time,
	startLatitude float64, startLongitude float64, creationTime time.Time,
) *Itinerary {
	return &Itinerary{
		id:             id,
		userId:         userId,
		name:           name,
		description:    description,
		date:           date,
		startLatitude:  startLatitude,
		startLongitude: startLongitude,
		stops:          []Stop{},
		createdAt:      creationTime,
	}
}

func HydrateItinerary(
	id, userId, name, description string, date time.Time,
	startLatitude float64, startLongitude float64, stops []Stop, createdAt time.Time,
) *Itinerary {
	return &Itinerary{
		id:             id,
		userId:         userId,
		name:           name,
		description:    description,
		date:           date,
		startLatitude:  startLatitude,
		startLongitude: startLongitude,
		stops:          stops,
		createdAt:      createdAt,
	}
}

func CreateStop(itineraryId string, saleId string, position int) Stop {
	return Stop{
		itineraryId: itineraryId,
		saleId:      saleId,
		position:    position,
		status:      StopStatusPlanned,
	}
}

func HydrateStop(itineraryId string, saleId string, position int, status StopStatus) Stop {
	return Stop{
		itineraryId: itineraryId,
		saleId:      saleId,
		position:    position,
		status:      status,
	}
}
