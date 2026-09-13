package controllers

import (
	"GarageSaleAPI/application/server"
	"GarageSaleAPI/application/services"
	"GarageSaleAPI/interfaces"
	"GarageSaleAPI/interfaces/requests"
	"GarageSaleAPI/interfaces/responses"
	"net/http"
)

type SellerController struct {
	sellerService  *services.SellerService
	authMiddleware *interfaces.AuthMiddleware
}

func NewSellerController(sellerService *services.SellerService, authMiddleware *interfaces.AuthMiddleware) *SellerController {
	return &SellerController{sellerService, authMiddleware}
}

func (controller *SellerController) AddSalesHandlersToMux(mux *http.ServeMux) {
	mux.HandleFunc("POST /seller", controller.authMiddleware.Authenticate(controller.addSeller))
	mux.HandleFunc("GET /seller/user/{userId}", controller.getSellerByUserId)
	mux.HandleFunc("GET /seller/{id}", controller.getSellerById)
}

func (controller *SellerController) addSeller(w http.ResponseWriter, r *http.Request, userId string) {
	var sellerDTO requests.SellerRequest
	if !interfaces.DecodeBody(w, r, &sellerDTO) {
		return
	}

	sellerId, err := controller.sellerService.AddSeller(r.Context(), userId, sellerDTO.Username)
	if err != nil {
		server.WriteError(w, err)
		return
	}

	w.Header().Set("Location", *sellerId)
	w.WriteHeader(http.StatusCreated)
}

func (controller *SellerController) getSellerByUserId(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")

	s, err := controller.sellerService.GetSellerByUserId(r.Context(), userId)
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
