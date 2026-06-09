package services

import (
	"GarageSaleAPI/domain/user"
	"GarageSaleAPI/infrastructure/persistence/memory"
	"GarageSaleAPI/interfaces/dto"
	"errors"
	"log/slog"
	"net/mail"
	"time"
)

var userRepository = new(memory.InMemoryUserRepository)

func AddUser(userDTO dto.UserDTO) error {
	email, err := mail.ParseAddress(userDTO.Email)
	if err != nil {
		slog.Error("error parsing email")
		return errors.New("bad request email")
	}

	newUser := user.CreateUser(userDTO.Id, userDTO.Username, userDTO.Password, *email, time.Now())
	userRepository.AddUser(newUser)

	return nil
}

func GetUserByUsername(username string) (*user.User, error) {
	u, err := userRepository.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	return u, nil
}
