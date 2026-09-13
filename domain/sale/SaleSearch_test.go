package sale

import (
	"GarageSaleAPI/domain/address"
	"testing"
	"time"
)

var searchBase = time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)

func searchSale(id string, day int, status Status) Sale {
	date := searchBase.AddDate(0, 0, day)

	return *HydrateSale(
		id, "seller", "sale "+id,
		address.CreateAddress("1 Main St", nil, "Ottawa", "ON", "K1P 5N2", "CA"),
		date, "", []SaleItem{}, status, searchBase.AddDate(0, 0, -day),
	)
}

func idsOf(sales []Sale) []string {
	ids := make([]string, 0, len(sales))
	for _, s := range sales {
		ids = append(ids, s.Id())
	}

	return ids
}

func equalIds(got []string, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}

	return true
}

func TestSearchCriteria_Normalize(t *testing.T) {
	tests := []struct {
		name       string
		criteria   SearchCriteria
		wantLimit  int
		wantOffset int
		wantSort   SortOrder
		wantStatus []Status
	}{
		{
			name:       "zero value becomes the default first page",
			criteria:   SearchCriteria{},
			wantLimit:  DefaultSearchLimit,
			wantOffset: 0,
			wantSort:   SortDate,
			wantStatus: DefaultSearchStatuses(),
		},
		{
			name:       "limit above the maximum is clamped",
			criteria:   SearchCriteria{Limit: 5000},
			wantLimit:  MaxSearchLimit,
			wantOffset: 0,
			wantSort:   SortDate,
			wantStatus: DefaultSearchStatuses(),
		},
		{
			name:       "limit below one falls back to the default",
			criteria:   SearchCriteria{Limit: -3},
			wantLimit:  DefaultSearchLimit,
			wantOffset: 0,
			wantSort:   SortDate,
			wantStatus: DefaultSearchStatuses(),
		},
		{
			name:       "negative offset becomes the first page",
			criteria:   SearchCriteria{Offset: -10},
			wantLimit:  DefaultSearchLimit,
			wantOffset: 0,
			wantSort:   SortDate,
			wantStatus: DefaultSearchStatuses(),
		},
		{
			name:       "unknown sort falls back to date",
			criteria:   SearchCriteria{Sort: SortOrder("price")},
			wantLimit:  DefaultSearchLimit,
			wantOffset: 0,
			wantSort:   SortDate,
			wantStatus: DefaultSearchStatuses(),
		},
		{
			name:       "supplied values are kept",
			criteria:   SearchCriteria{Statuses: []Status{StatusCancelled}, Sort: SortCreated, Limit: 50, Offset: 100},
			wantLimit:  50,
			wantOffset: 100,
			wantSort:   SortCreated,
			wantStatus: []Status{StatusCancelled},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.criteria.Normalize()

			if got.Limit != tt.wantLimit {
				t.Errorf("Limit = %d, want %d", got.Limit, tt.wantLimit)
			}
			if got.Offset != tt.wantOffset {
				t.Errorf("Offset = %d, want %d", got.Offset, tt.wantOffset)
			}
			if got.Sort != tt.wantSort {
				t.Errorf("Sort = %q, want %q", got.Sort, tt.wantSort)
			}
			if len(got.Statuses) != len(tt.wantStatus) {
				t.Fatalf("Statuses = %v, want %v", got.Statuses, tt.wantStatus)
			}
			for i := range got.Statuses {
				if got.Statuses[i] != tt.wantStatus[i] {
					t.Errorf("Statuses = %v, want %v", got.Statuses, tt.wantStatus)
					break
				}
			}
		})
	}
}

func TestDefaultSearchStatuses_ExcludesCancelled(t *testing.T) {
	for _, status := range DefaultSearchStatuses() {
		if status == StatusCancelled {
			t.Errorf("DefaultSearchStatuses() = %v, want it to leave out %q", DefaultSearchStatuses(), StatusCancelled)
		}
	}
}

func TestSearchCriteria_Matches(t *testing.T) {
	early := searchBase.AddDate(0, 0, 1)
	late := searchBase.AddDate(0, 0, 3)

	tests := []struct {
		name     string
		criteria SearchCriteria
		sale     Sale
		want     bool
	}{
		{
			name:     "default statuses keep a scheduled sale",
			criteria: SearchCriteria{}.Normalize(),
			sale:     searchSale("a", 2, StatusScheduled),
			want:     true,
		},
		{
			name:     "default statuses drop a cancelled sale",
			criteria: SearchCriteria{}.Normalize(),
			sale:     searchSale("a", 2, StatusCancelled),
			want:     false,
		},
		{
			name:     "an explicit status matches it",
			criteria: SearchCriteria{Statuses: []Status{StatusCancelled}}.Normalize(),
			sale:     searchSale("a", 2, StatusCancelled),
			want:     true,
		},
		{
			name:     "a sale before the lower bound is dropped",
			criteria: SearchCriteria{DateFrom: &late}.Normalize(),
			sale:     searchSale("a", 2, StatusScheduled),
			want:     false,
		},
		{
			name:     "a sale after the upper bound is dropped",
			criteria: SearchCriteria{DateTo: &early}.Normalize(),
			sale:     searchSale("a", 2, StatusScheduled),
			want:     false,
		},
		{
			name:     "both bounds are inclusive",
			criteria: SearchCriteria{DateFrom: &early, DateTo: &early}.Normalize(),
			sale:     searchSale("a", 1, StatusScheduled),
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.criteria.Matches(tt.sale); got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApply_OrdersAndPages(t *testing.T) {
	sales := []Sale{
		searchSale("c", 3, StatusScheduled),
		searchSale("a", 1, StatusScheduled),
		searchSale("d", 4, StatusCancelled),
		searchSale("b", 2, StatusActive),
	}

	tests := []struct {
		name     string
		criteria SearchCriteria
		want     []string
	}{
		{
			name:     "orders by date and hides cancelled sales",
			criteria: SearchCriteria{},
			want:     []string{"a", "b", "c"},
		},
		{
			name:     "orders by date descending",
			criteria: SearchCriteria{Sort: SortDateReverse},
			want:     []string{"c", "b", "a"},
		},
		{
			name:     "orders by newest created",
			criteria: SearchCriteria{Sort: SortCreated},
			want:     []string{"a", "b", "c"},
		},
		{
			name:     "limits the page",
			criteria: SearchCriteria{Limit: 2},
			want:     []string{"a", "b"},
		},
		{
			name:     "offsets into the results",
			criteria: SearchCriteria{Limit: 2, Offset: 1},
			want:     []string{"b", "c"},
		},
		{
			name:     "an offset past the end returns nothing",
			criteria: SearchCriteria{Offset: 99},
			want:     []string{},
		},
		{
			name:     "an explicit status includes cancelled sales",
			criteria: SearchCriteria{Statuses: []Status{StatusCancelled}},
			want:     []string{"d"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := idsOf(Apply(sales, tt.criteria))

			if !equalIds(got, tt.want) {
				t.Errorf("Apply() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApply_BreaksTiesOnId(t *testing.T) {
	sameDay := []Sale{
		searchSale("c", 1, StatusScheduled),
		searchSale("a", 1, StatusScheduled),
		searchSale("b", 1, StatusScheduled),
	}

	got := idsOf(Apply(sameDay, SearchCriteria{}))

	if !equalIds(got, []string{"a", "b", "c"}) {
		t.Errorf("Apply() = %v, want [a b c]", got)
	}
}

func TestApply_DoesNotMutateTheInput(t *testing.T) {
	sales := []Sale{
		searchSale("c", 3, StatusScheduled),
		searchSale("a", 1, StatusScheduled),
	}

	Apply(sales, SearchCriteria{})

	if sales[0].Id() != "c" {
		t.Errorf("input order = %v, want it untouched", idsOf(sales))
	}
}
