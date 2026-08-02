package interfaces

import (
	"GarageSaleAPI/application/services"
	"net/http"
	"strings"
)

type AuthenticatedHandler func(w http.ResponseWriter, r *http.Request, userID string)

type AuthMiddleware struct {
	tokenService services.TokenVerifier
}

func NewAuthenticationMiddleware(tokenService services.TokenVerifier) *AuthMiddleware {
	return &AuthMiddleware{tokenService: tokenService}
}

func (m *AuthMiddleware) Authenticate(next AuthenticatedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondUnauthorized(w, "missing authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondUnauthorized(w, "invalid authorization header format")
			return
		}
		tokenStr := parts[1]

		claims, err := m.tokenService.Verify(tokenStr)
		if err != nil {
			respondUnauthorized(w, "invalid or expired token")
			return
		}

		userID, ok := claims["sub"].(string)
		if !ok || userID == "" {
			respondUnauthorized(w, "invalid token claims")
			return
		}

		next(w, r, userID)
	}
}

func respondUnauthorized(w http.ResponseWriter, message string) {
	WriteResponse(w, message, http.StatusUnauthorized, "application/json")
	Encode(w, map[string]string{"error": message})
}
