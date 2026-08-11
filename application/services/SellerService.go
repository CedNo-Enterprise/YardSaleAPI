package services

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/seller"
	"GarageSaleAPI/domain/user"
	"context"
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

func (service *SellerService) AddSeller(ctx context.Context, userId string) (*string, error) {
	sellerId := uuid.NewString()
	s := seller.CreateSeller(sellerId, userId, time.Now())

	canAdd := service.userExists(ctx, userId)
	if !canAdd {
		err := apperror.Invalid("invalid username", nil)
		slog.Error("invalid username", "err", err)
		return nil, err
	}

	err := service.sellerRepository.Create(ctx, s)
	if err != nil {
		return nil, err
	}

	return &sellerId, nil
}

func (service *SellerService) GetSellerById(ctx context.Context, sellerId string) (*seller.Seller, error) {
	s, err := service.sellerRepository.GetById(ctx, sellerId)
	if err != nil {
		slog.Error("error getting seller", "err", err.Error())
		return nil, err
	}

	return s, nil
}

func (service *SellerService) GetSellerByUserId(ctx context.Context, userId string) (*seller.Seller, error) {
	s, err := service.sellerRepository.GetByUserId(ctx, userId)
	if err != nil {
		slog.Error("error getting seller", "err", err.Error())
		return nil, err
	}

	return s, nil
}

func (service *SellerService) userExists(ctx context.Context, userId string) bool {
	u, err := service.userRepository.GetById(ctx, userId)
	if err != nil {
		return false
	}

	return u != nil
}
