package services

import (
	"crypto/rand"
	"crypto/rsa"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenService_GenerateAndVerify(t *testing.T) {
	svc := NewTokenService([]byte("test-secret-key"), 15*time.Minute)

	tokenStr, expiresAt, err := svc.Generate("user-123")
	require.NoError(t, err)
	require.NotEmpty(t, tokenStr)

	assert.WithinDuration(t, time.Now().Add(15*time.Minute), expiresAt, 2*time.Second)

	claims, err := svc.Verify(tokenStr)
	require.NoError(t, err)

	assert.Equal(t, "user-123", claims["sub"])

	expClaim, ok := claims["exp"].(float64)
	require.True(t, ok)
	assert.InDelta(t, float64(expiresAt.Unix()), expClaim, 2)
}

func TestTokenService_Verify_WrongKey(t *testing.T) {
	svc := NewTokenService([]byte("correct-key"), 15*time.Minute)
	tokenStr, _, err := svc.Generate("user-123")
	require.NoError(t, err)

	otherSvc := NewTokenService([]byte("wrong-key"), 15*time.Minute)
	_, err = otherSvc.Verify(tokenStr)
	assert.Error(t, err)
}

func TestTokenService_Verify_ExpiredToken(t *testing.T) {
	svc := NewTokenService([]byte("test-secret-key"), -1*time.Minute)
	tokenStr, _, err := svc.Generate("user-123")
	require.NoError(t, err)

	_, err = svc.Verify(tokenStr)
	assert.Error(t, err)
}

func TestTokenService_Verify_TamperedToken(t *testing.T) {
	svc := NewTokenService([]byte("test-secret-key"), 15*time.Minute)
	tokenStr, _, err := svc.Generate("user-123")
	require.NoError(t, err)

	tampered := tokenStr[:len(tokenStr)-1] + "x"

	_, err = svc.Verify(tampered)
	assert.Error(t, err)
}

func TestTokenService_Verify_MalformedToken(t *testing.T) {
	svc := NewTokenService([]byte("test-secret-key"), 15*time.Minute)

	_, err := svc.Verify("not-a-real-token")
	assert.Error(t, err)
}

func TestTokenService_Verify_EmptyToken(t *testing.T) {
	svc := NewTokenService([]byte("test-secret-key"), 15*time.Minute)

	_, err := svc.Verify("")
	assert.Error(t, err)
}

func TestTokenService_Verify_RejectsAlgNone(t *testing.T) {
	svc := NewTokenService([]byte("test-secret-key"), 15*time.Minute)

	claims := jwt.MapClaims{
		"sub": "attacker",
		"exp": time.Now().Add(15 * time.Minute).Unix(),
	}
	unsignedTok := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenStr, err := unsignedTok.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	_, err = svc.Verify(tokenStr)
	assert.Error(t, err)
}

func TestTokenService_Verify_RejectsNonHMACAlgorithm(t *testing.T) {
	svc := NewTokenService([]byte("test-secret-key"), 15*time.Minute)

	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(15 * time.Minute).Unix(),
	})

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	signed, err := tok.SignedString(key)
	require.NoError(t, err)

	_, err = svc.Verify(signed)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected signing method")
}

func TestTokenService_Generate_DifferentUsersProduceDifferentTokens(t *testing.T) {
	svc := NewTokenService([]byte("test-secret-key"), 15*time.Minute)

	tok1, _, err := svc.Generate("user-1")
	require.NoError(t, err)
	tok2, _, err := svc.Generate("user-2")
	require.NoError(t, err)

	assert.NotEqual(t, tok1, tok2)
}

func TestTokenService_Generate_ProducesThreePartJWT(t *testing.T) {
	svc := NewTokenService([]byte("test-secret-key"), 15*time.Minute)

	tokenStr, _, err := svc.Generate("user-123")
	require.NoError(t, err)

	parts := strings.Split(tokenStr, ".")
	assert.Len(t, parts, 3, "JWT should have header.payload.signature")
}
