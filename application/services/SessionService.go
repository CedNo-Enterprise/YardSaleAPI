package services

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/token"
	"context"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type RevocationChecker interface {
	IsRevoked(ctx context.Context, jti string) (bool, error)
}

type SessionService struct {
	revokedTokens token.RevokedTokenRepository
}

func NewSessionService(revokedTokens token.RevokedTokenRepository) *SessionService {
	return &SessionService{revokedTokens: revokedTokens}
}

// Logout revokes the token the claims came from. The claims are expected to
// have been verified already by the authentication middleware.
func (service *SessionService) Logout(ctx context.Context, claims jwt.MapClaims) error {
	jti, ok := claims["jti"].(string)
	if !ok || jti == "" {
		return apperror.Unauthorized("invalid token claims", nil)
	}

	userId, ok := claims["sub"].(string)
	if !ok || userId == "" {
		return apperror.Unauthorized("invalid token claims", nil)
	}

	exp, err := claims.GetExpirationTime()
	if err != nil || exp == nil {
		return apperror.Unauthorized("invalid token claims", err)
	}

	revoked := token.CreateRevokedToken(jti, userId, exp.Time, time.Now())
	if err := service.revokedTokens.Revoke(ctx, revoked); err != nil {
		slog.Error("error revoking token", "jti", jti, "err", err.Error())
		return err
	}

	return nil
}

func (service *SessionService) IsRevoked(ctx context.Context, jti string) (bool, error) {
	return service.revokedTokens.IsRevoked(ctx, jti)
}
