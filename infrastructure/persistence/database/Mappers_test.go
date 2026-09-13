package database

import (
	"GarageSaleAPI/domain/seller"
	"testing"
	"time"
)

func TestSellerToRecord_carriesEveryField(t *testing.T) {
	createdAt := time.Now()
	s := seller.CreateSeller("seller-id", "user-id", "username", createdAt)

	record := sellerToRecord(s)

	if record.Id != "seller-id" {
		t.Errorf("Id = %q, want %q", record.Id, "seller-id")
	}
	if record.UserId != "user-id" {
		t.Errorf("UserId = %q, want %q", record.UserId, "user-id")
	}
	if record.Name != "username" {
		t.Errorf("Name = %q, want %q", record.Name, "username")
	}
	if !record.CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt = %v, want %v", record.CreatedAt, createdAt)
	}
}
