package interfaces

import (
	"GarageSaleAPI/application/services"
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type AuthenticatedHandler func(w http.ResponseWriter, r *http.Request, userID string)

type claimsContextKey struct{}

// ClaimsFrom returns the verified token claims the authentication middleware
// attached to the request.
func ClaimsFrom(ctx context.Context) (jwt.MapClaims, bool) {
	claims, ok := ctx.Value(claimsContextKey{}).(jwt.MapClaims)
	return claims, ok
}

type AuthMiddleware struct {
	tokenService      services.TokenVerifier
	revocationChecker services.RevocationChecker
}

func NewAuthenticationMiddleware(tokenService services.TokenVerifier, revocationChecker services.RevocationChecker) *AuthMiddleware {
	return &AuthMiddleware{tokenService: tokenService, revocationChecker: revocationChecker}
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

		jti, ok := claims["jti"].(string)
		if !ok || jti == "" {
			respondUnauthorized(w, "invalid token claims")
			return
		}

		revoked, err := m.revocationChecker.IsRevoked(r.Context(), jti)
		if err != nil {
			// Fail closed: a lookup failure must not be read as "not revoked".
			slog.Error("error checking token revocation", "jti", jti, "err", err.Error())
			respondInternalError(w)
			return
		}
		if revoked {
			respondUnauthorized(w, "invalid or expired token")
			return
		}

		next(w, r.WithContext(context.WithValue(r.Context(), claimsContextKey{}, claims)), userID)
	}
}

func respondUnauthorized(w http.ResponseWriter, message string) {
	WriteResponse(w, map[string]string{"error": message}, http.StatusUnauthorized, "application/json")
}

func respondInternalError(w http.ResponseWriter) {
	WriteResponse(w, map[string]string{"error": "internal server error"}, http.StatusInternalServerError, "application/json")
}
