package requests

type BuyerRequest struct {
	DisplayName string          `json:"displayName" validate:"required,min=3,max=32"`
	HomeAddress *AddressRequest `json:"homeAddress" validate:"omitempty"`
}

// UpdateBuyerRequest uses pointers so a PATCH can tell "field absent" from
// "field set to empty", matching UpdateItineraryRequest. A home address is
// replaced wholesale rather than patched field by field, because a half-edited
// address is not an address.
type UpdateBuyerRequest struct {
	DisplayName *string         `json:"displayName" validate:"omitempty,min=3,max=32"`
	HomeAddress *AddressRequest `json:"homeAddress" validate:"omitempty"`
}
