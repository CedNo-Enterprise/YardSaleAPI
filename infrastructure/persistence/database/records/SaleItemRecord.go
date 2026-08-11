package records

type SaleItemRecord struct {
	Id              int64   `gorm:"column:id;primaryKey;autoIncrement"`
	SaleId          string  `gorm:"column:sale_id;type:uuid;not null;index"`
	InventoryItemId int64   `gorm:"column:inventory_item_id;not null"`
	Name            string  `gorm:"column:name"`
	Price           float64 `gorm:"column:price;not null"`
	Status          string  `gorm:"column:status;default:available"`
}

func (SaleItemRecord) TableName() string { return "sale_items" }
