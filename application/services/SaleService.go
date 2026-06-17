package services

import (
	"GarageSaleAPI/domain/sale"
	"GarageSaleAPI/interfaces/requests"
	"errors"
	"log/slog"

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
		return errors.New("invalid sale")
	}

	return nil
}

func (service *SaleService) AddSale(saleDTO requests.SaleRequest) (*string, error) {
	err := validateSale(saleDTO)
	if err != nil {
		return nil, err
	}

	saleId := uuid.NewString()
	s := sale.CreateSale(saleId, saleDTO.Name, saleDTO.Address)

	err = service.saleRepository.AddSale(s)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	return &saleId, nil
}

func (service *SaleService) GetSaleById(saleId string) (*sale.Sale, error) {
	s, err := service.saleRepository.GetSaleById(saleId)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	return s, nil
}
