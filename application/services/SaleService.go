package services

import (
	"GarageSaleAPI/domain/sale"
	"GarageSaleAPI/interfaces/dto"
	"errors"
	"log/slog"

	"github.com/google/uuid"
)

type SaleService struct {
	saleRepository sale.SaleRepository
}

func NewSaleService(saleRepository sale.SaleRepository) *SaleService {
	return &SaleService{saleRepository: saleRepository}
}

func validateSaleName(saleName string) error {
	if saleName == "" {
		return errors.New("sale name is empty")
	} else if len(saleName) < 8 {
		return errors.New("sale name is too short")
	} else if len(saleName) > 64 {
		return errors.New("sale name is too long")
	}
	return nil
}

func validateSaleAddress(saleAddress string) error {
	if saleAddress == "" {
		return errors.New("sale address is empty")
	}
	return nil
}

func validateSale(saleDTO dto.SaleDTO) error {
	if validateSaleName(saleDTO.Name) != nil {
		return errors.New("sale name is invalid")
	}

	if validateSaleAddress(saleDTO.Address) != nil {
		return errors.New("sale address is invalid")
	}

	return nil
}

func (service *SaleService) AddSale(saleDTO dto.SaleDTO) error {
	if validateSale(saleDTO) != nil {
		return errors.New("invalid sale")
	}

	saleId := uuid.NewString()
	sale := sale.CreateSale(saleId, saleDTO.Name, saleDTO.Address)

	err := service.saleRepository.AddSale(sale)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	return nil
}

func (service *SaleService) GetSaleById(saleId string) (*sale.Sale, error) {
	s, err := service.saleRepository.GetSaleById(saleId)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	return s, nil
}
