package controllers

import (
	"GarageSaleAPI/application/server"
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/application/services"
	"GarageSaleAPI/interfaces"
	"GarageSaleAPI/interfaces/requests"
	"GarageSaleAPI/interfaces/responses"
	"encoding/json"
	"errors"
	"math"
	"net"
	"net/http"
	"strconv"
	"time"
)

type UserController struct {
	userService    *services.UserService
	sessionService *services.SessionService
	loginThrottle  *services.LoginThrottleService
	authMiddleware *interfaces.AuthMiddleware
}

func NewUserController(
	userService *services.UserService,
	sessionService *services.SessionService,
	loginThrottle *services.LoginThrottleService,
	authMiddleware *interfaces.AuthMiddleware,
) *UserController {
	return &UserController{userService, sessionService, loginThrottle, authMiddleware}
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

	clientIP := clientIPOf(r)

	// Checked before Login, so a blocked attempt never reaches bcrypt.
	if retryAfter, allowed := controller.loginThrottle.Check(clientIP, loginDTO.Email); !allowed {
		respondTooManyRequests(w, retryAfter)
		return
	}

	result, err := controller.userService.Login(r.Context(), loginDTO)
	if err != nil {
		// Only credential failures count against the budget. A malformed or
		// invalid request is a client mistake, not a guess.
		if appErr, ok := errors.AsType[*apperror.AppError](err); ok && appErr.Kind == apperror.KindUnauthorized {
			controller.loginThrottle.RecordFailure(clientIP, loginDTO.Email)
		}
		server.WriteError(w, err)
		return
	}

	controller.loginThrottle.RecordSuccess(clientIP, loginDTO.Email)

	response := responses.NewLoginResponse(&result.AccessToken, &result.ExpiresAt, &result.User)
	interfaces.WriteResponse(w, response, http.StatusOK, "application/json")
}

// clientIPOf reports the address the request came from. X-Forwarded-For is
// deliberately ignored: trusting it without a proxy allowlist would let a
// caller spoof a fresh address per request and shed the per-IP budget.
func clientIPOf(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func respondTooManyRequests(w http.ResponseWriter, retryAfter time.Duration) {
	seconds := int(math.Ceil(retryAfter.Seconds()))
	if seconds < 1 {
		seconds = 1
	}

	w.Header().Set("Retry-After", strconv.Itoa(seconds))
	interfaces.WriteResponse(w, map[string]string{"error": "too many login attempts"}, http.StatusTooManyRequests, "application/json")
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
