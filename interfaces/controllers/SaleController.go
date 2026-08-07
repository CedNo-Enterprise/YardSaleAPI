package controllers

import (
	"GarageSaleAPI/application/server"
	"GarageSaleAPI/application/services"
	"GarageSaleAPI/interfaces"
	"GarageSaleAPI/interfaces/requests"
	"GarageSaleAPI/interfaces/responses"
	"encoding/json"
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
	mux.HandleFunc("GET /sale/{id}", controller.getSale)
}

func (controller *SaleController) addSale(w http.ResponseWriter, r *http.Request, userId string) {
	interfaces.ValidateContentType(w, r, "application/json")

	requestBody := http.MaxBytesReader(w, r.Body, 1048576)

	decoder := json.NewDecoder(requestBody)
	decoder.DisallowUnknownFields()

	var saleDTO requests.SaleRequest
	interfaces.Decode(w, decoder, &saleDTO)

	saleId, err := controller.saleService.AddSale(r.Context(), saleDTO)
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
