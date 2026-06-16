package memory

import (
	"GarageSaleAPI/domain/sale"
	"errors"
)

type InMemorySaleRepository struct {
	saleList []sale.Sale
}

func (repo *InMemorySaleRepository) AddSale(sale sale.Sale) error {
	duplicate, _ := repo.GetSaleById(sale.Id())
	if duplicate != nil {
		return errors.New("sale already exists")
	}

	repo.saleList = append(repo.saleList, sale)
	return nil
}

func (repo *InMemorySaleRepository) GetSaleById(id string) (*sale.Sale, error) {
	for _, value := range repo.saleList {
		if value.Id() == id {
			return &value, nil
		}
	}
	return nil, errors.New("sale not found")
}
