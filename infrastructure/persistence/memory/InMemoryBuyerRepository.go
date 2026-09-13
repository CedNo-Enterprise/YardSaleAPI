package memory

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/buyer"
	"context"
)

type InMemoryBuyerRepository struct {
	buyers []buyer.Buyer
}

// copyBuyer detaches a buyer from the store so a caller mutating what it was
// handed cannot reach back into the slice the repository holds.
func copyBuyer(b buyer.Buyer) buyer.Buyer {
	return *buyer.HydrateBuyer(b.Id(), b.UserId(), b.DisplayName(), b.HomeAddress(), b.CreatedAt())
}

func (repo *InMemoryBuyerRepository) indexOf(id string) int {
	for index, value := range repo.buyers {
		if value.Id() == id {
			return index
		}
	}

	return -1
}

func (repo *InMemoryBuyerRepository) indexOfUser(userId string) int {
	for index, value := range repo.buyers {
		if value.UserId() == userId {
			return index
		}
	}

	return -1
}

func (repo *InMemoryBuyerRepository) Create(ctx context.Context, b *buyer.Buyer) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if repo.indexOf(b.Id()) != -1 {
		return apperror.Conflict("buyer already exists", nil)
	}

	// buyers.user_id is unique in the database, so one user cannot hold two
	// buyer profiles here either.
	if repo.indexOfUser(b.UserId()) != -1 {
		return apperror.Conflict("buyer already exists", nil)
	}

	repo.buyers = append(repo.buyers, copyBuyer(*b))
	return nil
}

func (repo *InMemoryBuyerRepository) GetByUserId(ctx context.Context, userId string) (*buyer.Buyer, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	index := repo.indexOfUser(userId)
	if index == -1 {
		return nil, apperror.NotFound("buyer not found", nil)
	}

	found := copyBuyer(repo.buyers[index])
	return &found, nil
}

func (repo *InMemoryBuyerRepository) Update(ctx context.Context, b *buyer.Buyer) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	index := repo.indexOf(b.Id())
	if index == -1 {
		return apperror.NotFound("buyer not found", nil)
	}

	// The database repository writes the display name and home address only, so
	// the stored user and creation time survive an update untouched.
	stored := repo.buyers[index]
	home := b.HomeAddress()
	if home == nil {
		home = stored.HomeAddress()
	}

	repo.buyers[index] = *buyer.HydrateBuyer(
		stored.Id(), stored.UserId(), b.DisplayName(), home, stored.CreatedAt(),
	)

	return nil
}
