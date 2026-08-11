package sale

import "context"

type SaleRepository interface {
	Create(context.Context, *Sale) error
	GetById(context.Context, string) (*Sale, error)
}
