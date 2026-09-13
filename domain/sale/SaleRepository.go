package sale

import "context"

type SaleRepository interface {
	Create(context.Context, *Sale) error
	GetById(context.Context, string) (*Sale, error)
	// GetByIds returns only the sales that exist, in unspecified order. Callers
	// detect missing ids themselves so they can report which one was unknown.
	GetByIds(context.Context, []string) ([]Sale, error)
}
