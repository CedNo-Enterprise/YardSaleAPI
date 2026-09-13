package buyer

import (
	"GarageSaleAPI/domain/address"
	"testing"
	"time"
)

func newTestAddress() address.Address {
	return address.CreateAddress("100 Bank St", nil, "Ottawa", "ON", "K1P 5N2", "CA")
}

func TestCreateBuyer(t *testing.T) {
	createdAt := time.Now()

	b := CreateBuyer("buyer", "user", "Sam", createdAt)

	if b.Id() != "buyer" {
		t.Errorf("Id() = %q, want %q", b.Id(), "buyer")
	}
	if b.UserId() != "user" {
		t.Errorf("UserId() = %q, want %q", b.UserId(), "user")
	}
	if b.DisplayName() != "Sam" {
		t.Errorf("DisplayName() = %q, want %q", b.DisplayName(), "Sam")
	}
	if !b.CreatedAt().Equal(createdAt) {
		t.Errorf("CreatedAt() = %v, want %v", b.CreatedAt(), createdAt)
	}
}

func TestCreateBuyer_HasNoHomeAddress(t *testing.T) {
	b := CreateBuyer("buyer", "user", "Sam", time.Now())

	if b.HomeAddress() != nil {
		t.Errorf("HomeAddress() = %v, want nil", b.HomeAddress())
	}
}

func TestBuyer_Rename(t *testing.T) {
	b := CreateBuyer("buyer", "user", "Sam", time.Now())

	b.Rename("Samira")

	if b.DisplayName() != "Samira" {
		t.Errorf("DisplayName() = %q, want %q", b.DisplayName(), "Samira")
	}
}

func TestBuyer_SetHomeAddress(t *testing.T) {
	b := CreateBuyer("buyer", "user", "Sam", time.Now())

	b.SetHomeAddress(newTestAddress())

	home := b.HomeAddress()
	if home == nil {
		t.Fatalf("HomeAddress() = nil, want an address")
	}
	if home.City() != "Ottawa" {
		t.Errorf("HomeAddress().City() = %q, want %q", home.City(), "Ottawa")
	}
}

func TestBuyer_SetHomeAddressReplacesTheExistingOne(t *testing.T) {
	b := CreateBuyer("buyer", "user", "Sam", time.Now())
	b.SetHomeAddress(newTestAddress())

	b.SetHomeAddress(address.CreateAddress("9 Rideau St", nil, "Gatineau", "QC", "J8X 1A1", "CA"))

	if b.HomeAddress().City() != "Gatineau" {
		t.Errorf("HomeAddress().City() = %q, want %q", b.HomeAddress().City(), "Gatineau")
	}
}

func TestBuyer_HomeAddressDoesNotExposeStoredState(t *testing.T) {
	b := CreateBuyer("buyer", "user", "Sam", time.Now())
	b.SetHomeAddress(newTestAddress())

	b.HomeAddress().AddLatLong(45.4215, -75.6972)

	if b.HomeAddress().Latitude() != 0 {
		t.Errorf("HomeAddress().Latitude() = %v, want 0", b.HomeAddress().Latitude())
	}
}

func TestHydrateBuyer(t *testing.T) {
	home := newTestAddress()
	createdAt := time.Now()

	b := HydrateBuyer("buyer", "user", "Sam", &home, createdAt)

	if b.Id() != "buyer" || b.UserId() != "user" || b.DisplayName() != "Sam" {
		t.Errorf("HydrateBuyer() = %q/%q/%q, want buyer/user/Sam", b.Id(), b.UserId(), b.DisplayName())
	}
	if b.HomeAddress() == nil || b.HomeAddress().City() != "Ottawa" {
		t.Errorf("HomeAddress() = %v, want the Ottawa address", b.HomeAddress())
	}
	if !b.CreatedAt().Equal(createdAt) {
		t.Errorf("CreatedAt() = %v, want %v", b.CreatedAt(), createdAt)
	}
}

func TestHydrateBuyer_WithoutHomeAddress(t *testing.T) {
	b := HydrateBuyer("buyer", "user", "Sam", nil, time.Now())

	if b.HomeAddress() != nil {
		t.Errorf("HomeAddress() = %v, want nil", b.HomeAddress())
	}
}
