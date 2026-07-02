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

func NewSellerService(repo seller.SellerRepository) *SellerService {
	return &SellerService{sellerRepository: repo}
}

func (service *SellerService) AddSeller(ctx context.Context, userId string) (*string, error) {
	sellerId := uuid.NewString()
	s := seller.CreateSeller(sellerId, userId, time.Now())

	canAdd := service.userExists(ctx, userId)
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

func (service *SellerService) GetSellerByUserId(ctx context.Context, userId string) (*seller.Seller, error) {
	s, err := service.sellerRepository.GetByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (service *SellerService) userExists(ctx context.Context, userId string) bool {
	u, err := service.userRepository.GetUserByUsername(ctx, userId)
	if err != nil {
		return false
	}

	return u != nil
}
