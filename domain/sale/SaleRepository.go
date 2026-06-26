package sale

import "context"

type SaleRepository interface {
	Save(ctx context.Context, sale Sale) error
	GetById(ctx context.Context, id string) (*Sale, error)
}
