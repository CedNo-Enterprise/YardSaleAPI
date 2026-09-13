package memory

import (
	"GarageSaleAPI/domain/sale"
	"GarageSaleAPI/test"
	"testing"
	"time"
)

var searchBase = time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)

func newSearchRepository(t *testing.T) *InMemorySaleRepository {
	t.Helper()

	repo := &InMemorySaleRepository{}
	seeds := []struct {
		id     string
		day    int
		status sale.Status
	}{
		{"a", 1, sale.StatusScheduled},
		{"b", 2, sale.StatusActive},
		{"c", 3, sale.StatusCompleted},
		{"d", 4, sale.StatusCancelled},
	}

	for _, seed := range seeds {
		s := sale.HydrateSale(
			seed.id, "seller", "sale "+seed.id, validAddress,
			searchBase.AddDate(0, 0, seed.day), "", []sale.SaleItem{},
			seed.status, searchBase,
		)
		if err := repo.Create(test.CreateTestContext(t), s); err != nil {
			t.Fatalf("seeding sale %q: %v", seed.id, err)
		}
	}

	return repo
}

func searchIds(t *testing.T, repo *InMemorySaleRepository, criteria sale.SearchCriteria) []string {
	t.Helper()

	found, err := repo.Search(test.CreateTestContext(t), criteria)
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}

	ids := make([]string, 0, len(found))
	for _, s := range found {
		ids = append(ids, s.Id())
	}

	return ids
}

func TestInMemorySaleRepository_Search(t *testing.T) {
	from := searchBase.AddDate(0, 0, 2)
	to := searchBase.AddDate(0, 0, 3)

	tests := []struct {
		name     string
		criteria sale.SearchCriteria
		want     []string
	}{
		{
			name:     "default criteria hide cancelled sales",
			criteria: sale.SearchCriteria{},
			want:     []string{"a", "b", "c"},
		},
		{
			name:     "date bounds are inclusive",
			criteria: sale.SearchCriteria{DateFrom: &from, DateTo: &to},
			want:     []string{"b", "c"},
		},
		{
			name:     "an explicit status is honoured",
			criteria: sale.SearchCriteria{Statuses: []sale.Status{sale.StatusCancelled}},
			want:     []string{"d"},
		},
		{
			name:     "several statuses are honoured",
			criteria: sale.SearchCriteria{Statuses: []sale.Status{sale.StatusActive, sale.StatusCompleted}},
			want:     []string{"b", "c"},
		},
		{
			name:     "results are paged",
			criteria: sale.SearchCriteria{Limit: 2, Offset: 1},
			want:     []string{"b", "c"},
		},
		{
			name:     "date descending reverses the order",
			criteria: sale.SearchCriteria{Sort: sale.SortDateReverse},
			want:     []string{"c", "b", "a"},
		},
		{
			name:     "an offset past the end returns nothing",
			criteria: sale.SearchCriteria{Offset: 50},
			want:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newSearchRepository(t)

			got := searchIds(t, repo, tt.criteria)

			if len(got) != len(tt.want) {
				t.Fatalf("Search() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("Search() = %v, want %v", got, tt.want)
					break
				}
			}
		})
	}
}

func TestInMemorySaleRepository_SearchOnAnEmptyStore(t *testing.T) {
	repo := &InMemorySaleRepository{}

	found, err := repo.Search(test.CreateTestContext(t), sale.SearchCriteria{})

	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if len(found) != 0 {
		t.Errorf("Search() = %v, want no results", found)
	}
}

func TestInMemorySaleRepository_SearchWithCancelledContext(t *testing.T) {
	repo := newSearchRepository(t)

	_, err := repo.Search(test.CreateCancelledTestContext(), sale.SearchCriteria{})

	if err == nil {
		t.Errorf("Search() error = nil, want a context error")
	}
}
