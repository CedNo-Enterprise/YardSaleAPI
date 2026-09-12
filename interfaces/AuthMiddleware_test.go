package interfaces

import (
	"GarageSaleAPI/application/services"
	"GarageSaleAPI/infrastructure/persistence/memory"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testKey = "f81d4fae-7dec-11d0-a765-00a0c91e6bf6"

// failingRevocationChecker stands in for a database that is down.
type failingRevocationChecker struct{}

func (failingRevocationChecker) IsRevoked(context.Context, string) (bool, error) {
	return false, errors.New("database unavailable")
}

func newTestMiddleware(checker services.RevocationChecker) (*AuthMiddleware, *services.TokenService) {
	tokenService := services.NewTokenService([]byte(testKey), time.Hour)
	return NewAuthenticationMiddleware(tokenService, checker), tokenService
}

// spyHandler records whether the protected handler was reached.
func spyHandler(reached *bool, gotUserID *string) AuthenticatedHandler {
	return func(w http.ResponseWriter, r *http.Request, userID string) {
		*reached = true
		*gotUserID = userID
		w.WriteHeader(http.StatusOK)
	}
}

func requestWithToken(tokenStr string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if tokenStr != "" {
		r.Header.Set("Authorization", "Bearer "+tokenStr)
	}
	return r
}

func TestAuthMiddleware_ValidTokenReachesHandler(t *testing.T) {
	middleware, tokenService := newTestMiddleware(services.NewSessionService(&memory.InMemoryRevokedTokenRepository{}))

	tokenStr, _, err := tokenService.Generate("user-123")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	var reached bool
	var gotUserID string
	w := httptest.NewRecorder()

	middleware.Authenticate(spyHandler(&reached, &gotUserID))(w, requestWithToken(tokenStr))

	if !reached {
		t.Error("expected the handler to be reached")
	}
	if gotUserID != "user-123" {
		t.Errorf("userID = %q, want %q", gotUserID, "user-123")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAuthMiddleware_ValidTokenExposesClaims(t *testing.T) {
	middleware, tokenService := newTestMiddleware(services.NewSessionService(&memory.InMemoryRevokedTokenRepository{}))

	tokenStr, _, err := tokenService.Generate("user-123")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	var claims jwt.MapClaims
	var found bool
	handler := func(w http.ResponseWriter, r *http.Request, _ string) {
		claims, found = ClaimsFrom(r.Context())
	}

	middleware.Authenticate(handler)(httptest.NewRecorder(), requestWithToken(tokenStr))

	if !found {
		t.Fatal("expected claims on the request context")
	}
	if jti, ok := claims["jti"].(string); !ok || jti == "" {
		t.Errorf("expected a jti claim, got %v", claims["jti"])
	}
}

func TestAuthMiddleware_RevokedTokenIsRejected(t *testing.T) {
	sessionService := services.NewSessionService(&memory.InMemoryRevokedTokenRepository{})
	middleware, tokenService := newTestMiddleware(sessionService)

	tokenStr, _, err := tokenService.Generate("user-123")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	claims, err := tokenService.Verify(tokenStr)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if err := sessionService.Logout(context.Background(), claims); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}

	var reached bool
	var gotUserID string
	w := httptest.NewRecorder()

	middleware.Authenticate(spyHandler(&reached, &gotUserID))(w, requestWithToken(tokenStr))

	if reached {
		t.Error("expected a revoked token not to reach the handler")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_TokenWithoutJtiIsRejected(t *testing.T) {
	middleware, _ := newTestMiddleware(services.NewSessionService(&memory.InMemoryRevokedTokenRepository{}))

	// A token in the pre-logout format: signed correctly, but carrying no jti.
	legacy := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := legacy.SignedString([]byte(testKey))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	var reached bool
	var gotUserID string
	w := httptest.NewRecorder()

	middleware.Authenticate(spyHandler(&reached, &gotUserID))(w, requestWithToken(tokenStr))

	if reached {
		t.Error("expected a token without jti not to reach the handler")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_RevocationLookupFailureFailsClosed(t *testing.T) {
	middleware, tokenService := newTestMiddleware(failingRevocationChecker{})

	tokenStr, _, err := tokenService.Generate("user-123")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	var reached bool
	var gotUserID string
	w := httptest.NewRecorder()

	middleware.Authenticate(spyHandler(&reached, &gotUserID))(w, requestWithToken(tokenStr))

	if reached {
		t.Error("expected a failed revocation lookup not to reach the handler")
	}
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestAuthMiddleware_RejectsMalformedAuthorizationHeaders(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{name: "missing header", header: ""},
		{name: "no bearer prefix", header: "some-token"},
		{name: "wrong scheme", header: "Basic some-token"},
		{name: "garbage token", header: "Bearer not-a-real-token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware, _ := newTestMiddleware(services.NewSessionService(&memory.InMemoryRevokedTokenRepository{}))

			r := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.header != "" {
				r.Header.Set("Authorization", tt.header)
			}

			var reached bool
			var gotUserID string
			w := httptest.NewRecorder()

			middleware.Authenticate(spyHandler(&reached, &gotUserID))(w, r)

			if reached {
				t.Error("expected the handler not to be reached")
			}
			if w.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestAuthMiddleware_TokenWithoutSubIsRejected(t *testing.T) {
	middleware, _ := newTestMiddleware(services.NewSessionService(&memory.InMemoryRevokedTokenRepository{}))

	// Correctly signed, but names no user.
	anonymous := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"jti": "token-abc",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := anonymous.SignedString([]byte(testKey))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	var reached bool
	var gotUserID string
	w := httptest.NewRecorder()

	middleware.Authenticate(spyHandler(&reached, &gotUserID))(w, requestWithToken(tokenStr))

	if reached {
		t.Error("expected a token without sub not to reach the handler")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
