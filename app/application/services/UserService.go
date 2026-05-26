package services

import (
	"GarageSaleAPI/app/domain/user"
	"GarageSaleAPI/app/infrastructure/persistence/memory"
	"GarageSaleAPI/app/interfaces/dto"
	"errors"
	"log/slog"
	"net/http"
	"net/mail"
	"time"
)

type UserService struct {
	userRepository user.UserRepository
}

var service = UserService{
	userRepository: memory.InMemoryUserRepository{
		UserList: make([]user.User, 0),
	},
}

func AddUser(userDTO dto.UserDTO) (error, int) {
	email, err := mail.ParseAddress(userDTO.Email)
	if err != nil {
		slog.Error("error parsing email")
		return errors.New("bad request email"), http.StatusBadRequest
	}

	newUser := user.CreateUser(userDTO.Id, userDTO.Username, userDTO.Password, *email, time.Now())
	service.userRepository.AddUser(newUser)

	return nil, http.StatusCreated
}

func GetUserByUsername(username string) (*user.User, error) {
	u, err := service.userRepository.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	return u, nil
}
