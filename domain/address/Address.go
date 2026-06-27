package address

type Address struct {
	line1      string
	line2      *string
	city       string
	state      string
	postalCode string
	country    string
	latitude   float64
	longitude  float64
}

func (a *Address) Line1() string {
	return a.line1
}

func (a *Address) Line2() string {
	return *a.line2
}

func (a *Address) City() string {
	return a.city
}

func (a *Address) State() string {
	return a.state
}

func (a *Address) PostalCode() string {
	return a.postalCode
}

func (a *Address) Country() string {
	return a.country
}

func (a *Address) Latitude() float64 {
	return a.latitude
}

func (a *Address) Longitude() float64 {
	return a.longitude
}
