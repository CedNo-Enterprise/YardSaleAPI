package controllers

import (
	"GarageSaleAPI/application/services"
	"GarageSaleAPI/interfaces"
	"GarageSaleAPI/interfaces/dto"
	"GarageSaleAPI/interfaces/responses"
	"encoding/json"
	"log/slog"
	"net/http"
)

type SaleController struct {
	saleService *services.SaleService
}

func NewSaleController(saleService *services.SaleService) *SaleController {
	return &SaleController{saleService}
}

func (controller *SaleController) AddSalesHandlersToMux(mux *http.ServeMux) {
	mux.HandleFunc("POST /sale", controller.addSale)
	mux.HandleFunc("GET /sale/{id}", controller.getSale)
}

func (controller *SaleController) addSale(w http.ResponseWriter, r *http.Request) {
	interfaces.ValidateContentType(w, r, "application/json")

	requestBody := http.MaxBytesReader(w, r.Body, 1048576)

	decoder := json.NewDecoder(requestBody)
	decoder.DisallowUnknownFields()

	var saleDTO dto.SaleDTO
	interfaces.Decode(w, decoder, &saleDTO)

	saleId, err := controller.saleService.AddSale(saleDTO)
	if err != nil {
		slog.Error("error adding user", "err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", *saleId)
	w.WriteHeader(http.StatusCreated)
}

func (controller *SaleController) getSale(w http.ResponseWriter, r *http.Request) {
	saleId := r.PathValue("id")

	s, err := controller.saleService.GetSaleById(saleId)
	if err != nil {
		slog.Error("error getting user", "err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := responses.NewSaleResponse(*s)

	interfaces.Marshal(w, response)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	interfaces.Encode(w, response)

}
