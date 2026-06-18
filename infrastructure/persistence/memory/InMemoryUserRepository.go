package memory

import (
	"GarageSaleAPI/domain/user"
	"errors"
)

type InMemoryUserRepository struct {
	UserList []user.User
}

func (repo *InMemoryUserRepository) AddUser(user user.User) error {
	duplicate, _ := repo.GetUserByUsername(user.Username())
	if duplicate != nil {
		return errors.New("user already exists")
	}

	repo.UserList = append(repo.UserList, user)
	return nil
}

func (repo *InMemoryUserRepository) GetUserByUsername(username string) (*user.User, error) {
	for _, foundUser := range repo.UserList {
		if foundUser.Username() == username {
			return &foundUser, nil
		}
	}
	return nil, errors.New("user not found")
}
