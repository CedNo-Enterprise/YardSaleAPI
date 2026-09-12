package services

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/infrastructure/persistence/memory"
	"GarageSaleAPI/test"
	"testing"
	"time"

	"GarageSaleAPI/domain/token"
	"context"
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// issueClaims mints a real token and returns its verified claims, so the tests
// see the same claim types the middleware hands to Logout.
func issueClaims(t *testing.T, userID string) jwt.MapClaims {
	t.Helper()

	tokenService := NewTokenService([]byte("test-secret-key"), time.Hour)
	tokenStr, _, err := tokenService.Generate(userID)
	require.NoError(t, err)

	claims, err := tokenService.Verify(tokenStr)
	require.NoError(t, err)

	return claims
}

func TestSessionService_Logout_RevokesToken(t *testing.T) {
	service := NewSessionService(&memory.InMemoryRevokedTokenRepository{})
	ctx := test.CreateTestContext(t)

	claims := issueClaims(t, "user-123")
	jti := claims["jti"].(string)

	revoked, err := service.IsRevoked(ctx, jti)
	require.NoError(t, err)
	require.False(t, revoked, "token should not start out revoked")

	require.NoError(t, service.Logout(ctx, claims))

	revoked, err = service.IsRevoked(ctx, jti)
	require.NoError(t, err)
	assert.True(t, revoked, "token should be revoked after logout")
}

func TestSessionService_Logout_LeavesOtherSessionsAlone(t *testing.T) {
	service := NewSessionService(&memory.InMemoryRevokedTokenRepository{})
	ctx := test.CreateTestContext(t)

	firstSession := issueClaims(t, "user-123")
	secondSession := issueClaims(t, "user-123")

	require.NoError(t, service.Logout(ctx, firstSession))

	revoked, err := service.IsRevoked(ctx, secondSession["jti"].(string))
	require.NoError(t, err)
	assert.False(t, revoked, "the same user's other session should survive")
}

func TestSessionService_Logout_RejectsIncompleteClaims(t *testing.T) {
	validExp := float64(time.Now().Add(time.Hour).Unix())

	tests := []struct {
		name   string
		claims jwt.MapClaims
	}{
		{
			name:   "missing jti",
			claims: jwt.MapClaims{"sub": "user-123", "exp": validExp},
		},
		{
			name:   "empty jti",
			claims: jwt.MapClaims{"sub": "user-123", "jti": "", "exp": validExp},
		},
		{
			name:   "non-string jti",
			claims: jwt.MapClaims{"sub": "user-123", "jti": 42, "exp": validExp},
		},
		{
			name:   "missing sub",
			claims: jwt.MapClaims{"jti": "token-abc", "exp": validExp},
		},
		{
			name:   "missing exp",
			claims: jwt.MapClaims{"sub": "user-123", "jti": "token-abc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewSessionService(&memory.InMemoryRevokedTokenRepository{})

			err := service.Logout(test.CreateTestContext(t), tt.claims)

			require.Error(t, err)
			test.AssertKind(t, err, apperror.KindUnauthorized)
		})
	}
}

func TestSessionService_Logout_IsIdempotent(t *testing.T) {
	service := NewSessionService(&memory.InMemoryRevokedTokenRepository{})
	ctx := test.CreateTestContext(t)

	claims := issueClaims(t, "user-123")

	require.NoError(t, service.Logout(ctx, claims))
	require.NoError(t, service.Logout(ctx, claims))

	revoked, err := service.IsRevoked(ctx, claims["jti"].(string))
	require.NoError(t, err)
	assert.True(t, revoked)
}

// failingRevokedTokenRepository stands in for a database that rejects writes.
type failingRevokedTokenRepository struct{}

func (failingRevokedTokenRepository) Revoke(context.Context, *token.RevokedToken) error {
	return apperror.Internal(errors.New("database unavailable"))
}

func (failingRevokedTokenRepository) IsRevoked(context.Context, string) (bool, error) {
	return false, apperror.Internal(errors.New("database unavailable"))
}

func (failingRevokedTokenRepository) DeleteExpired(context.Context, time.Time) error {
	return apperror.Internal(errors.New("database unavailable"))
}

// A failed write must surface. Reporting success here would tell the caller the
// token is dead while it stays usable for the rest of its lifetime.
func TestSessionService_Logout_PropagatesRepositoryFailure(t *testing.T) {
	service := NewSessionService(failingRevokedTokenRepository{})

	err := service.Logout(test.CreateTestContext(t), issueClaims(t, "user-123"))

	require.Error(t, err)
	test.AssertKind(t, err, apperror.KindInternal)
}
