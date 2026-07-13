package services

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/address"
	"GarageSaleAPI/domain/sale"
	"GarageSaleAPI/interfaces/requests"
	"context"
	"log/slog"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type SaleService struct {
	saleRepository sale.SaleRepository
}

func NewSaleService(saleRepository sale.SaleRepository) *SaleService {
	return &SaleService{saleRepository: saleRepository}
}

func validateSale(saleDTO requests.SaleRequest) error {
	validate := validator.New()
	err := validate.Struct(saleDTO)
	if err != nil {
		return apperror.Invalid("invalid sale", err)
	}

	return nil
}

func (service *SaleService) AddSale(ctx context.Context, saleDTO requests.SaleRequest) (*string, error) {
	err := validateSale(saleDTO)
	if err != nil {
		slog.Error("error adding sale", "err", err.Error())
		return nil, err
	}

	saleAddress := address.CreateAddress(
		saleDTO.Address.Line1, &saleDTO.Address.Line2,
		saleDTO.Address.City, saleDTO.Address.State, saleDTO.Address.PostalCode,
		saleDTO.Address.Country,
	)

	saleId := uuid.NewString()
	s := sale.CreateSale(
		saleId, saleDTO.SellerId, saleDTO.Name,
		saleAddress, saleDTO.Date, saleDTO.Description, time.Now(),
	)

	err = service.saleRepository.Save(ctx, s)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	return &saleId, nil
}

func (service *SaleService) GetSaleById(ctx context.Context, saleId string) (*sale.Sale, error) {
	s, err := service.saleRepository.GetById(ctx, saleId)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	return s, nil
}
