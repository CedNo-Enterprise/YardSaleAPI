package buyer

import (
	"GarageSaleAPI/domain/address"
	"time"
)

type Buyer struct {
	id          string
	userId      string
	displayName string
	homeAddress *address.Address
	createdAt   time.Time
}

func (b *Buyer) Id() string {
	return b.id
}

func (b *Buyer) UserId() string {
	return b.userId
}

func (b *Buyer) DisplayName() string {
	return b.displayName
}

// HomeAddress is nil until the buyer saves one. The address is handed back by
// value behind the pointer so a caller cannot reach through it and edit what
// the buyer holds, which is what the other entities get for free by returning
// addresses as values.
func (b *Buyer) HomeAddress() *address.Address {
	if b.homeAddress == nil {
		return nil
	}

	copied := *b.homeAddress
	return &copied
}

func (b *Buyer) CreatedAt() time.Time {
	return b.createdAt
}

func (b *Buyer) Rename(displayName string) {
	b.displayName = displayName
}

func (b *Buyer) SetHomeAddress(homeAddress address.Address) {
	b.homeAddress = &homeAddress
}
