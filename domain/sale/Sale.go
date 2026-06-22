package sale

import "GarageSaleAPI/domain/address"

type Sale struct {
	id      string
	name    string
	address address.Address
}

func (s *Sale) Id() string {
	return s.id
}

func (s *Sale) Name() string {
	return s.name
}

func (s *Sale) Address() address.Address {
	return s.address
}
