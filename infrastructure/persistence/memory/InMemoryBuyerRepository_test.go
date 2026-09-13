package memory

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/address"
	"GarageSaleAPI/domain/buyer"
	"GarageSaleAPI/test"
	"testing"
	"time"
)

func newTestBuyer(id string, userId string) *buyer.Buyer {
	return buyer.CreateBuyer(id, userId, "Sam", time.Now())
}

func TestInMemoryBuyerRepository_Create(t *testing.T) {
	tests := []struct {
		name     string
		existing []buyer.Buyer
		create   *buyer.Buyer
		wantLen  int
		wantKind apperror.Kind
	}{
		{
			name:    "creates a buyer",
			create:  newTestBuyer("1", "user"),
			wantLen: 1,
		},
		{
			name:     "rejects a duplicate id",
			existing: []buyer.Buyer{*newTestBuyer("1", "user")},
			create:   newTestBuyer("1", "other-user"),
			wantLen:  1,
			wantKind: apperror.KindConflict,
		},
		{
			name:     "rejects a second profile for the same user",
			existing: []buyer.Buyer{*newTestBuyer("1", "user")},
			create:   newTestBuyer("2", "user"),
			wantLen:  1,
			wantKind: apperror.KindConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &InMemoryBuyerRepository{buyers: tt.existing}

			err := repo.Create(test.CreateTestContext(t), tt.create)

			if tt.wantKind == "" && err != nil {
				t.Fatalf("Create() error = %v, want nil", err)
			}
			if tt.wantKind != "" {
				test.AssertKind(t, err, tt.wantKind)
			}
			if len(repo.buyers) != tt.wantLen {
				t.Errorf("len(buyers) = %d, want %d", len(repo.buyers), tt.wantLen)
			}
		})
	}
}

func TestInMemoryBuyerRepository_CreateWithCancelledContext(t *testing.T) {
	repo := &InMemoryBuyerRepository{}

	err := repo.Create(test.CreateCancelledTestContext(), newTestBuyer("1", "user"))

	if err == nil {
		t.Fatalf("Create() error = nil, want a context error")
	}
	if len(repo.buyers) != 0 {
		t.Errorf("len(buyers) = %d, want 0", len(repo.buyers))
	}
}

func TestInMemoryBuyerRepository_GetByUserId(t *testing.T) {
	repo := &InMemoryBuyerRepository{buyers: []buyer.Buyer{*newTestBuyer("1", "user")}}

	found, err := repo.GetByUserId(test.CreateTestContext(t), "user")

	if err != nil {
		t.Fatalf("GetByUserId() error = %v, want nil", err)
	}
	if found.Id() != "1" {
		t.Errorf("GetByUserId() id = %q, want %q", found.Id(), "1")
	}
}

func TestInMemoryBuyerRepository_GetByUserIdMissing(t *testing.T) {
	repo := &InMemoryBuyerRepository{}

	_, err := repo.GetByUserId(test.CreateTestContext(t), "user")

	test.AssertKind(t, err, apperror.KindNotFound)
}

func TestInMemoryBuyerRepository_GetByUserIdDetachesFromTheStore(t *testing.T) {
	repo := &InMemoryBuyerRepository{buyers: []buyer.Buyer{*newTestBuyer("1", "user")}}

	found, _ := repo.GetByUserId(test.CreateTestContext(t), "user")
	found.Rename("Somebody Else")

	stored, _ := repo.GetByUserId(test.CreateTestContext(t), "user")
	if stored.DisplayName() != "Sam" {
		t.Errorf("DisplayName() = %q, want %q", stored.DisplayName(), "Sam")
	}
}

func TestInMemoryBuyerRepository_Update(t *testing.T) {
	stored := newTestBuyer("1", "user")
	repo := &InMemoryBuyerRepository{buyers: []buyer.Buyer{*stored}}

	updated := buyer.CreateBuyer("1", "user", "Samira", time.Now())
	updated.SetHomeAddress(address.CreateAddress("100 Bank St", nil, "Ottawa", "ON", "K1P 5N2", "CA"))
	if err := repo.Update(test.CreateTestContext(t), updated); err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}

	found, _ := repo.GetByUserId(test.CreateTestContext(t), "user")
	if found.DisplayName() != "Samira" {
		t.Errorf("DisplayName() = %q, want %q", found.DisplayName(), "Samira")
	}
	if found.HomeAddress() == nil || found.HomeAddress().City() != "Ottawa" {
		t.Errorf("HomeAddress() = %v, want the Ottawa address", found.HomeAddress())
	}
	if !found.CreatedAt().Equal(stored.CreatedAt()) {
		t.Errorf("CreatedAt() = %v, want %v", found.CreatedAt(), stored.CreatedAt())
	}
}

func TestInMemoryBuyerRepository_UpdateKeepsTheStoredHomeAddress(t *testing.T) {
	stored := newTestBuyer("1", "user")
	stored.SetHomeAddress(address.CreateAddress("100 Bank St", nil, "Ottawa", "ON", "K1P 5N2", "CA"))
	repo := &InMemoryBuyerRepository{buyers: []buyer.Buyer{*stored}}

	if err := repo.Update(test.CreateTestContext(t), buyer.CreateBuyer("1", "user", "Samira", time.Now())); err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}

	found, _ := repo.GetByUserId(test.CreateTestContext(t), "user")
	if found.HomeAddress() == nil || found.HomeAddress().City() != "Ottawa" {
		t.Errorf("HomeAddress() = %v, want the Ottawa address", found.HomeAddress())
	}
}

func TestInMemoryBuyerRepository_UpdateMissing(t *testing.T) {
	repo := &InMemoryBuyerRepository{}

	err := repo.Update(test.CreateTestContext(t), newTestBuyer("1", "user"))

	test.AssertKind(t, err, apperror.KindNotFound)
}
