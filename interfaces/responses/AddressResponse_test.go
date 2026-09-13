package responses

import (
	"GarageSaleAPI/domain/address"
	"testing"
)

func TestNewAddressResponse_nilLine2DoesNotPanic(t *testing.T) {
	a := address.CreateAddress("1 Unknown Rd", nil, "Ottawa", "ON", "K1A 0B1", "CA")

	got := NewAddressResponse(a)

	if got.Line2 != "" {
		t.Errorf("Line2 = %q, want %q", got.Line2, "")
	}
	if got.Line1 != "1 Unknown Rd" {
		t.Errorf("Line1 = %q, want %q", got.Line1, "1 Unknown Rd")
	}
}

func TestNewAddressResponse_populatedLine2(t *testing.T) {
	line2 := "Unit 4"
	a := address.CreateAddress("55 ByWard Market Sq", &line2, "Ottawa", "ON", "K1A 0B1", "CA")

	got := NewAddressResponse(a)

	if got.Line2 != line2 {
		t.Errorf("Line2 = %q, want %q", got.Line2, line2)
	}
}
