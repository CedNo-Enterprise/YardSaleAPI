package controllers

import (
	"GarageSaleAPI/application/services"
	"GarageSaleAPI/domain/address"
	"GarageSaleAPI/domain/sale"
	"GarageSaleAPI/infrastructure/persistence/memory"
	"GarageSaleAPI/interfaces"
	"GarageSaleAPI/interfaces/responses"
	"GarageSaleAPI/test"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var searchBase = time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)

func newSearchController(t *testing.T) *SaleController {
	t.Helper()

	repo := &memory.InMemorySaleRepository{}
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
			seed.id, "seller", "sale "+seed.id,
			address.CreateAddress("1 Main St", nil, "Ottawa", "ON", "K1P 5N2", "CA"),
			searchBase.AddDate(0, 0, seed.day), "weekend clearout", []sale.SaleItem{},
			seed.status, searchBase,
		)
		if err := repo.Create(test.CreateTestContext(t), s); err != nil {
			t.Fatalf("seeding sale %q: %v", seed.id, err)
		}
	}

	tokenService := services.NewTokenService([]byte("f81d4fae-7dec-11d0-a765-00a0c91e6bf6"), 24*time.Hour)
	authMiddleware := interfaces.NewAuthenticationMiddleware(
		tokenService, services.NewSessionService(&memory.InMemoryRevokedTokenRepository{}),
	)

	return NewSaleController(services.NewSaleService(repo), authMiddleware)
}

func searchSales(t *testing.T, controller *SaleController, query string) *httptest.ResponseRecorder {
	t.Helper()

	w := httptest.NewRecorder()
	controller.searchSales(w, httptest.NewRequest(http.MethodGet, "/sale"+query, nil))

	return w
}

func decodeSearch(t *testing.T, w *httptest.ResponseRecorder) responses.SaleSearchResponse {
	t.Helper()

	var response responses.SaleSearchResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("decoding body %q: %v", w.Body.String(), err)
	}

	return response
}

func searchedIds(response responses.SaleSearchResponse) []string {
	ids := make([]string, 0, len(response.Sales))
	for _, s := range response.Sales {
		ids = append(ids, s.Id)
	}

	return ids
}

func TestSaleController_SearchSalesDefaults(t *testing.T) {
	controller := newSearchController(t)

	w := searchSales(t, controller, "")

	if w.Code != http.StatusOK {
		t.Fatalf("searchSales() code = %d, want %d (%s)", w.Code, http.StatusOK, w.Body.String())
	}

	response := decodeSearch(t, w)
	if got := strings.Join(searchedIds(response), ","); got != "a,b,c" {
		t.Errorf("searchSales() ids = %q, want %q", got, "a,b,c")
	}
	if response.Limit != sale.DefaultSearchLimit {
		t.Errorf("searchSales() limit = %d, want %d", response.Limit, sale.DefaultSearchLimit)
	}
	if response.Offset != 0 {
		t.Errorf("searchSales() offset = %d, want 0", response.Offset)
	}
	if response.HasMore {
		t.Errorf("searchSales() has_more = true, want false")
	}
}

func TestSaleController_SearchSalesFilters(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantIds string
	}{
		{
			name:    "a single status",
			query:   "?status=cancelled",
			wantIds: "d",
		},
		{
			name:    "a repeated status",
			query:   "?status=active&status=completed",
			wantIds: "b,c",
		},
		{
			name:    "a date range",
			query:   "?dateFrom=2026-06-03T00:00:00Z&dateTo=2026-06-05T00:00:00Z",
			wantIds: "b,c",
		},
		{
			name:    "a date range that excludes everything",
			query:   "?dateFrom=2026-07-01T00:00:00Z",
			wantIds: "",
		},
		{
			name:    "date descending",
			query:   "?sort=-date",
			wantIds: "c,b,a",
		},
		{
			name:    "a page",
			query:   "?limit=2&offset=1",
			wantIds: "b,c",
		},
		{
			name:    "an empty status value is ignored",
			query:   "?status=",
			wantIds: "a,b,c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := newSearchController(t)

			w := searchSales(t, controller, tt.query)

			if w.Code != http.StatusOK {
				t.Fatalf("searchSales() code = %d, want %d (%s)", w.Code, http.StatusOK, w.Body.String())
			}
			if got := strings.Join(searchedIds(decodeSearch(t, w)), ","); got != tt.wantIds {
				t.Errorf("searchSales() ids = %q, want %q", got, tt.wantIds)
			}
		})
	}
}

func TestSaleController_SearchSalesRejectsBadQueries(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{name: "unparseable limit", query: "?limit=abc"},
		{name: "unparseable offset", query: "?offset=abc"},
		{name: "unparseable dateFrom", query: "?dateFrom=nonsense"},
		{name: "unparseable dateTo", query: "?dateTo=nonsense"},
		{name: "unknown status", query: "?status=haggling"},
		{name: "unknown sort", query: "?sort=price"},
		{name: "limit below one", query: "?limit=0"},
		{name: "negative offset", query: "?offset=-1"},
		{name: "dateTo before dateFrom", query: "?dateFrom=2026-06-09T00:00:00Z&dateTo=2026-06-01T00:00:00Z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := newSearchController(t)

			w := searchSales(t, controller, tt.query)

			if w.Code != http.StatusBadRequest {
				t.Errorf("searchSales() code = %d, want %d (%s)", w.Code, http.StatusBadRequest, w.Body.String())
			}
		})
	}
}

func TestSaleController_SearchSalesClampsLimit(t *testing.T) {
	controller := newSearchController(t)

	w := searchSales(t, controller, "?limit=5000")

	if w.Code != http.StatusOK {
		t.Fatalf("searchSales() code = %d, want %d (%s)", w.Code, http.StatusOK, w.Body.String())
	}
	if limit := decodeSearch(t, w).Limit; limit != sale.MaxSearchLimit {
		t.Errorf("searchSales() limit = %d, want %d", limit, sale.MaxSearchLimit)
	}
}

func TestSaleController_SearchSalesReturnsAnEmptyList(t *testing.T) {
	controller := newSearchController(t)

	w := searchSales(t, controller, "?offset=99")

	if w.Code != http.StatusOK {
		t.Fatalf("searchSales() code = %d, want %d (%s)", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"sales":[]`) {
		t.Errorf("searchSales() body = %s, want an empty sales list rather than null", w.Body.String())
	}
}

func TestSaleController_SearchSalesIsPublic(t *testing.T) {
	controller := newSearchController(t)
	mux := http.NewServeMux()
	controller.AddSalesHandlersToMux(mux)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/sale", nil))

	if w.Code != http.StatusOK {
		t.Errorf("GET /sale code = %d, want %d (%s)", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestSaleController_SearchSalesSummaryCarriesTheBrowseFields(t *testing.T) {
	controller := newSearchController(t)

	w := searchSales(t, controller, "?limit=1")

	response := decodeSearch(t, w)
	if len(response.Sales) != 1 {
		t.Fatalf("searchSales() sales = %d, want 1", len(response.Sales))
	}

	summary := response.Sales[0]
	if summary.Id != "a" {
		t.Errorf("summary id = %q, want %q", summary.Id, "a")
	}
	if summary.SellerId != "seller" {
		t.Errorf("summary seller_id = %q, want %q", summary.SellerId, "seller")
	}
	if summary.Status != sale.StatusScheduled {
		t.Errorf("summary status = %q, want %q", summary.Status, sale.StatusScheduled)
	}
	if summary.Description != "weekend clearout" {
		t.Errorf("summary description = %q, want %q", summary.Description, "weekend clearout")
	}
	if summary.Address.City != "Ottawa" {
		t.Errorf("summary address.city = %q, want %q", summary.Address.City, "Ottawa")
	}
	if !summary.Date.Equal(searchBase.AddDate(0, 0, 1)) {
		t.Errorf("summary date = %v, want %v", summary.Date, searchBase.AddDate(0, 0, 1))
	}
}
