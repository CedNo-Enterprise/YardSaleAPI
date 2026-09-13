package controllers

import (
	"GarageSaleAPI/application/services"
	"GarageSaleAPI/domain/user"
	"GarageSaleAPI/infrastructure/persistence/memory"
	"GarageSaleAPI/interfaces"
	"GarageSaleAPI/interfaces/requests"
	"GarageSaleAPI/interfaces/responses"
	"GarageSaleAPI/test"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	buyerUserId      = "bbbbbbbb-0000-4000-8000-000000000001"
	strangerUserId   = "bbbbbbbb-0000-4000-8000-000000000002"
	buyerTokenSecret = "f81d4fae-7dec-11d0-a765-00a0c91e6bf6"
)

func buyerRequest(method string, target string, body string) *http.Request {
	var reader io.Reader
	if body != "" {
		reader = bytes.NewBufferString(body)
	}

	r := httptest.NewRequest(method, target, reader)
	if body != "" {
		r.Header.Set("Content-Type", "application/json")
	}

	return r
}

func newBuyerController(t *testing.T) (*BuyerController, *services.BuyerService, *services.TokenService) {
	t.Helper()

	userRepo := &memory.InMemoryUserRepository{}
	for userId, email := range map[string]string{buyerUserId: "buyer@example.com", strangerUserId: "stranger@example.com"} {
		u := user.CreateUser(userId, "seedbuyer", "hashed", email, time.Now())
		if err := userRepo.Create(test.CreateTestContext(t), u); err != nil {
			t.Fatalf("seeding user %q: %v", userId, err)
		}
	}

	service := services.NewBuyerService(&memory.InMemoryBuyerRepository{}, userRepo)
	tokenService := services.NewTokenService([]byte(buyerTokenSecret), 24*time.Hour)
	authMiddleware := interfaces.NewAuthenticationMiddleware(
		tokenService, services.NewSessionService(&memory.InMemoryRevokedTokenRepository{}),
	)

	return NewBuyerController(service, authMiddleware), service, tokenService
}

func seedBuyer(t *testing.T, service *services.BuyerService, userId string) string {
	t.Helper()

	buyerId, err := service.AddBuyer(test.CreateTestContext(t), userId, requests.BuyerRequest{DisplayName: "Sam"})
	if err != nil {
		t.Fatalf("seeding buyer: %v", err)
	}

	return *buyerId
}

func decodeBuyer(t *testing.T, w *httptest.ResponseRecorder) responses.BuyerResponse {
	t.Helper()

	var response responses.BuyerResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("decoding body %q: %v", w.Body.String(), err)
	}

	return response
}

func TestBuyerController_AddBuyer(t *testing.T) {
	controller, _, _ := newBuyerController(t)
	w := httptest.NewRecorder()

	controller.addBuyer(w, buyerRequest(http.MethodPost, "/buyer", `{"displayName":"Sam"}`), buyerUserId)

	if w.Code != http.StatusCreated {
		t.Fatalf("addBuyer() code = %d, want %d (%s)", w.Code, http.StatusCreated, w.Body.String())
	}
	if w.Header().Get("Location") == "" {
		t.Errorf("addBuyer() Location header is empty")
	}
}

func TestBuyerController_AddBuyerWithHomeAddress(t *testing.T) {
	controller, service, _ := newBuyerController(t)
	body := `{"displayName":"Sam","homeAddress":{"line1":"100 Bank St","city":"Ottawa","state":"ON","postal_code":"K1P 5N2","country":"CA"}}`
	w := httptest.NewRecorder()

	controller.addBuyer(w, buyerRequest(http.MethodPost, "/buyer", body), buyerUserId)

	if w.Code != http.StatusCreated {
		t.Fatalf("addBuyer() code = %d, want %d (%s)", w.Code, http.StatusCreated, w.Body.String())
	}

	b, _ := service.GetBuyerByUserId(test.CreateTestContext(t), buyerUserId)
	if b.HomeAddress() == nil || b.HomeAddress().City() != "Ottawa" {
		t.Errorf("HomeAddress() = %v, want the Ottawa address", b.HomeAddress())
	}
}

func TestBuyerController_AddBuyerRejectsBadRequests(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
		wantCode    int
	}{
		{
			name:     "missing display name",
			body:     `{}`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "unknown field",
			body:     `{"displayName":"Sam","nickname":"Sammy"}`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "malformed json",
			body:     `{"displayName":`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:        "wrong content type",
			contentType: "text/plain",
			body:        `{"displayName":"Sam"}`,
			wantCode:    http.StatusUnsupportedMediaType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, _, _ := newBuyerController(t)
			r := buyerRequest(http.MethodPost, "/buyer", tt.body)
			if tt.contentType != "" {
				r.Header.Set("Content-Type", tt.contentType)
			}
			w := httptest.NewRecorder()

			controller.addBuyer(w, r, buyerUserId)

			if w.Code != tt.wantCode {
				t.Errorf("addBuyer() code = %d, want %d (%s)", w.Code, tt.wantCode, w.Body.String())
			}
		})
	}
}

func TestBuyerController_AddBuyerTwice(t *testing.T) {
	controller, service, _ := newBuyerController(t)
	seedBuyer(t, service, buyerUserId)
	w := httptest.NewRecorder()

	controller.addBuyer(w, buyerRequest(http.MethodPost, "/buyer", `{"displayName":"Sam"}`), buyerUserId)

	if w.Code != http.StatusConflict {
		t.Errorf("addBuyer() code = %d, want %d (%s)", w.Code, http.StatusConflict, w.Body.String())
	}
}

func TestBuyerController_GetBuyer(t *testing.T) {
	controller, service, _ := newBuyerController(t)
	buyerId := seedBuyer(t, service, buyerUserId)
	w := httptest.NewRecorder()

	controller.getBuyer(w, buyerRequest(http.MethodGet, "/buyer/me", ""), buyerUserId)

	if w.Code != http.StatusOK {
		t.Fatalf("getBuyer() code = %d, want %d (%s)", w.Code, http.StatusOK, w.Body.String())
	}

	response := decodeBuyer(t, w)
	if response.Id != buyerId {
		t.Errorf("getBuyer() id = %q, want %q", response.Id, buyerId)
	}
	if response.DisplayName != "Sam" {
		t.Errorf("getBuyer() display_name = %q, want %q", response.DisplayName, "Sam")
	}
	if response.HomeAddress != nil {
		t.Errorf("getBuyer() home_address = %v, want it omitted", response.HomeAddress)
	}
}

func TestBuyerController_GetBuyerWhenAbsent(t *testing.T) {
	controller, _, _ := newBuyerController(t)
	w := httptest.NewRecorder()

	controller.getBuyer(w, buyerRequest(http.MethodGet, "/buyer/me", ""), buyerUserId)

	if w.Code != http.StatusNotFound {
		t.Errorf("getBuyer() code = %d, want %d (%s)", w.Code, http.StatusNotFound, w.Body.String())
	}
}

func TestBuyerController_GetBuyerIsScopedToTheCaller(t *testing.T) {
	controller, service, _ := newBuyerController(t)
	seedBuyer(t, service, buyerUserId)
	w := httptest.NewRecorder()

	controller.getBuyer(w, buyerRequest(http.MethodGet, "/buyer/me", ""), strangerUserId)

	if w.Code != http.StatusNotFound {
		t.Errorf("getBuyer() code = %d, want %d (%s)", w.Code, http.StatusNotFound, w.Body.String())
	}
}

func TestBuyerController_UpdateBuyer(t *testing.T) {
	controller, service, _ := newBuyerController(t)
	seedBuyer(t, service, buyerUserId)
	w := httptest.NewRecorder()

	controller.updateBuyer(w, buyerRequest(http.MethodPatch, "/buyer/me", `{"displayName":"Samira"}`), buyerUserId)

	if w.Code != http.StatusOK {
		t.Fatalf("updateBuyer() code = %d, want %d (%s)", w.Code, http.StatusOK, w.Body.String())
	}
	if decodeBuyer(t, w).DisplayName != "Samira" {
		t.Errorf("updateBuyer() display_name = %q, want %q", decodeBuyer(t, w).DisplayName, "Samira")
	}
}

func TestBuyerController_UpdateBuyerRejectsAnEmptyPatch(t *testing.T) {
	controller, service, _ := newBuyerController(t)
	seedBuyer(t, service, buyerUserId)
	w := httptest.NewRecorder()

	controller.updateBuyer(w, buyerRequest(http.MethodPatch, "/buyer/me", `{}`), buyerUserId)

	if w.Code != http.StatusBadRequest {
		t.Errorf("updateBuyer() code = %d, want %d (%s)", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestBuyerController_UpdateBuyerWhenAbsent(t *testing.T) {
	controller, _, _ := newBuyerController(t)
	w := httptest.NewRecorder()

	controller.updateBuyer(w, buyerRequest(http.MethodPatch, "/buyer/me", `{"displayName":"Samira"}`), buyerUserId)

	if w.Code != http.StatusNotFound {
		t.Errorf("updateBuyer() code = %d, want %d (%s)", w.Code, http.StatusNotFound, w.Body.String())
	}
}

func TestBuyerController_RoutesRequireAuthentication(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		target     string
		body       string
		authHeader string
		wantCode   int
	}{
		{
			name:     "create without a token",
			method:   http.MethodPost,
			target:   "/buyer",
			body:     `{"displayName":"Sam"}`,
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "read without a token",
			method:   http.MethodGet,
			target:   "/buyer/me",
			wantCode: http.StatusUnauthorized,
		},
		{
			name:       "read with a malformed token",
			method:     http.MethodGet,
			target:     "/buyer/me",
			authHeader: "Bearer not-a-token",
			wantCode:   http.StatusUnauthorized,
		},
		{
			name:     "update without a token",
			method:   http.MethodPatch,
			target:   "/buyer/me",
			body:     `{"displayName":"Samira"}`,
			wantCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, _, _ := newBuyerController(t)
			mux := http.NewServeMux()
			controller.AddBuyerHandlersToMux(mux)

			r := buyerRequest(tt.method, tt.target, tt.body)
			if tt.authHeader != "" {
				r.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, r)

			if w.Code != tt.wantCode {
				t.Errorf("%s %s code = %d, want %d (%s)", tt.method, tt.target, w.Code, tt.wantCode, w.Body.String())
			}
		})
	}
}

func TestBuyerController_RoutesAcceptAValidToken(t *testing.T) {
	controller, service, tokenService := newBuyerController(t)
	seedBuyer(t, service, buyerUserId)
	token, _, err := tokenService.Generate(buyerUserId)
	if err != nil {
		t.Fatalf("generating token: %v", err)
	}

	mux := http.NewServeMux()
	controller.AddBuyerHandlersToMux(mux)
	r := buyerRequest(http.MethodGet, "/buyer/me", "")
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /buyer/me code = %d, want %d (%s)", w.Code, http.StatusOK, w.Body.String())
	}
	if decodeBuyer(t, w).DisplayName != "Sam" {
		t.Errorf("GET /buyer/me display_name = %q, want %q", decodeBuyer(t, w).DisplayName, "Sam")
	}
}
