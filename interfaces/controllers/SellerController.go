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

type SellerController struct {
	sellerService *services.SellerService
}

func NewSellerController(sellerService *services.SellerService) *SellerController {
	return &SellerController{
		sellerService: sellerService,
	}
}

func (controller *SellerController) AddSalesHandlersToMux(mux *http.ServeMux) {
	mux.HandleFunc("POST /seller", controller.addSeller)
	mux.HandleFunc("GET /seller/user/{username}", controller.getSellerByUsername)
	mux.HandleFunc("GET /seller/{id}", controller.getSellerById)
}

func (controller *SellerController) addSeller(w http.ResponseWriter, r *http.Request) {
	interfaces.ValidateContentType(w, r, "application/json")

	requestBody := http.MaxBytesReader(w, r.Body, 1048576)

	decoder := json.NewDecoder(requestBody)
	decoder.DisallowUnknownFields()

	var sellerDTO requests.SellerRequest
	interfaces.Decode(w, decoder, &sellerDTO)

	sellerId, err := controller.sellerService.AddSeller(r.Context(), sellerDTO.Username)
	if err != nil {
		server.WriteError(w, err)
		return
	}

	w.Header().Set("Location", *sellerId)
	w.WriteHeader(http.StatusCreated)
}

func (controller *SellerController) getSellerByUsername(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")

	s, err := controller.sellerService.GetSellerByUsername(r.Context(), username)
	if err != nil {
		server.WriteError(w, err)
		return
	}

	response := responses.NewSellerResponse(s)

	interfaces.WriteResponse(w, response, http.StatusOK, "application/json")
}

func (controller *SellerController) getSellerById(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	s, err := controller.sellerService.GetSellerById(r.Context(), id)
	if err != nil {
		server.WriteError(w, err)
		return
	}

	response := responses.NewSellerResponse(s)

	interfaces.WriteResponse(w, response, http.StatusOK, "application/json")
}
