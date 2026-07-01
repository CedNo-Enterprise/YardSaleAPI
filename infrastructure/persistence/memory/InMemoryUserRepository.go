package memory

import (
	"GarageSaleAPI/domain/user"
	"context"
	"errors"
)

type InMemoryUserRepository struct {
	userList []user.User
}

func (repo *InMemoryUserRepository) AddUser(ctx context.Context, user user.User) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	duplicate, _ := repo.GetUserByUsername(ctx, user.Username())
	if duplicate != nil {
		return errors.New("user already exists")
	}

	repo.userList = append(repo.userList, user)
	return nil
}

func (repo *InMemoryUserRepository) GetUserByUsername(ctx context.Context, username string) (*user.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	for _, foundUser := range repo.userList {
		if foundUser.Username() == username {
			return &foundUser, nil
		}
	}
	return nil, errors.New("user not found")
}
