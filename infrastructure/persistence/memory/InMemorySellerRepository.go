package memory

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/seller"
	"context"
)

type InMemorySellerRepository struct {
	sellerList []seller.Seller
}

func (repo *InMemorySellerRepository) Save(ctx context.Context, seller seller.Seller) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	duplicate, _ := repo.GetById(ctx, seller.Id())
	if duplicate != nil {
		return apperror.Conflict("seller already exists", nil)
	}

	repo.sellerList = append(repo.sellerList, seller)
	return nil
}

func (repo *InMemorySellerRepository) GetById(ctx context.Context, id string) (*seller.Seller, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	for _, value := range repo.sellerList {
		if value.Id() == id {
			return &value, nil
		}
	}
	return nil, apperror.NotFound("seller not found", nil)
}

func (repo *InMemorySellerRepository) GetByUsername(ctx context.Context, username string) (*seller.Seller, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	for _, value := range repo.sellerList {
		if value.Username() == username {
			return &value, nil
		}
	}
	return nil, apperror.NotFound("seller not found", nil)
}
