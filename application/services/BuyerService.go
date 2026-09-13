package services

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/buyer"
	"GarageSaleAPI/domain/user"
	"GarageSaleAPI/interfaces/requests"
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type BuyerService struct {
	buyerRepository buyer.BuyerRepository
	userRepository  user.UserRepository
}

func NewBuyerService(buyerRepo buyer.BuyerRepository, userRepo user.UserRepository) *BuyerService {
	return &BuyerService{buyerRepository: buyerRepo, userRepository: userRepo}
}

func validateBuyer(buyerDTO requests.BuyerRequest) error {
	validate := validator.New()
	if err := validate.Struct(buyerDTO); err != nil {
		return apperror.Invalid("invalid buyer", err)
	}

	return nil
}

func validateUpdateBuyer(updateDTO requests.UpdateBuyerRequest) error {
	validate := validator.New()
	if err := validate.Struct(updateDTO); err != nil {
		return apperror.Invalid("invalid buyer", err)
	}

	// An entirely empty patch is a client mistake, not a silent no-op.
	if updateDTO.DisplayName == nil && updateDTO.HomeAddress == nil {
		return apperror.Invalid("no fields to update", nil)
	}

	return nil
}

func (service *BuyerService) AddBuyer(
	ctx context.Context, userId string, buyerDTO requests.BuyerRequest,
) (*string, error) {
	if err := validateBuyer(buyerDTO); err != nil {
		slog.Error("error adding buyer", "err", err.Error())
		return nil, err
	}

	canAdd, err := service.userExists(ctx, userId)
	if err != nil {
		slog.Error("error checking user before adding buyer", "err", err.Error())
		return nil, err
	}
	if !canAdd {
		err = apperror.Invalid("invalid userId", nil)
		slog.Error("invalid userId", "err", err)
		return nil, err
	}

	// The unique index on user_id is what actually settles a race; this check is
	// here so the ordinary second attempt gets a message that names the reason.
	if existing, err := service.buyerRepository.GetByUserId(ctx, userId); err == nil && existing != nil {
		return nil, apperror.Conflict("buyer already exists", nil)
	}

	buyerId := uuid.NewString()
	b := buyer.CreateBuyer(buyerId, userId, buyerDTO.DisplayName, time.Now())
	if buyerDTO.HomeAddress != nil {
		b.SetHomeAddress(addressFromRequest(*buyerDTO.HomeAddress))
	}

	if err := service.buyerRepository.Create(ctx, b); err != nil {
		return nil, err
	}

	return &buyerId, nil
}

func (service *BuyerService) GetBuyerByUserId(ctx context.Context, userId string) (*buyer.Buyer, error) {
	if err := requireUuid(userId, "buyer not found"); err != nil {
		return nil, err
	}

	b, err := service.buyerRepository.GetByUserId(ctx, userId)
	if err != nil {
		slog.Error("error getting buyer", "err", err.Error())
		return nil, err
	}

	return b, nil
}

func (service *BuyerService) UpdateBuyer(
	ctx context.Context, userId string, updateDTO requests.UpdateBuyerRequest,
) (*buyer.Buyer, error) {
	if err := validateUpdateBuyer(updateDTO); err != nil {
		slog.Error("error updating buyer", "err", err.Error())
		return nil, err
	}

	b, err := service.GetBuyerByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	if updateDTO.DisplayName != nil {
		b.Rename(*updateDTO.DisplayName)
	}
	if updateDTO.HomeAddress != nil {
		b.SetHomeAddress(addressFromRequest(*updateDTO.HomeAddress))
	}

	if err := service.buyerRepository.Update(ctx, b); err != nil {
		slog.Error("error updating buyer", "err", err.Error())
		return nil, err
	}

	return b, nil
}

// userExists reports whether the user is there, distinguishing "no such user"
// from "could not tell" so a failed lookup is never read as a missing user.
func (service *BuyerService) userExists(ctx context.Context, userId string) (bool, error) {
	u, err := service.userRepository.GetById(ctx, userId)
	if err == nil {
		return u != nil, nil
	}

	if appErr, ok := errors.AsType[*apperror.AppError](err); ok && appErr.Kind == apperror.KindNotFound {
		return false, nil
	}

	return false, err
}
