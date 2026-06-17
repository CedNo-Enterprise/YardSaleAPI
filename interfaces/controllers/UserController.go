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

type UserController struct {
	userService *services.UserService
}

func NewUserController(userService *services.UserService) *UserController {
	return &UserController{userService}
}

func (controller *UserController) AddUserHandlersToMux(mux *http.ServeMux) {
	mux.HandleFunc("POST /user", controller.addUser)
	mux.HandleFunc("GET /user/{username}", controller.getUser)
}

func (controller *UserController) addUser(w http.ResponseWriter, r *http.Request) {
	interfaces.ValidateContentType(w, r, "application/json")

	requestBody := http.MaxBytesReader(w, r.Body, 1048576)

	decoder := json.NewDecoder(requestBody)
	decoder.DisallowUnknownFields()

	var userDTO dto.UserDTO
	interfaces.Decode(w, decoder, &userDTO)

	err := controller.userService.AddUser(userDTO)
	if err != nil {
		slog.Error("error adding user", "err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (controller *UserController) getUser(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")

	u, err := controller.userService.GetUserByUsername(username)
	if err != nil {
		slog.Error("Error getting user by username", "username", username, "err", err)
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	response := responses.NewUserResponse(u)

	interfaces.Marshal(w, response)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	interfaces.Encode(w, response)
}
