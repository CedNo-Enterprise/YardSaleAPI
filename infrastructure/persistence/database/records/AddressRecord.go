package records

type AddressRecord struct {
	Id         int64   `gorm:"column:id;primaryKey;autoIncrement"`
	Line1      string  `gorm:"column:line1;not null"`
	Line2      *string `gorm:"column:line2"`
	City       string  `gorm:"column:city;not null"`
	State      string  `gorm:"column:state;not null"`
	PostalCode string  `gorm:"column:postalcode;not null"`
	Country    string  `gorm:"column:country;not null"`
	Latitude   float64 `gorm:"column:latitude;not null"`
	Longitude  float64 `gorm:"column:longitude;not null"`
}

func (AddressRecord) TableName() string { return "addresses" }
