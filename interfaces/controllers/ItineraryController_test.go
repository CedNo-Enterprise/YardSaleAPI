package controllers

import (
	"GarageSaleAPI/application/services"
	"GarageSaleAPI/domain/address"
	"GarageSaleAPI/domain/sale"
	"GarageSaleAPI/infrastructure/persistence/memory"
	"GarageSaleAPI/interfaces"
	"GarageSaleAPI/interfaces/requests"
	"GarageSaleAPI/interfaces/responses"
	"GarageSaleAPI/test"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

const (
	nearSale   = "aaaaaaaa-0000-4000-8000-000000000001"
	middleSale = "aaaaaaaa-0000-4000-8000-000000000002"
	farSale    = "aaaaaaaa-0000-4000-8000-000000000003"
	extraSale  = "aaaaaaaa-0000-4000-8000-000000000004"
	absentSale = "aaaaaaaa-0000-4000-8000-00000000dead"
)

func itineraryRequest(method string, target string, body string, pathValues map[string]string) *http.Request {
	var reader io.Reader
	if body != "" {
		reader = bytes.NewBufferString(body)
	}

	r := httptest.NewRequest(method, target, reader)
	if body != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	for key, value := range pathValues {
		r.SetPathValue(key, value)
	}

	return r
}

func newItineraryController(t *testing.T) (*ItineraryController, *services.ItineraryService) {
	t.Helper()

	itineraryRepo := &memory.InMemoryItineraryRepository{}
	saleRepo := &memory.InMemorySaleRepository{}

	for saleId, latitude := range map[string]float64{nearSale: 45.1, middleSale: 45.2, farSale: 45.3} {
		saleAddress := address.CreateAddress("northern", nil, "Washington", "WS", "U1A 2C5", "US")
		saleAddress.AddLatLong(latitude, -75.0)
		s := sale.CreateSale(
			saleId, uuid.NewString(), "seeded sale",
			saleAddress, time.Now(), "", time.Now(),
		)
		if err := saleRepo.Create(test.CreateTestContext(t), s); err != nil {
			t.Fatalf("seeding sale %q: %v", saleId, err)
		}
	}

	service := services.NewItineraryService(itineraryRepo, saleRepo)
	tokenService := services.NewTokenService([]byte("f81d4fae-7dec-11d0-a765-00a0c91e6bf6"), 24*time.Hour)
	authMiddleware := interfaces.NewAuthenticationMiddleware(
		tokenService, services.NewSessionService(&memory.InMemoryRevokedTokenRepository{}),
	)

	return NewItineraryController(service, authMiddleware), service
}

func seedItinerary(t *testing.T, service *services.ItineraryService, userId string) string {
	t.Helper()

	itineraryId, err := service.AddItinerary(test.CreateTestContext(t), userId, requests.ItineraryRequest{
		Name:           "Saturday run",
		Description:    "east end",
		Date:           time.Now(),
		StartLatitude:  &[]float64{45.0}[0],
		StartLongitude: &[]float64{-75.0}[0],
		SaleIds:        []string{farSale, nearSale, middleSale},
	})
	if err != nil {
		t.Fatalf("seeding itinerary: %v", err)
	}

	return *itineraryId
}

func decodeItinerary(t *testing.T, w *httptest.ResponseRecorder) responses.ItineraryResponse {
	t.Helper()

	var response responses.ItineraryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("decoding response %q: %v", w.Body.String(), err)
	}

	return response
}

func TestItineraryController_addItinerary(t *testing.T) {
	validBody := fmt.Sprintf(`{
		"name": "Saturday run",
		"date": "2026-09-19T08:00:00Z",
		"startLatitude": 45.0,
		"startLongitude": -75.0,
		"saleIds": [%q, %q, %q]
	}`, farSale, nearSale, middleSale)

	tests := []struct {
		name           string
		body           string
		contentType    string
		wantStatusCode int
	}{
		{
			name:           "add valid itinerary",
			body:           validBody,
			contentType:    "application/json",
			wantStatusCode: http.StatusCreated,
		},
		{
			name:           "wrong content type",
			body:           validBody,
			contentType:    "",
			wantStatusCode: http.StatusUnsupportedMediaType,
		},
		{
			name:           "malformed body",
			body:           `{"name": `,
			contentType:    "application/json",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "unknown field is rejected",
			body:           `{"name": "run", "date": "2026-09-19T08:00:00Z", "startLat": 45.0}`,
			contentType:    "application/json",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "missing name",
			body:           `{"date": "2026-09-19T08:00:00Z"}`,
			contentType:    "application/json",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "unknown sale",
			body:           `{"name": "run", "date": "2026-09-19T08:00:00Z", "saleIds": ["ghost"]}`,
			contentType:    "application/json",
			wantStatusCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, _ := newItineraryController(t)
			w := httptest.NewRecorder()
			r := test.CreateRequest(http.MethodPost, "/itinerary", bytes.NewBufferString(tt.body), tt.contentType)

			controller.addItinerary(w, r, uuid.NewString())

			if w.Code != tt.wantStatusCode {
				t.Errorf("addItinerary() status = %d, want %d (body %q)", w.Code, tt.wantStatusCode, w.Body.String())
			}
			if tt.wantStatusCode == http.StatusCreated && w.Header().Get("Location") == "" {
				t.Errorf("addItinerary() did not set a Location header")
			}
		})
	}
}

func TestItineraryController_getItinerary(t *testing.T) {
	controller, service := newItineraryController(t)
	userId := uuid.NewString()
	itineraryId := seedItinerary(t, service, userId)

	t.Run("get existing itinerary without authenticating", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := itineraryRequest(http.MethodGet, "/itinerary/", "", map[string]string{"id": itineraryId})

		controller.getItinerary(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("getItinerary() status = %d, want %d", w.Code, http.StatusOK)
		}

		response := decodeItinerary(t, w)
		if response.Id != itineraryId {
			t.Errorf("id = %q, want %q", response.Id, itineraryId)
		}
		if len(response.Stops) != 3 {
			t.Fatalf("len(stops) = %d, want 3", len(response.Stops))
		}
		if response.Stops[0].SaleId != nearSale {
			t.Errorf("first stop = %q, want %q", response.Stops[0].SaleId, nearSale)
		}
		if response.Stops[0].Status != "planned" {
			t.Errorf("first stop status = %q, want %q", response.Stops[0].Status, "planned")
		}
	})

	t.Run("get nonexistent itinerary", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := itineraryRequest(http.MethodGet, "/itinerary/", "", map[string]string{"id": uuid.NewString()})

		controller.getItinerary(w, r)

		if w.Code != http.StatusNotFound {
			t.Errorf("getItinerary() status = %d, want %d", w.Code, http.StatusNotFound)
		}
		if w.Body.String() != "itinerary not found\n" {
			t.Errorf("body = %q, want %q", w.Body.String(), "itinerary not found\n")
		}
	})
}

func TestItineraryController_listItineraries(t *testing.T) {
	controller, service := newItineraryController(t)
	userId := uuid.NewString()
	seedItinerary(t, service, userId)
	seedItinerary(t, service, uuid.NewString())

	w := httptest.NewRecorder()
	r := itineraryRequest(http.MethodGet, "/itinerary", "", nil)

	controller.listItineraries(w, r, userId)

	if w.Code != http.StatusOK {
		t.Fatalf("listItineraries() status = %d, want %d", w.Code, http.StatusOK)
	}

	var response []responses.ItineraryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("decoding response %q: %v", w.Body.String(), err)
	}
	if len(response) != 1 {
		t.Errorf("returned %d itineraries, want only the caller's 1", len(response))
	}
}

func TestItineraryController_updateItinerary(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		asOwner        bool
		wantStatusCode int
	}{
		{
			name:           "rename",
			body:           `{"name": "Sunday run"}`,
			asOwner:        true,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "empty patch",
			body:           `{}`,
			asOwner:        true,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "half a start point",
			body:           `{"startLatitude": 45.0}`,
			asOwner:        true,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "not the owner",
			body:           `{"name": "stolen"}`,
			asOwner:        false,
			wantStatusCode: http.StatusForbidden,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, service := newItineraryController(t)
			ownerId := uuid.NewString()
			itineraryId := seedItinerary(t, service, ownerId)

			callerId := ownerId
			if !tt.asOwner {
				callerId = uuid.NewString()
			}

			w := httptest.NewRecorder()
			r := itineraryRequest(http.MethodPatch, "/itinerary/", tt.body, map[string]string{"id": itineraryId})

			controller.updateItinerary(w, r, callerId)

			if w.Code != tt.wantStatusCode {
				t.Errorf("updateItinerary() status = %d, want %d (body %q)",
					w.Code, tt.wantStatusCode, w.Body.String())
			}
		})
	}
}

func TestItineraryController_deleteItinerary(t *testing.T) {
	tests := []struct {
		name           string
		asOwner        bool
		wantStatusCode int
	}{
		{name: "owner deletes", asOwner: true, wantStatusCode: http.StatusNoContent},
		{name: "not the owner", asOwner: false, wantStatusCode: http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, service := newItineraryController(t)
			ownerId := uuid.NewString()
			itineraryId := seedItinerary(t, service, ownerId)

			callerId := ownerId
			if !tt.asOwner {
				callerId = uuid.NewString()
			}

			w := httptest.NewRecorder()
			r := itineraryRequest(http.MethodDelete, "/itinerary/", "", map[string]string{"id": itineraryId})

			controller.deleteItinerary(w, r, callerId)

			if w.Code != tt.wantStatusCode {
				t.Errorf("deleteItinerary() status = %d, want %d", w.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestItineraryController_stopHandlers(t *testing.T) {
	t.Run("add stop", func(t *testing.T) {
		controller, service := newItineraryController(t)
		ownerId := uuid.NewString()
		itineraryId := seedItinerary(t, service, ownerId)

		if err := service.RemoveStop(test.CreateTestContext(t), itineraryId, ownerId, farSale); err != nil {
			t.Fatalf("RemoveStop() error = %v", err)
		}

		w := httptest.NewRecorder()
		r := itineraryRequest(http.MethodPost, "/itinerary/stop", fmt.Sprintf(`{"saleId": %q}`, farSale),
			map[string]string{"id": itineraryId})

		controller.addStop(w, r, ownerId)

		if w.Code != http.StatusOK {
			t.Fatalf("addStop() status = %d, want %d (body %q)", w.Code, http.StatusOK, w.Body.String())
		}
		if response := decodeItinerary(t, w); len(response.Stops) != 3 {
			t.Errorf("len(stops) = %d, want 3", len(response.Stops))
		}
	})

	t.Run("add duplicate stop", func(t *testing.T) {
		controller, service := newItineraryController(t)
		ownerId := uuid.NewString()
		itineraryId := seedItinerary(t, service, ownerId)

		w := httptest.NewRecorder()
		r := itineraryRequest(http.MethodPost, "/itinerary/stop", fmt.Sprintf(`{"saleId": %q}`, nearSale),
			map[string]string{"id": itineraryId})

		controller.addStop(w, r, ownerId)

		if w.Code != http.StatusConflict {
			t.Errorf("addStop() status = %d, want %d", w.Code, http.StatusConflict)
		}
	})

	t.Run("update stop status", func(t *testing.T) {
		controller, service := newItineraryController(t)
		ownerId := uuid.NewString()
		itineraryId := seedItinerary(t, service, ownerId)

		w := httptest.NewRecorder()
		r := itineraryRequest(http.MethodPatch, "/itinerary/stop", `{"status": "visited"}`,
			map[string]string{"id": itineraryId, "saleId": nearSale})

		controller.updateStopStatus(w, r, ownerId)

		if w.Code != http.StatusOK {
			t.Fatalf("updateStopStatus() status = %d, want %d (body %q)",
				w.Code, http.StatusOK, w.Body.String())
		}
		if response := decodeItinerary(t, w); response.Stops[0].Status != "visited" {
			t.Errorf("status = %q, want %q", response.Stops[0].Status, "visited")
		}
	})

	t.Run("update stop status with an unknown value", func(t *testing.T) {
		controller, service := newItineraryController(t)
		ownerId := uuid.NewString()
		itineraryId := seedItinerary(t, service, ownerId)

		w := httptest.NewRecorder()
		r := itineraryRequest(http.MethodPatch, "/itinerary/stop", `{"status": "loitering"}`,
			map[string]string{"id": itineraryId, "saleId": nearSale})

		controller.updateStopStatus(w, r, ownerId)

		if w.Code != http.StatusBadRequest {
			t.Errorf("updateStopStatus() status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("remove stop", func(t *testing.T) {
		controller, service := newItineraryController(t)
		ownerId := uuid.NewString()
		itineraryId := seedItinerary(t, service, ownerId)

		w := httptest.NewRecorder()
		r := itineraryRequest(http.MethodDelete, "/itinerary/stop", "",
			map[string]string{"id": itineraryId, "saleId": nearSale})

		controller.removeStop(w, r, ownerId)

		if w.Code != http.StatusNoContent {
			t.Errorf("removeStop() status = %d, want %d", w.Code, http.StatusNoContent)
		}
	})

	t.Run("remove stop as a non-owner", func(t *testing.T) {
		controller, service := newItineraryController(t)
		itineraryId := seedItinerary(t, service, uuid.NewString())

		w := httptest.NewRecorder()
		r := itineraryRequest(http.MethodDelete, "/itinerary/stop", "",
			map[string]string{"id": itineraryId, "saleId": nearSale})

		controller.removeStop(w, r, uuid.NewString())

		if w.Code != http.StatusForbidden {
			t.Errorf("removeStop() status = %d, want %d", w.Code, http.StatusForbidden)
		}
	})
}

func TestItineraryController_reorderStops(t *testing.T) {
	controller, service := newItineraryController(t)
	ownerId := uuid.NewString()
	itineraryId := seedItinerary(t, service, ownerId)

	w := httptest.NewRecorder()
	r := itineraryRequest(http.MethodPut, "/itinerary/stops/order",
		`{"startLatitude": 45.5, "startLongitude": -75.0}`,
		map[string]string{"id": itineraryId})

	controller.reorderStops(w, r, ownerId)

	if w.Code != http.StatusOK {
		t.Fatalf("reorderStops() status = %d, want %d (body %q)", w.Code, http.StatusOK, w.Body.String())
	}

	response := decodeItinerary(t, w)
	if want := []string{farSale, middleSale, nearSale}; len(response.Stops) != 3 ||
		response.Stops[0].SaleId != want[0] ||
		response.Stops[1].SaleId != want[1] ||
		response.Stops[2].SaleId != want[2] {
		t.Errorf("stops = %+v, want %v", response.Stops, want)
	}
}
