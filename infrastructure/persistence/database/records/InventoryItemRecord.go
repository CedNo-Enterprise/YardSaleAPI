package records

type InventoryItemRecord struct {
	Id          int64   `gorm:"column:id;primaryKey;autoIncrement"`
	SellerId    string  `gorm:"column:seller_id;type:uuid;not null;index"`
	Name        string  `gorm:"column:name"`
	Description string  `gorm:"column:description"`
	Price       float64 `gorm:"column:price;not null"`
	Status      string  `gorm:"column:status;default:available"`
}

func (InventoryItemRecord) TableName() string { return "inventory_items" }
