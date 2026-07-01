package memory

import (
	"GarageSaleAPI/domain/seller"
	"context"
	"errors"
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
		return errors.New("seller already exists")
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
	return nil, errors.New("seller not found")
}

func (repo *InMemorySellerRepository) GetByUserId(ctx context.Context, userId string) (*seller.Seller, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	for _, value := range repo.sellerList {
		if value.UserId() == userId {
			return &value, nil
		}
	}
	return nil, errors.New("seller not found")
}
