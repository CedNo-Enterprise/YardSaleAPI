package memory

import (
	"GarageSaleAPI/domain/sale"
	"errors"
)

type InMemorySaleRepository struct {
	SaleList []sale.Sale
}

func (repo *InMemorySaleRepository) AddSale(sale sale.Sale) error {
	duplicate, _ := repo.GetSaleById(sale.Id())
	if duplicate != nil {
		return errors.New("sale already exists")
	}

	repo.SaleList = append(repo.SaleList, sale)
	return nil
}

func (repo *InMemorySaleRepository) GetSaleById(id string) (*sale.Sale, error) {
	for _, value := range repo.SaleList {
		if value.Id() == id {
			return &value, nil
		}
	}
	return nil, errors.New("sale not found")
}
