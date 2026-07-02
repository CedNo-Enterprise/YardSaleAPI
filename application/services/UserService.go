package services

import (
	"GarageSaleAPI/domain/user"
	"GarageSaleAPI/interfaces/requests"
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/go-playground/validator/v10"
)

type UserService struct {
	userRepository user.UserRepository
}

func NewUserService(userRepository user.UserRepository) *UserService {
	return &UserService{userRepository: userRepository}
}

func validateUser(userDTO requests.UserRequest) error {
	validate := validator.New()
	err := validate.Struct(userDTO)
	if err != nil {
		return errors.New("invalid user")
	}
	return nil
}

func (service *UserService) AddUser(ctx context.Context, userDTO requests.UserRequest) error {
	userError := validateUser(userDTO)
	if userError != nil {
		return userError
	}

	newUser := user.CreateUser(userDTO.Username, userDTO.Password, userDTO.Email, time.Now())
	err := service.userRepository.Save(ctx, newUser)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	return nil
}

func (service *UserService) GetUserByUsername(ctx context.Context, username string) (*user.User, error) {
	u, err := service.userRepository.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	return u, nil
}
