package memory

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/user"
	"context"
)

type InMemoryUserRepository struct {
	userList []user.User
}

func (repo *InMemoryUserRepository) Save(ctx context.Context, user user.User) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	duplicate, _ := repo.GetByUsername(ctx, user.Username())
	if duplicate != nil {
		return apperror.Conflict("user already exists", nil)
	}

	repo.userList = append(repo.userList, user)
	return nil
}

func (repo *InMemoryUserRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	for _, foundUser := range repo.userList {
		if foundUser.Username() == username {
			return &foundUser, nil
		}
	}
	return nil, apperror.NotFound("user not found", nil)
}
