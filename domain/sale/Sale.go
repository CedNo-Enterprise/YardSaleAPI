package sale

type Sale struct {
	id      string
	name    string
	address string
}

func (s *Sale) Id() string {
	return s.id
}
