package records

type SavedAddressRecord struct {
	Id        int64         `gorm:"column:id;primaryKey;autoIncrement"`
	SellerId  string        `gorm:"column:seller_id;type:uuid;not null;index"`
	Label     string        `gorm:"column:label"`
	AddressId int64         `gorm:"column:address_id;not null"`
	IsDefault bool          `gorm:"column:is_default;default:false"`
	Address   AddressRecord `gorm:"foreignKey:AddressId;references:Id"`
}

func (SavedAddressRecord) TableName() string { return "saved_addresses" }
