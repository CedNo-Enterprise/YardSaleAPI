package address

func CreateAddress(
	line1 string,
	line2 *string,
	city string,
	state string,
	postalCode string,
	country string,
) Address {
	return Address{
		line1:      line1,
		line2:      line2,
		city:       city,
		state:      state,
		postalCode: postalCode,
		country:    country,
	}
}

func (a *Address) AddLatLong(latitude float64, longitude float64) {
	a.latitude = latitude
	a.longitude = longitude
}
