package sale

import "context"

type SaleRepository interface {
	Save(context.Context, Sale) error
	GetById(context.Context, string) (*Sale, error)
}
