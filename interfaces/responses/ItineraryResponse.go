package responses

import (
	"GarageSaleAPI/domain/itinerary"
	"time"
)

type ItineraryStopResponse struct {
	SaleId   string `json:"saleId"`
	Position int    `json:"position"`
	Status   string `json:"status"`
}

type ItineraryResponse struct {
	Id             string                  `json:"id"`
	Name           string                  `json:"name"`
	Description    string                  `json:"description,omitempty"`
	Date           time.Time               `json:"date"`
	StartLatitude  float64                 `json:"startLatitude"`
	StartLongitude float64                 `json:"startLongitude"`
	Stops          []ItineraryStopResponse `json:"stops"`
}

func NewItineraryResponse(i itinerary.Itinerary) *ItineraryResponse {
	return &ItineraryResponse{
		i.Id(),
		i.Name(),
		i.Description(),
		i.Date(),
		i.StartLatitude(),
		i.StartLongitude(),
		newItineraryStopResponses(i.Stops()),
	}
}

func NewItineraryResponses(itineraries []itinerary.Itinerary) []ItineraryResponse {
	responses := make([]ItineraryResponse, 0, len(itineraries))
	for _, i := range itineraries {
		responses = append(responses, *NewItineraryResponse(i))
	}

	return responses
}

func newItineraryStopResponses(stops []itinerary.Stop) []ItineraryStopResponse {
	responses := make([]ItineraryStopResponse, 0, len(stops))
	for _, stop := range stops {
		responses = append(responses, ItineraryStopResponse{
			stop.SaleId(),
			stop.Position(),
			string(stop.Status()),
		})
	}

	return responses
}
