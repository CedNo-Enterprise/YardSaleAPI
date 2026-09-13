package services

import (
	"GarageSaleAPI/application/server/apperror"
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

// SaleSearchResult is one page of a browse query. Limit and Offset echo what was
// actually applied after defaults and clamping, so a caller can page without
// having to reproduce those rules.
type SaleSearchResult struct {
	Sales   []sale.Sale
	Limit   int
	Offset  int
	HasMore bool
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

	saleAddress := addressFromRequest(saleDTO.Address)

	saleId := uuid.NewString()
	s := sale.CreateSale(
		saleId, saleDTO.SellerId, saleDTO.Name,
		saleAddress, saleDTO.Date, saleDTO.Description, time.Now(),
	)

	err = service.saleRepository.Create(ctx, s)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	return &saleId, nil
}

func (service *SaleService) GetSaleById(ctx context.Context, saleId string) (*sale.Sale, error) {
	if err := requireUuid(saleId, "sale not found"); err != nil {
		return nil, err
	}

	s, err := service.saleRepository.GetById(ctx, saleId)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	return s, nil
}

func validateSaleSearch(searchDTO requests.SaleSearchRequest) error {
	validate := validator.New()
	if err := validate.Struct(searchDTO); err != nil {
		return apperror.Invalid("invalid sale search", err)
	}

	// A limit over the maximum is clamped rather than refused, but a limit below
	// one is a client mistake with no sensible reading.
	if searchDTO.Limit != nil && *searchDTO.Limit < 1 {
		return apperror.Invalid("limit must be at least 1", nil)
	}
	if searchDTO.Offset != nil && *searchDTO.Offset < 0 {
		return apperror.Invalid("offset cannot be negative", nil)
	}
	if searchDTO.DateFrom != nil && searchDTO.DateTo != nil && searchDTO.DateTo.Before(*searchDTO.DateFrom) {
		return apperror.Invalid("dateTo cannot be before dateFrom", nil)
	}

	return nil
}

func searchCriteriaFrom(searchDTO requests.SaleSearchRequest) sale.SearchCriteria {
	statuses := make([]sale.Status, 0, len(searchDTO.Statuses))
	for _, status := range searchDTO.Statuses {
		statuses = append(statuses, sale.Status(status))
	}

	criteria := sale.SearchCriteria{
		Statuses: statuses,
		DateFrom: searchDTO.DateFrom,
		DateTo:   searchDTO.DateTo,
		Sort:     sale.SortOrder(searchDTO.Sort),
	}
	if searchDTO.Limit != nil {
		criteria.Limit = *searchDTO.Limit
	}
	if searchDTO.Offset != nil {
		criteria.Offset = *searchDTO.Offset
	}

	return criteria.Normalize()
}

func (service *SaleService) SearchSales(
	ctx context.Context, searchDTO requests.SaleSearchRequest,
) (*SaleSearchResult, error) {
	if err := validateSaleSearch(searchDTO); err != nil {
		slog.Error("error searching sales", "err", err.Error())
		return nil, err
	}

	criteria := searchCriteriaFrom(searchDTO)

	// One row beyond the page answers "is there more" without a count query.
	probe := criteria
	probe.Probe = true

	found, err := service.saleRepository.Search(ctx, probe)
	if err != nil {
		slog.Error("error searching sales", "err", err.Error())
		return nil, err
	}

	hasMore := len(found) > criteria.Limit
	if hasMore {
		found = found[:criteria.Limit]
	}

	return &SaleSearchResult{
		Sales:   found,
		Limit:   criteria.Limit,
		Offset:  criteria.Offset,
		HasMore: hasMore,
	}, nil
}
