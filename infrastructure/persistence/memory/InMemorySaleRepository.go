package memory

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/sale"
	"context"
)

type InMemorySaleRepository struct {
	saleList []sale.Sale
}

func (repo *InMemorySaleRepository) Save(ctx context.Context, sale sale.Sale) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	duplicate, _ := repo.GetById(ctx, sale.Id())
	if duplicate != nil {
		return apperror.Conflict("sale already exists", nil)
	}

	repo.saleList = append(repo.saleList, sale)
	return nil
}

func (repo *InMemorySaleRepository) GetById(ctx context.Context, id string) (*sale.Sale, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	for _, value := range repo.saleList {
		if value.Id() == id {
			return &value, nil
		}
	}
	return nil, apperror.NotFound("sale not found", nil)
}
