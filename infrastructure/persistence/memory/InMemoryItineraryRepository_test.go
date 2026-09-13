package memory

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/itinerary"
	"GarageSaleAPI/test"
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

var _ itinerary.ItineraryRepository = (*InMemoryItineraryRepository)(nil)

func newTestItinerary(id string, userId string, saleIds ...string) *itinerary.Itinerary {
	i := itinerary.CreateItinerary(
		id, userId, "Saturday run", "east end",
		time.Now(), 45.0, -75.0, time.Now(),
	)
	for _, saleId := range saleIds {
		i.AppendStop(saleId)
	}

	return i
}

func saleIdsOf(stops []itinerary.Stop) []string {
	ids := make([]string, 0, len(stops))
	for _, stop := range stops {
		ids = append(ids, stop.SaleId())
	}

	return ids
}

func TestInMemoryItineraryRepository_Create(t *testing.T) {
	existingId := uuid.NewString()
	existing := newTestItinerary(existingId, uuid.NewString())

	tests := []struct {
		name        string
		stored      []itinerary.Itinerary
		ctx         context.Context
		wantErr     bool
		wantErrKind apperror.Kind
	}{
		{
			name:    "create itinerary",
			stored:  []itinerary.Itinerary{},
			ctx:     test.CreateTestContext(t),
			wantErr: false,
		},
		{
			name:        "create duplicate itinerary",
			stored:      []itinerary.Itinerary{*existing},
			ctx:         test.CreateTestContext(t),
			wantErr:     true,
			wantErrKind: apperror.KindConflict,
		},
		{
			name:        "create with cancelled context",
			stored:      []itinerary.Itinerary{},
			ctx:         test.CreateCancelledTestContext(),
			wantErr:     true,
			wantErrKind: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &InMemoryItineraryRepository{itineraries: tt.stored}

			err := repo.Create(tt.ctx, existing)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErrKind != "" {
				test.AssertKind(t, err, tt.wantErrKind)
			}
		})
	}
}

func TestInMemoryItineraryRepository_GetById(t *testing.T) {
	id := uuid.NewString()
	stored := newTestItinerary(id, uuid.NewString(), "sale-a", "sale-b")
	repo := &InMemoryItineraryRepository{itineraries: []itinerary.Itinerary{*stored}}
	ctx := test.CreateTestContext(t)

	got, err := repo.GetById(ctx, id)
	if err != nil {
		t.Fatalf("GetById() error = %v", err)
	}
	if got.Id() != id {
		t.Errorf("GetById() id = %q, want %q", got.Id(), id)
	}
	if want := []string{"sale-a", "sale-b"}; !reflect.DeepEqual(saleIdsOf(got.Stops()), want) {
		t.Errorf("GetById() stops = %v, want %v", saleIdsOf(got.Stops()), want)
	}

	_, err = repo.GetById(ctx, uuid.NewString())
	test.AssertKind(t, err, apperror.KindNotFound)
}

func TestInMemoryItineraryRepository_GetById_returnsDetachedCopy(t *testing.T) {
	id := uuid.NewString()
	repo := &InMemoryItineraryRepository{
		itineraries: []itinerary.Itinerary{*newTestItinerary(id, uuid.NewString(), "sale-a")},
	}
	ctx := test.CreateTestContext(t)

	got, err := repo.GetById(ctx, id)
	if err != nil {
		t.Fatalf("GetById() error = %v", err)
	}
	got.SetStopStatus("sale-a", itinerary.StopStatusVisited)

	stored, err := repo.GetById(ctx, id)
	if err != nil {
		t.Fatalf("GetById() error = %v", err)
	}
	if stored.Stops()[0].Status() != itinerary.StopStatusPlanned {
		t.Errorf("mutating a returned itinerary changed the store: status = %v, want %v",
			stored.Stops()[0].Status(), itinerary.StopStatusPlanned)
	}
}

func TestInMemoryItineraryRepository_GetByUserId(t *testing.T) {
	userId := uuid.NewString()
	repo := &InMemoryItineraryRepository{
		itineraries: []itinerary.Itinerary{
			*newTestItinerary(uuid.NewString(), userId),
			*newTestItinerary(uuid.NewString(), uuid.NewString()),
			*newTestItinerary(uuid.NewString(), userId),
		},
	}
	ctx := test.CreateTestContext(t)

	got, err := repo.GetByUserId(ctx, userId)
	if err != nil {
		t.Fatalf("GetByUserId() error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("GetByUserId() returned %d itineraries, want 2", len(got))
	}

	none, err := repo.GetByUserId(ctx, uuid.NewString())
	if err != nil {
		t.Fatalf("GetByUserId() error = %v", err)
	}
	if none == nil || len(none) != 0 {
		t.Errorf("GetByUserId() = %v, want an empty slice", none)
	}
}

func TestInMemoryItineraryRepository_stopOperations(t *testing.T) {
	id := uuid.NewString()
	repo := &InMemoryItineraryRepository{
		itineraries: []itinerary.Itinerary{*newTestItinerary(id, uuid.NewString(), "sale-a", "sale-b")},
	}
	ctx := test.CreateTestContext(t)

	added := itinerary.CreateStop(id, "sale-c", 2)
	if err := repo.AddStop(ctx, &added); err != nil {
		t.Fatalf("AddStop() error = %v", err)
	}
	duplicate := itinerary.CreateStop(id, "sale-a", 3)
	test.AssertKind(t, repo.AddStop(ctx, &duplicate), apperror.KindConflict)

	if err := repo.UpdateStopStatus(ctx, id, "sale-c", itinerary.StopStatusVisited); err != nil {
		t.Fatalf("UpdateStopStatus() error = %v", err)
	}
	test.AssertKind(t, repo.UpdateStopStatus(ctx, id, "sale-z", itinerary.StopStatusVisited), apperror.KindNotFound)

	if err := repo.RemoveStop(ctx, id, "sale-a"); err != nil {
		t.Fatalf("RemoveStop() error = %v", err)
	}
	test.AssertKind(t, repo.RemoveStop(ctx, id, "sale-a"), apperror.KindNotFound)

	got, err := repo.GetById(ctx, id)
	if err != nil {
		t.Fatalf("GetById() error = %v", err)
	}
	if want := []string{"sale-b", "sale-c"}; !reflect.DeepEqual(saleIdsOf(got.Stops()), want) {
		t.Errorf("stops = %v, want %v", saleIdsOf(got.Stops()), want)
	}
	for index, stop := range got.Stops() {
		if stop.Position() != index {
			t.Errorf("stop %q position = %d, want %d", stop.SaleId(), stop.Position(), index)
		}
	}
	if got.Stops()[1].Status() != itinerary.StopStatusVisited {
		t.Errorf("sale-c status = %v, want %v", got.Stops()[1].Status(), itinerary.StopStatusVisited)
	}
}

func TestInMemoryItineraryRepository_Update(t *testing.T) {
	id := uuid.NewString()
	userId := uuid.NewString()
	repo := &InMemoryItineraryRepository{
		itineraries: []itinerary.Itinerary{*newTestItinerary(id, userId, "sale-a")},
	}
	ctx := test.CreateTestContext(t)

	changed := newTestItinerary(id, userId)
	changed.Rename("Sunday run")

	if err := repo.Update(ctx, changed); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := repo.GetById(ctx, id)
	if err != nil {
		t.Fatalf("GetById() error = %v", err)
	}
	if got.Name() != "Sunday run" {
		t.Errorf("Name() = %q, want %q", got.Name(), "Sunday run")
	}
	if len(got.Stops()) != 1 {
		t.Errorf("Update() changed the stops: got %d, want 1", len(got.Stops()))
	}

	test.AssertKind(t, repo.Update(ctx, newTestItinerary(uuid.NewString(), userId)), apperror.KindNotFound)
}

func TestInMemoryItineraryRepository_Delete(t *testing.T) {
	id := uuid.NewString()
	repo := &InMemoryItineraryRepository{
		itineraries: []itinerary.Itinerary{*newTestItinerary(id, uuid.NewString())},
	}
	ctx := test.CreateTestContext(t)

	if err := repo.Delete(ctx, id); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if len(repo.itineraries) != 0 {
		t.Errorf("len(itineraries) = %d, want 0", len(repo.itineraries))
	}

	test.AssertKind(t, repo.Delete(ctx, id), apperror.KindNotFound)
}

func TestInMemoryItineraryRepository_ReorderStops_preservesStopStatus(t *testing.T) {
	id := uuid.NewString()
	repo := &InMemoryItineraryRepository{
		itineraries: []itinerary.Itinerary{*newTestItinerary(id, uuid.NewString(), "far", "near")},
	}
	ctx := test.CreateTestContext(t)

	if err := repo.UpdateStopStatus(ctx, id, "far", itinerary.StopStatusVisited); err != nil {
		t.Fatalf("UpdateStopStatus() error = %v", err)
	}

	loaded, err := repo.GetById(ctx, id)
	if err != nil {
		t.Fatalf("GetById() error = %v", err)
	}
	loaded.ReorderFrom(
		itinerary.GeoPoint{Latitude: 45.0, Longitude: -75.0},
		map[string]itinerary.GeoPoint{
			"near": {Latitude: 45.1, Longitude: -75.0},
			"far":  {Latitude: 45.3, Longitude: -75.0},
		},
	)

	if err := repo.ReorderStops(ctx, loaded); err != nil {
		t.Fatalf("ReorderStops() error = %v", err)
	}

	got, err := repo.GetById(ctx, id)
	if err != nil {
		t.Fatalf("GetById() error = %v", err)
	}
	if want := []string{"near", "far"}; !reflect.DeepEqual(saleIdsOf(got.Stops()), want) {
		t.Errorf("stops = %v, want %v", saleIdsOf(got.Stops()), want)
	}
	if got.Stops()[1].Status() != itinerary.StopStatusVisited {
		t.Errorf("reorder lost stop status: got %v, want %v",
			got.Stops()[1].Status(), itinerary.StopStatusVisited)
	}
}
