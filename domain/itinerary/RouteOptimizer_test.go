package itinerary

import (
	"math"
	"reflect"
	"testing"
)

var (
	ottawa  = GeoPoint{Latitude: 45.4215, Longitude: -75.6972}
	toronto = GeoPoint{Latitude: 43.6532, Longitude: -79.3832}
)

func saleIdsOf(stops []Stop) []string {
	ids := make([]string, 0, len(stops))
	for _, stop := range stops {
		ids = append(ids, stop.saleId)
	}

	return ids
}

func stopsFor(saleIds ...string) []Stop {
	stops := make([]Stop, 0, len(saleIds))
	for index, saleId := range saleIds {
		stops = append(stops, CreateStop("itinerary", saleId, index))
	}

	return stops
}

func TestHaversineKm(t *testing.T) {
	type args struct {
		a GeoPoint
		b GeoPoint
	}
	tests := []struct {
		name        string
		args        args
		want        float64
		toleranceKm float64
	}{
		{
			name:        "same point is zero",
			args:        args{a: ottawa, b: ottawa},
			want:        0,
			toleranceKm: 0.0001,
		},
		{
			name:        "ottawa to toronto",
			args:        args{a: ottawa, b: toronto},
			want:        352,
			toleranceKm: 10,
		},
		{
			name:        "reversed is the same distance",
			args:        args{a: toronto, b: ottawa},
			want:        352,
			toleranceKm: 10,
		},
		{
			name:        "antipodal points are half the circumference",
			args:        args{a: GeoPoint{Latitude: 45, Longitude: 0}, b: GeoPoint{Latitude: -45, Longitude: 180}},
			want:        math.Pi * earthRadiusKm,
			toleranceKm: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HaversineKm(tt.args.a, tt.args.b)

			if math.IsNaN(got) {
				t.Fatalf("HaversineKm() = NaN")
			}
			if math.Abs(got-tt.want) > tt.toleranceKm {
				t.Errorf("HaversineKm() = %v, want %v (+/- %v)", got, tt.want, tt.toleranceKm)
			}
		})
	}
}

func TestGeoPoint_Locatable(t *testing.T) {
	tests := []struct {
		name  string
		point GeoPoint
		want  bool
	}{
		{name: "ungeocoded origin", point: GeoPoint{Latitude: 0, Longitude: 0}, want: false},
		{name: "longitude only", point: GeoPoint{Latitude: 0, Longitude: -75.6972}, want: true},
		{name: "latitude only", point: GeoPoint{Latitude: 45.4215, Longitude: 0}, want: true},
		{name: "both negative", point: GeoPoint{Latitude: -33.86, Longitude: -70.6}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.point.Locatable(); got != tt.want {
				t.Errorf("Locatable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOrderStopsByProximity(t *testing.T) {
	start := GeoPoint{Latitude: 45.0, Longitude: -75.0}

	northOf := map[string]GeoPoint{
		"near":    {Latitude: 45.1, Longitude: -75.0},
		"middle":  {Latitude: 45.2, Longitude: -75.0},
		"far":     {Latitude: 45.3, Longitude: -75.0},
		"nowhere": {Latitude: 0, Longitude: 0},
	}

	type args struct {
		stops       []Stop
		coordinates map[string]GeoPoint
		start       GeoPoint
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "no stops",
			args: args{stops: stopsFor(), coordinates: northOf, start: start},
			want: []string{},
		},
		{
			name: "single stop",
			args: args{stops: stopsFor("far"), coordinates: northOf, start: start},
			want: []string{"far"},
		},
		{
			name: "nearest neighbour walk",
			args: args{stops: stopsFor("far", "near", "middle"), coordinates: northOf, start: start},
			want: []string{"near", "middle", "far"},
		},
		{
			name: "unlocatable stops are appended in submitted order",
			args: args{stops: stopsFor("far", "nowhere", "near"), coordinates: northOf, start: start},
			want: []string{"near", "far", "nowhere"},
		},
		{
			name: "sale missing from the map counts as unlocatable",
			args: args{stops: stopsFor("unknown", "near"), coordinates: northOf, start: start},
			want: []string{"near", "unknown"},
		},
		{
			name: "all stops unlocatable keeps submitted order",
			args: args{stops: stopsFor("nowhere", "unknown"), coordinates: northOf, start: start},
			want: []string{"nowhere", "unknown"},
		},
		{
			name: "unlocatable start skips optimization",
			args: args{stops: stopsFor("far", "near", "middle"), coordinates: northOf, start: GeoPoint{}},
			want: []string{"far", "near", "middle"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			submitted := saleIdsOf(tt.args.stops)

			got := OrderStopsByProximity(tt.args.stops, tt.args.coordinates, tt.args.start)

			if !reflect.DeepEqual(saleIdsOf(got), tt.want) {
				t.Errorf("OrderStopsByProximity() = %v, want %v", saleIdsOf(got), tt.want)
			}
			for index, stop := range got {
				if stop.Position() != index {
					t.Errorf("stop %q has position %d, want %d", stop.SaleId(), stop.Position(), index)
				}
			}
			if !reflect.DeepEqual(saleIdsOf(tt.args.stops), submitted) {
				t.Errorf("OrderStopsByProximity() mutated its input: %v, want %v", saleIdsOf(tt.args.stops), submitted)
			}
		})
	}
}

func TestOrderStopsByProximity_preservesStatus(t *testing.T) {
	start := GeoPoint{Latitude: 45.0, Longitude: -75.0}
	coordinates := map[string]GeoPoint{
		"near": {Latitude: 45.1, Longitude: -75.0},
		"far":  {Latitude: 45.3, Longitude: -75.0},
	}

	stops := stopsFor("far", "near")
	stops[0].status = StopStatusVisited

	got := OrderStopsByProximity(stops, coordinates, start)

	for _, stop := range got {
		if stop.SaleId() == "far" && stop.Status() != StopStatusVisited {
			t.Errorf("reorder lost stop status: got %v, want %v", stop.Status(), StopStatusVisited)
		}
	}
}
