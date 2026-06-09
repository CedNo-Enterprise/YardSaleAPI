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

func AddUserHandlersToMux(mux *http.ServeMux) {
	mux.HandleFunc("POST /user/add", addUser)
	mux.HandleFunc("GET /user/{username}", getUser)
}

func addUser(w http.ResponseWriter, r *http.Request) {
	interfaces.ValidateContentType(w, r, "application/json")

	requestBody := http.MaxBytesReader(w, r.Body, 1048576)

	decoder := json.NewDecoder(requestBody)
	decoder.DisallowUnknownFields()

	var userDTO dto.UserDTO
	interfaces.Decode(w, decoder, &userDTO)

	err := services.AddUser(userDTO)
	if err != nil {
		slog.Error("Error adding user", "err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")

	u, err := services.GetUserByUsername(username)
	if err != nil {
		slog.Error("Error getting user by username", "username", username, "err", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	response := responses.NewUserResponse(u)

	interfaces.Marshal(w, response)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	interfaces.Encode(w, response)
}
