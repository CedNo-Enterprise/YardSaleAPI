package address

import "time"

func CreateAddress(
	id string,
	line1 string,
	line2 *string,
	city string,
	state string,
	postalCode string,
	country string,
	createdTime time.Time,
) Address {
	return Address{
		id:         id,
		line1:      line1,
		line2:      line2,
		city:       city,
		state:      state,
		postalCode: postalCode,
		country:    country,
		createdAt:  createdTime,
		updatedAt:  createdTime,
	}
}

func (a *Address) AddLatLong(latitude float64, longitude float64) {
	a.latitude = latitude
	a.longitude = longitude
}
