package buyer

import "context"

type BuyerRepository interface {
	Create(context.Context, *Buyer) error
	GetByUserId(context.Context, string) (*Buyer, error)
	Update(context.Context, *Buyer) error
}
