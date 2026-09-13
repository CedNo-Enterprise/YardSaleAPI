package controllers

import (
	"GarageSaleAPI/application/server"
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/application/services"
	"GarageSaleAPI/interfaces"
	"GarageSaleAPI/interfaces/requests"
	"GarageSaleAPI/interfaces/responses"
	"net/http"
)

type SaleController struct {
	saleService    *services.SaleService
	authMiddleware *interfaces.AuthMiddleware
}

func NewSaleController(saleService *services.SaleService, authMiddleware *interfaces.AuthMiddleware) *SaleController {
	return &SaleController{saleService, authMiddleware}
}

func (controller *SaleController) AddSalesHandlersToMux(mux *http.ServeMux) {
	mux.HandleFunc("POST /sale", controller.authMiddleware.Authenticate(controller.addSale))
	mux.HandleFunc("GET /sale", controller.searchSales)
	mux.HandleFunc("GET /sale/{id}", controller.getSale)
}

func (controller *SaleController) addSale(w http.ResponseWriter, r *http.Request, userId string) {
	var saleDTO requests.SaleRequest
	if !interfaces.DecodeBody(w, r, &saleDTO) {
		return
	}

	saleId, err := controller.saleService.AddSale(r.Context(), userId, saleDTO)
	if err != nil {
		server.WriteError(w, err)
		return
	}

	w.Header().Set("Location", *saleId)
	w.WriteHeader(http.StatusCreated)
}

func (controller *SaleController) getSale(w http.ResponseWriter, r *http.Request) {
	saleId := r.PathValue("id")

	s, err := controller.saleService.GetSaleById(r.Context(), saleId)
	if err != nil {
		server.WriteError(w, err)
		return
	}

	response := responses.NewSaleResponse(*s)

	interfaces.WriteResponse(w, response, http.StatusOK, "application/json")
}

// searchSales is public, like reading a single sale. It reads only its query
// string, so a buyer's saved home address changes nothing here — the client
// passes what it wants filtered on.
func (controller *SaleController) searchSales(w http.ResponseWriter, r *http.Request) {
	searchDTO, err := parseSaleSearch(r)
	if err != nil {
		server.WriteError(w, err)
		return
	}

	result, err := controller.saleService.SearchSales(r.Context(), *searchDTO)
	if err != nil {
		server.WriteError(w, err)
		return
	}

	response := responses.NewSaleSearchResponse(result.Sales, result.Limit, result.Offset, result.HasMore)

	interfaces.WriteResponse(w, response, http.StatusOK, "application/json")
}

// parseSaleSearch turns the query string into the request DTO. Only unparseable
// values are rejected here; whether a parsed value is allowed is the service's
// call, so the rules stay in one place.
func parseSaleSearch(r *http.Request) (*requests.SaleSearchRequest, error) {
	dateFrom, err := interfaces.QueryTime(r, "dateFrom")
	if err != nil {
		return nil, apperror.Invalid("invalid dateFrom", err)
	}

	dateTo, err := interfaces.QueryTime(r, "dateTo")
	if err != nil {
		return nil, apperror.Invalid("invalid dateTo", err)
	}

	limit, err := interfaces.QueryInt(r, "limit")
	if err != nil {
		return nil, apperror.Invalid("invalid limit", err)
	}

	offset, err := interfaces.QueryInt(r, "offset")
	if err != nil {
		return nil, apperror.Invalid("invalid offset", err)
	}

	return &requests.SaleSearchRequest{
		Statuses: interfaces.QueryStrings(r, "status"),
		DateFrom: dateFrom,
		DateTo:   dateTo,
		Sort:     r.URL.Query().Get("sort"),
		Limit:    limit,
		Offset:   offset,
	}, nil
}
