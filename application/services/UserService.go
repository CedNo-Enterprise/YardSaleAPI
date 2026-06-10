package services

import (
	"GarageSaleAPI/domain/user"
	"GarageSaleAPI/infrastructure/persistence/memory"
	"GarageSaleAPI/interfaces/dto"
	"errors"
	"log/slog"
	"net/mail"
	"regexp"
	"time"
)

var UserRepository = new(memory.InMemoryUserRepository)

func validateUsername(username string) error {
	if username == "" {
		return errors.New("username is empty")
	} else if len(username) < 3 {
		return errors.New("username is too short")
	} else if len(username) > 15 {
		return errors.New("username is too long")
	} else if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(username) {
		return errors.New("username must only alphanumeric characters")
	}
	return nil
}

func validatePassword(password string) error {
	if password == "" {
		return errors.New("password is empty")
	} else if len(password) < 12 {
		return errors.New("password is too short")
	} else if len(password) > 64 {
		return errors.New("password is too long")
	}
	return nil
}

func validateUser(userDTO dto.UserDTO) error {
	if validateUsername(userDTO.Username) != nil {
		return errors.New("username is invalid")
	}

	if validatePassword(userDTO.Password) != nil {
		return errors.New("password is invalid")
	}

	return nil
}

func parseEmail(email string) (*mail.Address, error) {
	address, err := mail.ParseAddress(email)

	if err != nil {
		slog.Error("error parsing email")
		return nil, errors.New("bad request email")
	}

	return address, nil
}

func AddUser(userDTO dto.UserDTO) error {
	userError := validateUser(userDTO)
	if userError != nil {
		return userError
	}

	email, err := parseEmail(userDTO.Email)
	if err != nil {
		slog.Error("error parsing email")
		return err
	}

	newUser := user.CreateUser(userDTO.Username, userDTO.Password, *email, time.Now())
	err = UserRepository.AddUser(newUser)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	return nil
}

func GetUserByUsername(username string) (*user.User, error) {
	u, err := UserRepository.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	return u, nil
}
