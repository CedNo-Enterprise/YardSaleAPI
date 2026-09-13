package services

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/user"
	"GarageSaleAPI/infrastructure/persistence/memory"
	"GarageSaleAPI/interfaces/requests"
	"GarageSaleAPI/test"
	"testing"
	"time"
)

const (
	buyerUserId   = "bbbbbbbb-0000-4000-8000-000000000001"
	unknownUserId = "bbbbbbbb-0000-4000-8000-00000000dead"
)

func newBuyerService(t *testing.T) *BuyerService {
	t.Helper()

	userRepo := &memory.InMemoryUserRepository{}
	u := user.CreateUser(buyerUserId, "seedbuyer", "hashed", "buyer@example.com", time.Now())
	if err := userRepo.Create(test.CreateTestContext(t), u); err != nil {
		t.Fatalf("seeding user: %v", err)
	}

	return NewBuyerService(&memory.InMemoryBuyerRepository{}, userRepo)
}

func testAddressRequest() *requests.AddressRequest {
	return &requests.AddressRequest{
		Line1:      "100 Bank St",
		City:       "Ottawa",
		State:      "ON",
		PostalCode: "K1P 5N2",
		Country:    "CA",
	}
}

func TestBuyerService_AddBuyer(t *testing.T) {
	service := newBuyerService(t)

	buyerId, err := service.AddBuyer(test.CreateTestContext(t), buyerUserId, requests.BuyerRequest{
		DisplayName: "Sam",
		HomeAddress: testAddressRequest(),
	})

	if err != nil {
		t.Fatalf("AddBuyer() error = %v, want nil", err)
	}
	if buyerId == nil || *buyerId == "" {
		t.Fatalf("AddBuyer() id = %v, want a uuid", buyerId)
	}

	b, err := service.GetBuyerByUserId(test.CreateTestContext(t), buyerUserId)
	if err != nil {
		t.Fatalf("GetBuyerByUserId() error = %v, want nil", err)
	}
	if b.DisplayName() != "Sam" {
		t.Errorf("DisplayName() = %q, want %q", b.DisplayName(), "Sam")
	}
	if b.HomeAddress() == nil || b.HomeAddress().City() != "Ottawa" {
		t.Errorf("HomeAddress() = %v, want the Ottawa address", b.HomeAddress())
	}
}

func TestBuyerService_AddBuyerWithoutHomeAddress(t *testing.T) {
	service := newBuyerService(t)

	_, err := service.AddBuyer(test.CreateTestContext(t), buyerUserId, requests.BuyerRequest{DisplayName: "Sam"})

	if err != nil {
		t.Fatalf("AddBuyer() error = %v, want nil", err)
	}

	b, _ := service.GetBuyerByUserId(test.CreateTestContext(t), buyerUserId)
	if b.HomeAddress() != nil {
		t.Errorf("HomeAddress() = %v, want nil", b.HomeAddress())
	}
}

func TestBuyerService_AddBuyerRejectsInvalidRequests(t *testing.T) {
	tests := []struct {
		name     string
		userId   string
		buyerDTO requests.BuyerRequest
		wantKind apperror.Kind
	}{
		{
			name:     "missing display name",
			userId:   buyerUserId,
			buyerDTO: requests.BuyerRequest{},
			wantKind: apperror.KindInvalid,
		},
		{
			name:     "display name below the minimum length",
			userId:   buyerUserId,
			buyerDTO: requests.BuyerRequest{DisplayName: "Sa"},
			wantKind: apperror.KindInvalid,
		},
		{
			name:     "home address missing its country",
			userId:   buyerUserId,
			buyerDTO: requests.BuyerRequest{DisplayName: "Sam", HomeAddress: &requests.AddressRequest{Line1: "100 Bank St", City: "Ottawa", State: "ON", PostalCode: "K1P 5N2"}},
			wantKind: apperror.KindInvalid,
		},
		{
			name:     "unknown user",
			userId:   unknownUserId,
			buyerDTO: requests.BuyerRequest{DisplayName: "Sam"},
			wantKind: apperror.KindInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := newBuyerService(t)

			_, err := service.AddBuyer(test.CreateTestContext(t), tt.userId, tt.buyerDTO)

			test.AssertKind(t, err, tt.wantKind)
		})
	}
}

func TestBuyerService_AddBuyerTwiceForTheSameUser(t *testing.T) {
	service := newBuyerService(t)
	buyerDTO := requests.BuyerRequest{DisplayName: "Sam"}
	if _, err := service.AddBuyer(test.CreateTestContext(t), buyerUserId, buyerDTO); err != nil {
		t.Fatalf("AddBuyer() error = %v, want nil", err)
	}

	_, err := service.AddBuyer(test.CreateTestContext(t), buyerUserId, buyerDTO)

	test.AssertKind(t, err, apperror.KindConflict)
}

func TestBuyerService_GetBuyerByUserIdWhenAbsent(t *testing.T) {
	service := newBuyerService(t)

	_, err := service.GetBuyerByUserId(test.CreateTestContext(t), buyerUserId)

	test.AssertKind(t, err, apperror.KindNotFound)
}

func TestBuyerService_GetBuyerByUserIdWithMalformedId(t *testing.T) {
	service := newBuyerService(t)

	_, err := service.GetBuyerByUserId(test.CreateTestContext(t), "not-a-uuid")

	test.AssertKind(t, err, apperror.KindNotFound)
}

func TestBuyerService_UpdateBuyer(t *testing.T) {
	service := newBuyerService(t)
	if _, err := service.AddBuyer(test.CreateTestContext(t), buyerUserId, requests.BuyerRequest{DisplayName: "Sam"}); err != nil {
		t.Fatalf("AddBuyer() error = %v, want nil", err)
	}

	name := "Samira"
	b, err := service.UpdateBuyer(test.CreateTestContext(t), buyerUserId, requests.UpdateBuyerRequest{
		DisplayName: &name,
		HomeAddress: testAddressRequest(),
	})

	if err != nil {
		t.Fatalf("UpdateBuyer() error = %v, want nil", err)
	}
	if b.DisplayName() != "Samira" {
		t.Errorf("DisplayName() = %q, want %q", b.DisplayName(), "Samira")
	}

	stored, _ := service.GetBuyerByUserId(test.CreateTestContext(t), buyerUserId)
	if stored.DisplayName() != "Samira" {
		t.Errorf("stored DisplayName() = %q, want %q", stored.DisplayName(), "Samira")
	}
	if stored.HomeAddress() == nil || stored.HomeAddress().City() != "Ottawa" {
		t.Errorf("stored HomeAddress() = %v, want the Ottawa address", stored.HomeAddress())
	}
}

func TestBuyerService_UpdateBuyerRejectsInvalidRequests(t *testing.T) {
	shortName := "Sa"
	name := "Samira"

	tests := []struct {
		name      string
		seedBuyer bool
		updateDTO requests.UpdateBuyerRequest
		wantKind  apperror.Kind
	}{
		{
			name:      "empty patch",
			seedBuyer: true,
			updateDTO: requests.UpdateBuyerRequest{},
			wantKind:  apperror.KindInvalid,
		},
		{
			name:      "display name below the minimum length",
			seedBuyer: true,
			updateDTO: requests.UpdateBuyerRequest{DisplayName: &shortName},
			wantKind:  apperror.KindInvalid,
		},
		{
			name:      "no buyer profile yet",
			seedBuyer: false,
			updateDTO: requests.UpdateBuyerRequest{DisplayName: &name},
			wantKind:  apperror.KindNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := newBuyerService(t)
			if tt.seedBuyer {
				if _, err := service.AddBuyer(test.CreateTestContext(t), buyerUserId, requests.BuyerRequest{DisplayName: "Sam"}); err != nil {
					t.Fatalf("AddBuyer() error = %v, want nil", err)
				}
			}

			_, err := service.UpdateBuyer(test.CreateTestContext(t), buyerUserId, tt.updateDTO)

			test.AssertKind(t, err, tt.wantKind)
		})
	}
}
