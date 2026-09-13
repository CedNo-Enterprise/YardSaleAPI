package services

import (
	"GarageSaleAPI/application/server/apperror"

	"github.com/google/uuid"
)

// requireUuid rejects an id that cannot name a record before it reaches the
// database. Columns typed as uuid fail the cast there, which surfaces as a
// driver error and a 500 rather than the missing row the caller asked about.
func requireUuid(id string, notFoundMessage string) error {
	if _, err := uuid.Parse(id); err != nil {
		return apperror.NotFound(notFoundMessage, err)
	}

	return nil
}
