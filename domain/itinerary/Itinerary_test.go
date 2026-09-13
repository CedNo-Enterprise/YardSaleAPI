package itinerary

import (
	"reflect"
	"testing"
	"time"
)

func newTestItinerary(saleIds ...string) *Itinerary {
	i := CreateItinerary(
		"itinerary", "user", "Saturday run", "east end",
		time.Now(), 45.0, -75.0, time.Now(),
	)
	for _, saleId := range saleIds {
		i.AppendStop(saleId)
	}

	return i
}

func TestItinerary_AppendStop(t *testing.T) {
	i := newTestItinerary()

	first := i.AppendStop("sale-a")
	second := i.AppendStop("sale-b")

	if first.Position() != 0 || second.Position() != 1 {
		t.Errorf("AppendStop() positions = %d, %d, want 0, 1", first.Position(), second.Position())
	}
	if first.Status() != StopStatusPlanned {
		t.Errorf("AppendStop() status = %v, want %v", first.Status(), StopStatusPlanned)
	}
	if first.ItineraryId() != "itinerary" {
		t.Errorf("AppendStop() itineraryId = %q, want %q", first.ItineraryId(), "itinerary")
	}
	if !i.HasStop("sale-a") || !i.HasStop("sale-b") {
		t.Errorf("HasStop() = false for an appended stop")
	}
	if i.HasStop("sale-c") {
		t.Errorf("HasStop() = true for a sale that was never added")
	}
}

func TestItinerary_RemoveStopBySaleId(t *testing.T) {
	tests := []struct {
		name       string
		saleIds    []string
		remove     string
		wantOk     bool
		wantRemain []string
	}{
		{
			name:       "removes and renumbers survivors",
			saleIds:    []string{"a", "b", "c"},
			remove:     "b",
			wantOk:     true,
			wantRemain: []string{"a", "c"},
		},
		{
			name:       "removes the first stop",
			saleIds:    []string{"a", "b", "c"},
			remove:     "a",
			wantOk:     true,
			wantRemain: []string{"b", "c"},
		},
		{
			name:       "unknown sale removes nothing",
			saleIds:    []string{"a", "b"},
			remove:     "z",
			wantOk:     false,
			wantRemain: []string{"a", "b"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := newTestItinerary(tt.saleIds...)

			got := i.RemoveStopBySaleId(tt.remove)

			if got != tt.wantOk {
				t.Errorf("RemoveStopBySaleId() = %v, want %v", got, tt.wantOk)
			}
			if !reflect.DeepEqual(saleIdsOf(i.Stops()), tt.wantRemain) {
				t.Errorf("stops = %v, want %v", saleIdsOf(i.Stops()), tt.wantRemain)
			}
			for index, stop := range i.Stops() {
				if stop.Position() != index {
					t.Errorf("stop %q has position %d, want %d", stop.SaleId(), stop.Position(), index)
				}
			}
		})
	}
}

func TestItinerary_SetStopStatus(t *testing.T) {
	i := newTestItinerary("a", "b")

	if !i.SetStopStatus("b", StopStatusVisited) {
		t.Fatalf("SetStopStatus() = false for an existing stop")
	}
	if i.SetStopStatus("z", StopStatusVisited) {
		t.Errorf("SetStopStatus() = true for an unknown stop")
	}

	stops := i.Stops()
	if stops[1].Status() != StopStatusVisited {
		t.Errorf("stop status = %v, want %v", stops[1].Status(), StopStatusVisited)
	}
	if stops[0].Status() != StopStatusPlanned {
		t.Errorf("unrelated stop status = %v, want %v", stops[0].Status(), StopStatusPlanned)
	}
}

func TestItinerary_ReorderFrom(t *testing.T) {
	i := newTestItinerary("far", "near")
	coordinates := map[string]GeoPoint{
		"near": {Latitude: 45.1, Longitude: -75.0},
		"far":  {Latitude: 45.3, Longitude: -75.0},
	}
	start := GeoPoint{Latitude: 45.0, Longitude: -75.0}

	i.ReorderFrom(start, coordinates)

	if got := saleIdsOf(i.Stops()); !reflect.DeepEqual(got, []string{"near", "far"}) {
		t.Errorf("stops = %v, want [near far]", got)
	}
	if i.StartLatitude() != start.Latitude || i.StartLongitude() != start.Longitude {
		t.Errorf("start = (%v, %v), want (%v, %v)",
			i.StartLatitude(), i.StartLongitude(), start.Latitude, start.Longitude)
	}
}

func TestItinerary_mutators(t *testing.T) {
	i := newTestItinerary()
	date := time.Date(2026, 9, 19, 8, 0, 0, 0, time.UTC)

	i.Rename("Sunday run")
	i.Describe("west end")
	i.Reschedule(date)
	i.SetStart(46.5, -76.5)

	if i.Name() != "Sunday run" {
		t.Errorf("Name() = %q, want %q", i.Name(), "Sunday run")
	}
	if i.Description() != "west end" {
		t.Errorf("Description() = %q, want %q", i.Description(), "west end")
	}
	if !i.Date().Equal(date) {
		t.Errorf("Date() = %v, want %v", i.Date(), date)
	}
	if i.StartLatitude() != 46.5 || i.StartLongitude() != -76.5 {
		t.Errorf("start = (%v, %v), want (46.5, -76.5)", i.StartLatitude(), i.StartLongitude())
	}
}
