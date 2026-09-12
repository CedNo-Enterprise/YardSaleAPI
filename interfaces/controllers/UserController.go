package controllers

import (
	"GarageSaleAPI/application/server"
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/application/services"
	"GarageSaleAPI/interfaces"
	"GarageSaleAPI/interfaces/requests"
	"GarageSaleAPI/interfaces/responses"
	"encoding/json"
	"net/http"
)

type UserController struct {
	userService    *services.UserService
	sessionService *services.SessionService
	authMiddleware *interfaces.AuthMiddleware
}

func NewUserController(userService *services.UserService, sessionService *services.SessionService, authMiddleware *interfaces.AuthMiddleware) *UserController {
	return &UserController{userService, sessionService, authMiddleware}
}

func (controller *UserController) AddUserHandlersToMux(mux *http.ServeMux) {
	mux.HandleFunc("POST /user", controller.addUser)
	mux.HandleFunc("GET /user/{id}", controller.getUser)
	mux.HandleFunc("POST /login", controller.login)
	mux.HandleFunc("POST /logout", controller.authMiddleware.Authenticate(controller.logout))
}

func (controller *UserController) addUser(w http.ResponseWriter, r *http.Request) {
	interfaces.ValidateContentType(w, r, "application/json")

	requestBody := http.MaxBytesReader(w, r.Body, 1048576)

	decoder := json.NewDecoder(requestBody)
	decoder.DisallowUnknownFields()

	var userDTO requests.UserRequest
	interfaces.Decode(w, decoder, &userDTO)

	err := controller.userService.AddUser(r.Context(), userDTO)
	if err != nil {
		server.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (controller *UserController) getUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	u, err := controller.userService.GetUserById(r.Context(), id)
	if err != nil {
		server.WriteError(w, err)
		return
	}

	response := responses.NewUserResponse(u)

	interfaces.WriteResponse(w, response, http.StatusOK, "application/json")
}

func (controller *UserController) login(w http.ResponseWriter, r *http.Request) {
	interfaces.ValidateContentType(w, r, "application/json")

	requestBody := http.MaxBytesReader(w, r.Body, 1048576)

	decoder := json.NewDecoder(requestBody)
	decoder.DisallowUnknownFields()

	var loginDTO requests.LoginRequest
	interfaces.Decode(w, decoder, &loginDTO)

	result, err := controller.userService.Login(r.Context(), loginDTO)
	if err != nil {
		server.WriteError(w, err)
		return
	}

	response := responses.NewLoginResponse(&result.AccessToken, &result.ExpiresAt, &result.User)
	interfaces.WriteResponse(w, response, http.StatusOK, "application/json")
}

func (controller *UserController) logout(w http.ResponseWriter, r *http.Request, _ string) {
	claims, ok := interfaces.ClaimsFrom(r.Context())
	if !ok {
		server.WriteError(w, apperror.Unauthorized("invalid token claims", nil))
		return
	}

	if err := controller.sessionService.Logout(r.Context(), claims); err != nil {
		server.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
