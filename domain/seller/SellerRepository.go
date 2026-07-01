package seller

import "context"

type SellerRepository interface {
	Save(context.Context, Seller) error
	GetById(context.Context, string) (*Seller, error)
	GetByUserId(context.Context, string) (*Seller, error)
}
