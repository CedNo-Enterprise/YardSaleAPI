package services

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/address"
	"GarageSaleAPI/domain/sale"
	"GarageSaleAPI/infrastructure/persistence/memory"
	"GarageSaleAPI/interfaces/requests"
	"GarageSaleAPI/test"
	"fmt"
	"testing"
	"time"
)

var searchBase = time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)

func searchAddress() address.Address {
	return address.CreateAddress("1 Main St", nil, "Ottawa", "ON", "K1P 5N2", "CA")
}

func newSearchService(t *testing.T, sales ...*sale.Sale) *SaleService {
	t.Helper()

	repo := &memory.InMemorySaleRepository{}
	for _, s := range sales {
		if err := repo.Create(test.CreateTestContext(t), s); err != nil {
			t.Fatalf("seeding sale %q: %v", s.Id(), err)
		}
	}

	return NewSaleService(repo, &memory.InMemorySellerRepository{})
}

func seededSale(id string, day int, status sale.Status) *sale.Sale {
	return sale.HydrateSale(
		id, "seller", "sale "+id, searchAddress(),
		searchBase.AddDate(0, 0, day), "", []sale.SaleItem{}, status, searchBase,
	)
}

func intPointer(value int) *int {
	return &value
}

func TestSaleService_SearchSalesDefaults(t *testing.T) {
	service := newSearchService(t,
		seededSale("a", 1, sale.StatusScheduled),
		seededSale("b", 2, sale.StatusActive),
		seededSale("c", 3, sale.StatusCancelled),
	)

	result, err := service.SearchSales(test.CreateTestContext(t), requests.SaleSearchRequest{})

	if err != nil {
		t.Fatalf("SearchSales() error = %v, want nil", err)
	}
	if result.Limit != sale.DefaultSearchLimit {
		t.Errorf("Limit = %d, want %d", result.Limit, sale.DefaultSearchLimit)
	}
	if result.Offset != 0 {
		t.Errorf("Offset = %d, want 0", result.Offset)
	}
	if result.HasMore {
		t.Errorf("HasMore = true, want false")
	}
	if len(result.Sales) != 2 {
		t.Fatalf("Sales = %d, want 2 (cancelled sales are hidden)", len(result.Sales))
	}
}

func TestSaleService_SearchSalesIncludesCancelledWhenAsked(t *testing.T) {
	service := newSearchService(t,
		seededSale("a", 1, sale.StatusScheduled),
		seededSale("c", 3, sale.StatusCancelled),
	)

	result, err := service.SearchSales(test.CreateTestContext(t), requests.SaleSearchRequest{
		Statuses: []string{"cancelled"},
	})

	if err != nil {
		t.Fatalf("SearchSales() error = %v, want nil", err)
	}
	if len(result.Sales) != 1 || result.Sales[0].Id() != "c" {
		t.Errorf("Sales = %v, want only the cancelled sale", result.Sales)
	}
}

func TestSaleService_SearchSalesClampsLimit(t *testing.T) {
	service := newSearchService(t, seededSale("a", 1, sale.StatusScheduled))

	result, err := service.SearchSales(test.CreateTestContext(t), requests.SaleSearchRequest{
		Limit: intPointer(5000),
	})

	if err != nil {
		t.Fatalf("SearchSales() error = %v, want nil", err)
	}
	if result.Limit != sale.MaxSearchLimit {
		t.Errorf("Limit = %d, want %d", result.Limit, sale.MaxSearchLimit)
	}
}

func TestSaleService_SearchSalesReportsMorePages(t *testing.T) {
	seeds := make([]*sale.Sale, 0, 25)
	for i := range 25 {
		seeds = append(seeds, seededSale(fmt.Sprintf("%02d", i), i, sale.StatusScheduled))
	}
	service := newSearchService(t, seeds...)

	result, err := service.SearchSales(test.CreateTestContext(t), requests.SaleSearchRequest{})

	if err != nil {
		t.Fatalf("SearchSales() error = %v, want nil", err)
	}
	if len(result.Sales) != sale.DefaultSearchLimit {
		t.Errorf("Sales = %d, want %d", len(result.Sales), sale.DefaultSearchLimit)
	}
	if !result.HasMore {
		t.Errorf("HasMore = false, want true")
	}
}

func TestSaleService_SearchSalesLastPageReportsNoMore(t *testing.T) {
	seeds := make([]*sale.Sale, 0, 25)
	for i := range 25 {
		seeds = append(seeds, seededSale(fmt.Sprintf("%02d", i), i, sale.StatusScheduled))
	}
	service := newSearchService(t, seeds...)

	result, err := service.SearchSales(test.CreateTestContext(t), requests.SaleSearchRequest{
		Offset: intPointer(20),
	})

	if err != nil {
		t.Fatalf("SearchSales() error = %v, want nil", err)
	}
	if len(result.Sales) != 5 {
		t.Errorf("Sales = %d, want 5", len(result.Sales))
	}
	if result.HasMore {
		t.Errorf("HasMore = true, want false")
	}
}

func TestSaleService_SearchSalesAtMaxLimitStillReportsMorePages(t *testing.T) {
	seeds := make([]*sale.Sale, 0, 120)
	for i := range 120 {
		seeds = append(seeds, seededSale(fmt.Sprintf("%03d", i), i, sale.StatusScheduled))
	}
	service := newSearchService(t, seeds...)

	result, err := service.SearchSales(test.CreateTestContext(t), requests.SaleSearchRequest{
		Limit: intPointer(sale.MaxSearchLimit),
	})

	if err != nil {
		t.Fatalf("SearchSales() error = %v, want nil", err)
	}
	if len(result.Sales) != sale.MaxSearchLimit {
		t.Errorf("Sales = %d, want %d", len(result.Sales), sale.MaxSearchLimit)
	}
	if !result.HasMore {
		t.Errorf("HasMore = false, want true")
	}
}

func TestSaleService_SearchSalesFiltersByDate(t *testing.T) {
	service := newSearchService(t,
		seededSale("a", 1, sale.StatusScheduled),
		seededSale("b", 5, sale.StatusScheduled),
		seededSale("c", 9, sale.StatusScheduled),
	)
	from := searchBase.AddDate(0, 0, 5)
	to := searchBase.AddDate(0, 0, 9)

	result, err := service.SearchSales(test.CreateTestContext(t), requests.SaleSearchRequest{
		DateFrom: &from,
		DateTo:   &to,
	})

	if err != nil {
		t.Fatalf("SearchSales() error = %v, want nil", err)
	}
	if len(result.Sales) != 2 {
		t.Errorf("Sales = %d, want 2", len(result.Sales))
	}
}

func TestSaleService_SearchSalesRejectsInvalidRequests(t *testing.T) {
	earlier := searchBase
	later := searchBase.AddDate(0, 0, 5)

	tests := []struct {
		name      string
		searchDTO requests.SaleSearchRequest
	}{
		{
			name:      "limit below one",
			searchDTO: requests.SaleSearchRequest{Limit: intPointer(0)},
		},
		{
			name:      "negative limit",
			searchDTO: requests.SaleSearchRequest{Limit: intPointer(-1)},
		},
		{
			name:      "negative offset",
			searchDTO: requests.SaleSearchRequest{Offset: intPointer(-1)},
		},
		{
			name:      "unknown status",
			searchDTO: requests.SaleSearchRequest{Statuses: []string{"haggling"}},
		},
		{
			name:      "unknown sort",
			searchDTO: requests.SaleSearchRequest{Sort: "price"},
		},
		{
			name:      "dateTo before dateFrom",
			searchDTO: requests.SaleSearchRequest{DateFrom: &later, DateTo: &earlier},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := newSearchService(t, seededSale("a", 1, sale.StatusScheduled))

			_, err := service.SearchSales(test.CreateTestContext(t), tt.searchDTO)

			test.AssertKind(t, err, apperror.KindInvalid)
		})
	}
}

func TestSaleService_SearchSalesOnAnEmptyStore(t *testing.T) {
	service := newSearchService(t)

	result, err := service.SearchSales(test.CreateTestContext(t), requests.SaleSearchRequest{})

	if err != nil {
		t.Fatalf("SearchSales() error = %v, want nil", err)
	}
	if len(result.Sales) != 0 {
		t.Errorf("Sales = %v, want none", result.Sales)
	}
	if result.HasMore {
		t.Errorf("HasMore = true, want false")
	}
}
