package controllers

import (
	"GarageSaleAPI/application/server"
	"GarageSaleAPI/application/services"
	"GarageSaleAPI/interfaces"
	"GarageSaleAPI/interfaces/requests"
	"GarageSaleAPI/interfaces/responses"
	"net/http"
)

type ItineraryController struct {
	itineraryService *services.ItineraryService
	authMiddleware   *interfaces.AuthMiddleware
}

func NewItineraryController(
	itineraryService *services.ItineraryService, authMiddleware *interfaces.AuthMiddleware,
) *ItineraryController {
	return &ItineraryController{itineraryService, authMiddleware}
}

func (controller *ItineraryController) AddItineraryHandlersToMux(mux *http.ServeMux) {
	authenticate := controller.authMiddleware.Authenticate

	mux.HandleFunc("POST /itinerary", authenticate(controller.addItinerary))
	mux.HandleFunc("GET /itinerary", authenticate(controller.listItineraries))
	mux.HandleFunc("GET /itinerary/{id}", controller.getItinerary)
	mux.HandleFunc("PATCH /itinerary/{id}", authenticate(controller.updateItinerary))
	mux.HandleFunc("DELETE /itinerary/{id}", authenticate(controller.deleteItinerary))
	mux.HandleFunc("POST /itinerary/{id}/stop", authenticate(controller.addStop))
	mux.HandleFunc("PATCH /itinerary/{id}/stop/{saleId}", authenticate(controller.updateStopStatus))
	mux.HandleFunc("DELETE /itinerary/{id}/stop/{saleId}", authenticate(controller.removeStop))
	mux.HandleFunc("PUT /itinerary/{id}/stops/order", authenticate(controller.reorderStops))
}

func (controller *ItineraryController) addItinerary(w http.ResponseWriter, r *http.Request, userId string) {
	var itineraryDTO requests.ItineraryRequest
	if !interfaces.DecodeBody(w, r, &itineraryDTO) {
		return
	}

	itineraryId, err := controller.itineraryService.AddItinerary(r.Context(), userId, itineraryDTO)
	if err != nil {
		server.WriteError(w, err)
		return
	}

	w.Header().Set("Location", *itineraryId)
	w.WriteHeader(http.StatusCreated)
}

func (controller *ItineraryController) listItineraries(w http.ResponseWriter, r *http.Request, userId string) {
	itineraries, err := controller.itineraryService.GetItinerariesByUserId(r.Context(), userId)
	if err != nil {
		server.WriteError(w, err)
		return
	}

	response := responses.NewItineraryResponses(itineraries)

	interfaces.WriteResponse(w, response, http.StatusOK, "application/json")
}

func (controller *ItineraryController) getItinerary(w http.ResponseWriter, r *http.Request) {
	itineraryId := r.PathValue("id")

	i, err := controller.itineraryService.GetItineraryById(r.Context(), itineraryId)
	if err != nil {
		server.WriteError(w, err)
		return
	}

	response := responses.NewItineraryResponse(*i)

	interfaces.WriteResponse(w, response, http.StatusOK, "application/json")
}

func (controller *ItineraryController) updateItinerary(w http.ResponseWriter, r *http.Request, userId string) {
	itineraryId := r.PathValue("id")

	var updateDTO requests.UpdateItineraryRequest
	if !interfaces.DecodeBody(w, r, &updateDTO) {
		return
	}

	i, err := controller.itineraryService.UpdateItinerary(r.Context(), itineraryId, userId, updateDTO)
	if err != nil {
		server.WriteError(w, err)
		return
	}

	response := responses.NewItineraryResponse(*i)

	interfaces.WriteResponse(w, response, http.StatusOK, "application/json")
}

func (controller *ItineraryController) deleteItinerary(w http.ResponseWriter, r *http.Request, userId string) {
	itineraryId := r.PathValue("id")

	err := controller.itineraryService.DeleteItinerary(r.Context(), itineraryId, userId)
	if err != nil {
		server.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (controller *ItineraryController) addStop(w http.ResponseWriter, r *http.Request, userId string) {
	itineraryId := r.PathValue("id")

	var stopDTO requests.AddStopRequest
	if !interfaces.DecodeBody(w, r, &stopDTO) {
		return
	}

	i, err := controller.itineraryService.AddStop(r.Context(), itineraryId, userId, stopDTO)
	if err != nil {
		server.WriteError(w, err)
		return
	}

	response := responses.NewItineraryResponse(*i)

	interfaces.WriteResponse(w, response, http.StatusOK, "application/json")
}

func (controller *ItineraryController) updateStopStatus(w http.ResponseWriter, r *http.Request, userId string) {
	itineraryId := r.PathValue("id")
	saleId := r.PathValue("saleId")

	var statusDTO requests.UpdateStopStatusRequest
	if !interfaces.DecodeBody(w, r, &statusDTO) {
		return
	}

	i, err := controller.itineraryService.SetStopStatus(r.Context(), itineraryId, userId, saleId, statusDTO)
	if err != nil {
		server.WriteError(w, err)
		return
	}

	response := responses.NewItineraryResponse(*i)

	interfaces.WriteResponse(w, response, http.StatusOK, "application/json")
}

func (controller *ItineraryController) removeStop(w http.ResponseWriter, r *http.Request, userId string) {
	itineraryId := r.PathValue("id")
	saleId := r.PathValue("saleId")

	err := controller.itineraryService.RemoveStop(r.Context(), itineraryId, userId, saleId)
	if err != nil {
		server.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (controller *ItineraryController) reorderStops(w http.ResponseWriter, r *http.Request, userId string) {
	itineraryId := r.PathValue("id")

	var reorderDTO requests.ReorderStopsRequest
	if !interfaces.DecodeBody(w, r, &reorderDTO) {
		return
	}

	i, err := controller.itineraryService.ReorderStops(r.Context(), itineraryId, userId, reorderDTO)
	if err != nil {
		server.WriteError(w, err)
		return
	}

	response := responses.NewItineraryResponse(*i)

	interfaces.WriteResponse(w, response, http.StatusOK, "application/json")
}
