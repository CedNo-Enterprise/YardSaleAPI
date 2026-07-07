package services

import (
	"GarageSaleAPI/domain/seller"
	"GarageSaleAPI/domain/user"
	"context"
	"errors"
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

func (service *SellerService) AddSeller(ctx context.Context, username string) (*string, error) {
	sellerId := uuid.NewString()
	s := seller.CreateSeller(sellerId, username, time.Now())

	canAdd := service.userExists(ctx, username)
	if !canAdd {
		return nil, errors.New("error adding seller")
	}

	err := service.sellerRepository.Save(ctx, s)
	if err != nil {
		return nil, err
	}

	return &sellerId, nil
}

func (service *SellerService) GetSellerById(ctx context.Context, sellerId string) (*seller.Seller, error) {
	s, err := service.sellerRepository.GetById(ctx, sellerId)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (service *SellerService) GetSellerByUsername(ctx context.Context, username string) (*seller.Seller, error) {
	s, err := service.sellerRepository.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (service *SellerService) userExists(ctx context.Context, username string) bool {
	u, err := service.userRepository.GetByUsername(ctx, username)
	if err != nil {
		return false
	}

	return u != nil
}
