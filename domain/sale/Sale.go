package sale

type Sale struct {
	id      string
	name    string
	address string
}

func (s *Sale) Id() string {
	return s.id
}

func (s *Sale) Name() string {
	return s.name
}

func (s *Sale) Address() string {
	return s.address
}
