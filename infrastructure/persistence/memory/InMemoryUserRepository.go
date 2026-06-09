package memory

import (
	"GarageSaleAPI/domain/user"
	"errors"
)

type InMemoryUserRepository struct {
	UserList []user.User
}

func (repo *InMemoryUserRepository) AddUser(user user.User) {
	//TODO: Check for duplicates
	repo.UserList = append(repo.UserList, user)
}

func (repo *InMemoryUserRepository) GetUserByUsername(username string) (*user.User, error) {
	for _, foundUser := range repo.UserList {
		if foundUser.Username == username {
			return &foundUser, nil
		}
	}
	return nil, errors.New("user not found")
}
