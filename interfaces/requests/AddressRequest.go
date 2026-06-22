package requests

type AddressRequest struct {
	Line1      string `json:"line1"       validate:"required"`
	Line2      string `json:"line2"`
	City       string `json:"city"        validate:"required"`
	State      string `json:"state"       validate:"required"`
	PostalCode string `json:"postal_code" validate:"required"`
	Country    string `json:"country"     validate:"required,iso3166_1_alpha2"`
}
