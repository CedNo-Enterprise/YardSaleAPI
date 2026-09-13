package services

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/address"
	"GarageSaleAPI/domain/itinerary"
	"GarageSaleAPI/domain/sale"
	"GarageSaleAPI/infrastructure/persistence/memory"
	"GarageSaleAPI/interfaces/requests"
	"GarageSaleAPI/test"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

func floatPtr(value float64) *float64 { return &value }

func stringPtr(value string) *string { return &value }

func itinerarySaleIds(stops []itinerary.Stop) []string {
	ids := make([]string, 0, len(stops))
	for _, stop := range stops {
		ids = append(ids, stop.SaleId())
	}

	return ids
}

func seedSale(t *testing.T, repo *memory.InMemorySaleRepository, saleId string, latitude float64, longitude float64) {
	t.Helper()

	saleAddress := address.CreateAddress("northern", nil, "Washington", "WS", "U1A 2C5", "US")
	saleAddress.AddLatLong(latitude, longitude)

	s := sale.CreateSale(
		saleId, uuid.NewString(), "seeded sale",
		saleAddress, time.Now(), "", time.Now(),
	)
	if err := repo.Create(test.CreateTestContext(t), s); err != nil {
		t.Fatalf("seeding sale %q: %v", saleId, err)
	}
}

func newItineraryFixture(t *testing.T) (
	*ItineraryService, *memory.InMemoryItineraryRepository, *memory.InMemorySaleRepository,
) {
	t.Helper()

	itineraryRepo := &memory.InMemoryItineraryRepository{}
	saleRepo := &memory.InMemorySaleRepository{}

	seedSale(t, saleRepo, "near", 45.1, -75.0)
	seedSale(t, saleRepo, "middle", 45.2, -75.0)
	seedSale(t, saleRepo, "far", 45.3, -75.0)

	return NewItineraryService(itineraryRepo, saleRepo), itineraryRepo, saleRepo
}

func validItineraryRequest() requests.ItineraryRequest {
	return requests.ItineraryRequest{
		Name:           "Saturday run",
		Description:    "east end",
		Date:           time.Now(),
		StartLatitude:  floatPtr(45.0),
		StartLongitude: floatPtr(-75.0),
		SaleIds:        []string{"far", "near", "middle"},
	}
}

func Test_validateItinerary(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*requests.ItineraryRequest)
		wantErr     bool
		wantErrKind apperror.Kind
	}{
		{
			name:    "valid itinerary",
			mutate:  func(r *requests.ItineraryRequest) {},
			wantErr: false,
		},
		{
			name:        "empty name",
			mutate:      func(r *requests.ItineraryRequest) { r.Name = "" },
			wantErr:     true,
			wantErrKind: apperror.KindInvalid,
		},
		{
			name: "name too long",
			mutate: func(r *requests.ItineraryRequest) {
				r.Name = "This itinerary name is far too long for our liking and will be rejected"
			},
			wantErr:     true,
			wantErrKind: apperror.KindInvalid,
		},
		{
			name:        "missing date",
			mutate:      func(r *requests.ItineraryRequest) { r.Date = time.Time{} },
			wantErr:     true,
			wantErrKind: apperror.KindInvalid,
		},
		{
			name:        "latitude out of range",
			mutate:      func(r *requests.ItineraryRequest) { r.StartLatitude = floatPtr(91) },
			wantErr:     true,
			wantErrKind: apperror.KindInvalid,
		},
		{
			name:        "half a start point",
			mutate:      func(r *requests.ItineraryRequest) { r.StartLongitude = nil },
			wantErr:     true,
			wantErrKind: apperror.KindInvalid,
		},
		{
			name:        "duplicate sale ids",
			mutate:      func(r *requests.ItineraryRequest) { r.SaleIds = []string{"near", "near"} },
			wantErr:     true,
			wantErrKind: apperror.KindInvalid,
		},
		{
			name:    "no start point at all",
			mutate:  func(r *requests.ItineraryRequest) { r.StartLatitude, r.StartLongitude = nil, nil },
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itineraryDTO := validItineraryRequest()
			tt.mutate(&itineraryDTO)

			err := validateItinerary(itineraryDTO)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateItinerary() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				test.AssertKind(t, err, tt.wantErrKind)
			}
		})
	}
}

func Test_validateUpdateItinerary(t *testing.T) {
	tests := []struct {
		name        string
		updateDTO   requests.UpdateItineraryRequest
		wantErr     bool
		wantErrKind apperror.Kind
	}{
		{
			name:      "rename only",
			updateDTO: requests.UpdateItineraryRequest{Name: stringPtr("Sunday run")},
			wantErr:   false,
		},
		{
			name:      "clearing the description is a real change",
			updateDTO: requests.UpdateItineraryRequest{Description: stringPtr("")},
			wantErr:   false,
		},
		{
			name:        "empty patch",
			updateDTO:   requests.UpdateItineraryRequest{},
			wantErr:     true,
			wantErrKind: apperror.KindInvalid,
		},
		{
			name:        "half a start point",
			updateDTO:   requests.UpdateItineraryRequest{StartLatitude: floatPtr(45.0)},
			wantErr:     true,
			wantErrKind: apperror.KindInvalid,
		},
		{
			name: "both coordinates together",
			updateDTO: requests.UpdateItineraryRequest{
				StartLatitude:  floatPtr(45.0),
				StartLongitude: floatPtr(-75.0),
			},
			wantErr: false,
		},
		{
			name:        "name too long",
			updateDTO:   requests.UpdateItineraryRequest{Name: stringPtr(string(make([]byte, 65)))},
			wantErr:     true,
			wantErrKind: apperror.KindInvalid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUpdateItinerary(tt.updateDTO)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateUpdateItinerary() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				test.AssertKind(t, err, tt.wantErrKind)
			}
		})
	}
}

func TestItineraryService_AddItinerary(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*requests.ItineraryRequest)
		wantErr     bool
		wantErrKind apperror.Kind
		wantOrder   []string
	}{
		{
			name:      "orders stops by proximity to the start",
			mutate:    func(r *requests.ItineraryRequest) {},
			wantErr:   false,
			wantOrder: []string{"near", "middle", "far"},
		},
		{
			name:        "unknown sale id",
			mutate:      func(r *requests.ItineraryRequest) { r.SaleIds = []string{"near", "nonexistent"} },
			wantErr:     true,
			wantErrKind: apperror.KindInvalid,
		},
		{
			name:      "no sales yields an empty route",
			mutate:    func(r *requests.ItineraryRequest) { r.SaleIds = nil },
			wantErr:   false,
			wantOrder: []string{},
		},
		{
			name: "without a start point the submitted order is kept",
			mutate: func(r *requests.ItineraryRequest) {
				r.StartLatitude, r.StartLongitude = nil, nil
			},
			wantErr:   false,
			wantOrder: []string{"far", "near", "middle"},
		},
		{
			name:        "invalid itinerary",
			mutate:      func(r *requests.ItineraryRequest) { r.Name = "" },
			wantErr:     true,
			wantErrKind: apperror.KindInvalid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, itineraryRepo, _ := newItineraryFixture(t)
			ctx := test.CreateTestContext(t)
			userId := uuid.NewString()

			itineraryDTO := validItineraryRequest()
			tt.mutate(&itineraryDTO)

			itineraryId, err := service.AddItinerary(ctx, userId, itineraryDTO)

			if (err != nil) != tt.wantErr {
				t.Fatalf("AddItinerary() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				test.AssertKind(t, err, tt.wantErrKind)
				return
			}

			stored, err := itineraryRepo.GetById(ctx, *itineraryId)
			if err != nil {
				t.Fatalf("GetById() error = %v", err)
			}
			if got := itinerarySaleIds(stored.Stops()); !reflect.DeepEqual(got, tt.wantOrder) {
				t.Errorf("stops = %v, want %v", got, tt.wantOrder)
			}
			if stored.UserId() != userId {
				t.Errorf("UserId() = %q, want %q", stored.UserId(), userId)
			}
			for index, stop := range stored.Stops() {
				if stop.Position() != index {
					t.Errorf("stop %q position = %d, want %d", stop.SaleId(), stop.Position(), index)
				}
			}
		})
	}
}

func TestItineraryService_rejectsNonOwners(t *testing.T) {
	service, _, _ := newItineraryFixture(t)
	ctx := test.CreateTestContext(t)

	ownerId := uuid.NewString()
	intruderId := uuid.NewString()

	itineraryId, err := service.AddItinerary(ctx, ownerId, validItineraryRequest())
	if err != nil {
		t.Fatalf("AddItinerary() error = %v", err)
	}

	tests := []struct {
		name string
		call func() error
	}{
		{
			name: "update",
			call: func() error {
				_, err := service.UpdateItinerary(ctx, *itineraryId, intruderId,
					requests.UpdateItineraryRequest{Name: stringPtr("stolen")})
				return err
			},
		},
		{
			name: "delete",
			call: func() error { return service.DeleteItinerary(ctx, *itineraryId, intruderId) },
		},
		{
			name: "add stop",
			call: func() error {
				_, err := service.AddStop(ctx, *itineraryId, intruderId,
					requests.AddStopRequest{SaleId: "near"})
				return err
			},
		},
		{
			name: "remove stop",
			call: func() error { return service.RemoveStop(ctx, *itineraryId, intruderId, "near") },
		},
		{
			name: "reorder stops",
			call: func() error {
				_, err := service.ReorderStops(ctx, *itineraryId, intruderId,
					requests.ReorderStopsRequest{})
				return err
			},
		},
		{
			name: "set stop status",
			call: func() error {
				_, err := service.SetStopStatus(ctx, *itineraryId, intruderId, "near",
					requests.UpdateStopStatusRequest{Status: "visited"})
				return err
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.AssertKind(t, tt.call(), apperror.KindForbidden)
		})
	}
}

func TestItineraryService_UpdateItinerary(t *testing.T) {
	service, _, _ := newItineraryFixture(t)
	ctx := test.CreateTestContext(t)
	userId := uuid.NewString()

	itineraryId, err := service.AddItinerary(ctx, userId, validItineraryRequest())
	if err != nil {
		t.Fatalf("AddItinerary() error = %v", err)
	}

	updated, err := service.UpdateItinerary(ctx, *itineraryId, userId,
		requests.UpdateItineraryRequest{Name: stringPtr("Sunday run")})
	if err != nil {
		t.Fatalf("UpdateItinerary() error = %v", err)
	}

	if updated.Name() != "Sunday run" {
		t.Errorf("Name() = %q, want %q", updated.Name(), "Sunday run")
	}
	if updated.Description() != "east end" {
		t.Errorf("Description() = %q, want %q", updated.Description(), "east end")
	}
	if len(updated.Stops()) != 3 {
		t.Errorf("len(Stops()) = %d, want 3", len(updated.Stops()))
	}

	_, err = service.UpdateItinerary(ctx, uuid.NewString(), userId,
		requests.UpdateItineraryRequest{Name: stringPtr("ghost")})
	test.AssertKind(t, err, apperror.KindNotFound)
}

func TestItineraryService_AddStop(t *testing.T) {
	service, _, saleRepo := newItineraryFixture(t)
	ctx := test.CreateTestContext(t)
	userId := uuid.NewString()

	itineraryDTO := validItineraryRequest()
	itineraryDTO.SaleIds = []string{"near"}
	itineraryId, err := service.AddItinerary(ctx, userId, itineraryDTO)
	if err != nil {
		t.Fatalf("AddItinerary() error = %v", err)
	}

	seedSale(t, saleRepo, "extra", 45.4, -75.0)

	got, err := service.AddStop(ctx, *itineraryId, userId, requests.AddStopRequest{SaleId: "extra"})
	if err != nil {
		t.Fatalf("AddStop() error = %v", err)
	}
	if want := []string{"near", "extra"}; !reflect.DeepEqual(itinerarySaleIds(got.Stops()), want) {
		t.Errorf("stops = %v, want %v", itinerarySaleIds(got.Stops()), want)
	}

	_, err = service.AddStop(ctx, *itineraryId, userId, requests.AddStopRequest{SaleId: "near"})
	test.AssertKind(t, err, apperror.KindConflict)

	_, err = service.AddStop(ctx, *itineraryId, userId, requests.AddStopRequest{SaleId: "nonexistent"})
	test.AssertKind(t, err, apperror.KindInvalid)

	_, err = service.AddStop(ctx, uuid.NewString(), userId, requests.AddStopRequest{SaleId: "far"})
	test.AssertKind(t, err, apperror.KindNotFound)
}

func TestItineraryService_RemoveStop(t *testing.T) {
	service, _, _ := newItineraryFixture(t)
	ctx := test.CreateTestContext(t)
	userId := uuid.NewString()

	itineraryId, err := service.AddItinerary(ctx, userId, validItineraryRequest())
	if err != nil {
		t.Fatalf("AddItinerary() error = %v", err)
	}

	if err = service.RemoveStop(ctx, *itineraryId, userId, "middle"); err != nil {
		t.Fatalf("RemoveStop() error = %v", err)
	}

	got, err := service.GetItineraryById(ctx, *itineraryId)
	if err != nil {
		t.Fatalf("GetItineraryById() error = %v", err)
	}
	if want := []string{"near", "far"}; !reflect.DeepEqual(itinerarySaleIds(got.Stops()), want) {
		t.Errorf("stops = %v, want %v", itinerarySaleIds(got.Stops()), want)
	}
	for index, stop := range got.Stops() {
		if stop.Position() != index {
			t.Errorf("stop %q position = %d, want %d", stop.SaleId(), stop.Position(), index)
		}
	}

	test.AssertKind(t, service.RemoveStop(ctx, *itineraryId, userId, "middle"), apperror.KindNotFound)
}

func TestItineraryService_ReorderStops_preservesStopStatus(t *testing.T) {
	service, _, _ := newItineraryFixture(t)
	ctx := test.CreateTestContext(t)
	userId := uuid.NewString()

	itineraryDTO := validItineraryRequest()
	itineraryDTO.StartLatitude, itineraryDTO.StartLongitude = nil, nil
	itineraryId, err := service.AddItinerary(ctx, userId, itineraryDTO)
	if err != nil {
		t.Fatalf("AddItinerary() error = %v", err)
	}

	if _, err = service.SetStopStatus(ctx, *itineraryId, userId, "far",
		requests.UpdateStopStatusRequest{Status: "visited"}); err != nil {
		t.Fatalf("SetStopStatus() error = %v", err)
	}

	reordered, err := service.ReorderStops(ctx, *itineraryId, userId, requests.ReorderStopsRequest{
		StartLatitude:  floatPtr(45.0),
		StartLongitude: floatPtr(-75.0),
	})
	if err != nil {
		t.Fatalf("ReorderStops() error = %v", err)
	}

	if want := []string{"near", "middle", "far"}; !reflect.DeepEqual(itinerarySaleIds(reordered.Stops()), want) {
		t.Errorf("stops = %v, want %v", itinerarySaleIds(reordered.Stops()), want)
	}
	if reordered.StartLatitude() != 45.0 || reordered.StartLongitude() != -75.0 {
		t.Errorf("start = (%v, %v), want (45, -75)",
			reordered.StartLatitude(), reordered.StartLongitude())
	}

	stored, err := service.GetItineraryById(ctx, *itineraryId)
	if err != nil {
		t.Fatalf("GetItineraryById() error = %v", err)
	}
	for _, stop := range stored.Stops() {
		if stop.SaleId() == "far" && stop.Status() != itinerary.StopStatusVisited {
			t.Errorf("reorder lost stop status: got %v, want %v",
				stop.Status(), itinerary.StopStatusVisited)
		}
	}
}

func TestItineraryService_ReorderStops_reusesStoredStart(t *testing.T) {
	service, _, _ := newItineraryFixture(t)
	ctx := test.CreateTestContext(t)
	userId := uuid.NewString()

	itineraryId, err := service.AddItinerary(ctx, userId, validItineraryRequest())
	if err != nil {
		t.Fatalf("AddItinerary() error = %v", err)
	}

	reordered, err := service.ReorderStops(ctx, *itineraryId, userId, requests.ReorderStopsRequest{})
	if err != nil {
		t.Fatalf("ReorderStops() error = %v", err)
	}

	if reordered.StartLatitude() != 45.0 || reordered.StartLongitude() != -75.0 {
		t.Errorf("start = (%v, %v), want the stored (45, -75)",
			reordered.StartLatitude(), reordered.StartLongitude())
	}
	if want := []string{"near", "middle", "far"}; !reflect.DeepEqual(itinerarySaleIds(reordered.Stops()), want) {
		t.Errorf("stops = %v, want %v", itinerarySaleIds(reordered.Stops()), want)
	}
}

func TestItineraryService_SetStopStatus(t *testing.T) {
	service, _, _ := newItineraryFixture(t)
	ctx := test.CreateTestContext(t)
	userId := uuid.NewString()

	itineraryId, err := service.AddItinerary(ctx, userId, validItineraryRequest())
	if err != nil {
		t.Fatalf("AddItinerary() error = %v", err)
	}

	_, err = service.SetStopStatus(ctx, *itineraryId, userId, "near",
		requests.UpdateStopStatusRequest{Status: "skipped"})
	if err != nil {
		t.Fatalf("SetStopStatus() error = %v", err)
	}

	stored, err := service.GetItineraryById(ctx, *itineraryId)
	if err != nil {
		t.Fatalf("GetItineraryById() error = %v", err)
	}
	if stored.Stops()[0].Status() != itinerary.StopStatusSkipped {
		t.Errorf("status = %v, want %v", stored.Stops()[0].Status(), itinerary.StopStatusSkipped)
	}

	_, err = service.SetStopStatus(ctx, *itineraryId, userId, "nonexistent",
		requests.UpdateStopStatusRequest{Status: "visited"})
	test.AssertKind(t, err, apperror.KindNotFound)

	_, err = service.SetStopStatus(ctx, *itineraryId, userId, "near",
		requests.UpdateStopStatusRequest{Status: "loitering"})
	test.AssertKind(t, err, apperror.KindInvalid)
}

func TestItineraryService_DeleteItinerary(t *testing.T) {
	service, _, _ := newItineraryFixture(t)
	ctx := test.CreateTestContext(t)
	userId := uuid.NewString()

	itineraryId, err := service.AddItinerary(ctx, userId, validItineraryRequest())
	if err != nil {
		t.Fatalf("AddItinerary() error = %v", err)
	}

	if err = service.DeleteItinerary(ctx, *itineraryId, userId); err != nil {
		t.Fatalf("DeleteItinerary() error = %v", err)
	}

	_, err = service.GetItineraryById(ctx, *itineraryId)
	test.AssertKind(t, err, apperror.KindNotFound)

	test.AssertKind(t, service.DeleteItinerary(ctx, *itineraryId, userId), apperror.KindNotFound)
}

func TestItineraryService_GetItinerariesByUserId(t *testing.T) {
	service, _, _ := newItineraryFixture(t)
	ctx := test.CreateTestContext(t)

	userId := uuid.NewString()
	otherUserId := uuid.NewString()

	if _, err := service.AddItinerary(ctx, userId, validItineraryRequest()); err != nil {
		t.Fatalf("AddItinerary() error = %v", err)
	}
	if _, err := service.AddItinerary(ctx, userId, validItineraryRequest()); err != nil {
		t.Fatalf("AddItinerary() error = %v", err)
	}
	if _, err := service.AddItinerary(ctx, otherUserId, validItineraryRequest()); err != nil {
		t.Fatalf("AddItinerary() error = %v", err)
	}

	got, err := service.GetItinerariesByUserId(ctx, userId)
	if err != nil {
		t.Fatalf("GetItinerariesByUserId() error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d itineraries, want 2", len(got))
	}

	none, err := service.GetItinerariesByUserId(ctx, uuid.NewString())
	if err != nil {
		t.Fatalf("GetItinerariesByUserId() error = %v", err)
	}
	if none == nil || len(none) != 0 {
		t.Errorf("got %v, want an empty slice", none)
	}
}
