package sale

import "context"

type SaleRepository interface {
	Create(context.Context, *Sale) error
	GetById(context.Context, string) (*Sale, error)
	// GetByIds returns only the sales that exist, in unspecified order. Callers
	// detect missing ids themselves so they can report which one was unknown.
	GetByIds(context.Context, []string) ([]Sale, error)
	// Search returns the sales matching the criteria, already ordered and paged.
	// Sale items are not loaded: no browse response shows them, so fetching them
	// would be a second query per page for nothing.
	Search(context.Context, SearchCriteria) ([]Sale, error)
}
