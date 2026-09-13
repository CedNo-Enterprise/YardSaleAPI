package controllers

import (
	"GarageSaleAPI/application/server"
	"GarageSaleAPI/application/services"
	"GarageSaleAPI/interfaces"
	"GarageSaleAPI/interfaces/requests"
	"GarageSaleAPI/interfaces/responses"
	"net/http"
)

type BuyerController struct {
	buyerService   *services.BuyerService
	authMiddleware *interfaces.AuthMiddleware
}

func NewBuyerController(
	buyerService *services.BuyerService, authMiddleware *interfaces.AuthMiddleware,
) *BuyerController {
	return &BuyerController{buyerService, authMiddleware}
}

// AddBuyerHandlersToMux serves the buyer profile at /buyer/me rather than by id,
// unlike the seller. A buyer's home address is not public the way a sale address
// is, and addressing the profile as "me" makes the token the only selector, so
// there is no ownership check to forget.
func (controller *BuyerController) AddBuyerHandlersToMux(mux *http.ServeMux) {
	authenticate := controller.authMiddleware.Authenticate

	mux.HandleFunc("POST /buyer", authenticate(controller.addBuyer))
	mux.HandleFunc("GET /buyer/me", authenticate(controller.getBuyer))
	mux.HandleFunc("PATCH /buyer/me", authenticate(controller.updateBuyer))
}

func (controller *BuyerController) addBuyer(w http.ResponseWriter, r *http.Request, userId string) {
	var buyerDTO requests.BuyerRequest
	if !interfaces.DecodeBody(w, r, &buyerDTO) {
		return
	}

	buyerId, err := controller.buyerService.AddBuyer(r.Context(), userId, buyerDTO)
	if err != nil {
		server.WriteError(w, err)
		return
	}

	w.Header().Set("Location", *buyerId)
	w.WriteHeader(http.StatusCreated)
}

func (controller *BuyerController) getBuyer(w http.ResponseWriter, r *http.Request, userId string) {
	b, err := controller.buyerService.GetBuyerByUserId(r.Context(), userId)
	if err != nil {
		server.WriteError(w, err)
		return
	}

	response := responses.NewBuyerResponse(b)

	interfaces.WriteResponse(w, response, http.StatusOK, "application/json")
}

func (controller *BuyerController) updateBuyer(w http.ResponseWriter, r *http.Request, userId string) {
	var updateDTO requests.UpdateBuyerRequest
	if !interfaces.DecodeBody(w, r, &updateDTO) {
		return
	}

	b, err := controller.buyerService.UpdateBuyer(r.Context(), userId, updateDTO)
	if err != nil {
		server.WriteError(w, err)
		return
	}

	response := responses.NewBuyerResponse(b)

	interfaces.WriteResponse(w, response, http.StatusOK, "application/json")
}
