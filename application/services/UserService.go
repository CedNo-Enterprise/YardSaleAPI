package services

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/user"
	"GarageSaleAPI/interfaces/requests"
	"context"
	"log/slog"
	"os"
	"time"
	"unsafe"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var envJwtKey = []byte(os.Getenv("JWT_SECRET"))

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

	hashedPassword, err := HashPassword(userDTO.Password)
	if err != nil {
		slog.Error("error adding user", "err", err.Error())
		return err
	}

	newUser := user.CreateUser(userDTO.Username, hashedPassword, userDTO.Email, time.Now())
	err = service.userRepository.Save(ctx, newUser)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	return nil
}

func (service *UserService) GetUserByUsername(ctx context.Context, username string) (*user.User, error) {
	u, err := service.userRepository.GetByUsername(ctx, username)
	if err != nil {
		slog.Error("Error getting user by username", "username", username, "err", err.Error())
		return nil, err
	}

	return u, nil
}

func HashPassword(password string) (string, error) {
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

func generateToken(jwtKey []byte, userID string) (time.Time, string, error) {
	expTime := time.Now().Add(24 * time.Hour)
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": expTime.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtKey)
	return expTime, signedToken, err
}

func (service *UserService) Login(ctx context.Context, loginDTO requests.LoginRequest) (*user.User, *time.Time, *string, error) {
	err := validateLogin(loginDTO)
	if err != nil {
		slog.Error("error validating login", "err", err.Error())
		return nil, nil, nil, err
	}

	u, err := service.GetUserByUsername(ctx, loginDTO.Username)
	if err != nil {
		slog.Error("login error", "username", loginDTO.Username, "err", err.Error())
		return nil, nil, nil, apperror.Unauthorized("invalid credentials", err)
	}

	err = comparePasswords(u.Password(), loginDTO.Password)
	if err != nil {
		slog.Error("login error", "err", err.Error())
		return nil, nil, nil, err
	}

	expTime, token, err := generateToken(envJwtKey, u.Username())
	if err != nil {
		slog.Error("error generating token", "err", err.Error())
		return nil, nil, nil, apperror.Internal(err)
	}

	return u, &expTime, &token, nil
}
