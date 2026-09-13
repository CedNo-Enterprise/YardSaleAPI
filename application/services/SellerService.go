package services

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/seller"
	"GarageSaleAPI/domain/user"
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type SellerService struct {
	sellerRepository seller.SellerRepository
	userRepository   user.UserRepository
}

func NewSellerService(sellerRepo seller.SellerRepository, userRepo user.UserRepository) *SellerService {
	return &SellerService{sellerRepository: sellerRepo, userRepository: userRepo}
}

func (service *SellerService) AddSeller(ctx context.Context, userId string, username string) (*string, error) {
	sellerId := uuid.NewString()
	s := seller.CreateSeller(sellerId, userId, username, time.Now())

	canAdd, err := service.userExists(ctx, userId)
	if err != nil {
		slog.Error("error checking user before adding seller", "err", err.Error())
		return nil, err
	}
	if !canAdd {
		err = apperror.Invalid("invalid userId", nil)
		slog.Error("invalid userId", "err", err)
		return nil, err
	}

	err = service.sellerRepository.Create(ctx, s)
	if err != nil {
		return nil, err
	}

	return &sellerId, nil
}

func (service *SellerService) GetSellerById(ctx context.Context, sellerId string) (*seller.Seller, error) {
	if err := requireUuid(sellerId, "seller not found"); err != nil {
		return nil, err
	}

	s, err := service.sellerRepository.GetById(ctx, sellerId)
	if err != nil {
		slog.Error("error getting seller", "err", err.Error())
		return nil, err
	}

	return s, nil
}

func (service *SellerService) GetSellerByUserId(ctx context.Context, userId string) (*seller.Seller, error) {
	if err := requireUuid(userId, "seller not found"); err != nil {
		return nil, err
	}

	s, err := service.sellerRepository.GetByUserId(ctx, userId)
	if err != nil {
		slog.Error("error getting seller", "err", err.Error())
		return nil, err
	}

	return s, nil
}

// userExists reports whether the user is there, distinguishing "no such user"
// from "could not tell" so a failed lookup is never read as a missing user.
func (service *SellerService) userExists(ctx context.Context, userId string) (bool, error) {
	u, err := service.userRepository.GetById(ctx, userId)
	if err == nil {
		return u != nil, nil
	}

	if appErr, ok := errors.AsType[*apperror.AppError](err); ok && appErr.Kind == apperror.KindNotFound {
		return false, nil
	}

	return false, err
}
