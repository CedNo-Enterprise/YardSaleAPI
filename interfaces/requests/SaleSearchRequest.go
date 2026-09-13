package requests

import "time"

// SaleSearchRequest is filled from the query string rather than a JSON body, so
// its fields carry no json tags. Pointers mark an absent parameter, which the
// service needs in order to reject an explicit bad value without rejecting an
// omitted one.
type SaleSearchRequest struct {
	Statuses []string   `validate:"omitempty,dive,oneof=scheduled active completed cancelled"`
	DateFrom *time.Time `validate:"omitempty"`
	DateTo   *time.Time `validate:"omitempty"`
	Sort     string     `validate:"omitempty,oneof=date -date created"`
	Limit    *int       `validate:"omitempty"`
	Offset   *int       `validate:"omitempty"`
}
