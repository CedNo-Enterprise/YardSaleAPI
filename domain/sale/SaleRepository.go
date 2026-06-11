package sale

type SaleRepository interface {
	AddSale(sale Sale) error
	GetSaleById(id string) (*Sale, error)
}
