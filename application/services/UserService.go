package services

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/user"
	"GarageSaleAPI/interfaces/requests"
	"context"
	"errors"
	"log/slog"
	"time"
	"unsafe"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepository user.UserRepository
	tokenGenerator TokenGenerator
}

func NewUserService(userRepository user.UserRepository, generator TokenGenerator) *UserService {
	return &UserService{userRepository: userRepository, tokenGenerator: generator}
}

func validateUser(userDTO requests.UserRequest) error {
	validate := validator.New()
	err := validate.Struct(userDTO)
	if err != nil {
		return apperror.Invalid("invalid user", err)
	}
	return nil
}

func (service *UserService) AddUser(ctx context.Context, userDTO requests.UserRequest) error {
	err := validateUser(userDTO)
	if err != nil {
		slog.Error("error adding user", "err", err.Error())
		return err
	}

	hashedPassword, err := hashPassword(userDTO.Password)
	if err != nil {
		slog.Error("error adding user", "err", err.Error())
		return err
	}

	userId := uuid.NewString()
	newUser := user.CreateUser(userId, userDTO.Username, hashedPassword, userDTO.Email, time.Now())
	err = service.userRepository.Create(ctx, newUser)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	return nil
}

func (service *UserService) GetUserById(ctx context.Context, id string) (*user.User, error) {
	if err := requireUuid(id, "user not found"); err != nil {
		return nil, err
	}

	u, err := service.userRepository.GetById(ctx, id)
	if err != nil {
		slog.Error("Error getting user by id", "id", id, "err", err.Error())
		return nil, err
	}

	return u, nil
}

func hashPassword(password string) (string, error) {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		slog.Error("Error hashing password", "err", err.Error())
		return "", err
	}
	hash := unsafe.String(&hashBytes[0], len(hashBytes))

	return hash, nil
}

func comparePasswords(hashedPassword string, plainPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	if err != nil {
		return apperror.Unauthorized("invalid credentials", err)
	}
	return nil
}

func validateLogin(loginDTO requests.LoginRequest) error {
	validate := validator.New()
	err := validate.Struct(loginDTO)
	if err != nil {
		return apperror.Unauthorized("invalid credentials", err)
	}
	return nil
}

type LoginResult struct {
	AccessToken string
	ExpiresAt   time.Time
	User        user.User
}

func (service *UserService) Login(ctx context.Context, loginDTO requests.LoginRequest) (*LoginResult, error) {
	err := validateLogin(loginDTO)
	if err != nil {
		slog.Error("error validating login", "err", err.Error())
		return nil, err
	}

	u, err := service.userRepository.GetByEmail(ctx, loginDTO.Email)
	if err != nil {
		// Only a missing user is a credential failure. A lookup that failed for
		// any other reason must not be reported as a bad password.
		if appErr, ok := errors.AsType[*apperror.AppError](err); ok && appErr.Kind != apperror.KindNotFound {
			slog.Error("error looking up user for login", "err", err.Error())
			return nil, err
		}
		return nil, apperror.Unauthorized("invalid credentials", err)
	}

	err = comparePasswords(u.Password(), loginDTO.Password)
	if err != nil {
		return nil, apperror.Unauthorized("invalid credentials", err)
	}

	token, expiresAt, err := service.tokenGenerator.Generate(u.Id())
	if err != nil {
		return nil, apperror.Internal(err)
	}

	return &LoginResult{
		AccessToken: token,
		ExpiresAt:   expiresAt,
		User:        *u,
	}, nil
}
