package services

import (
	"GarageSaleAPI/application/server/apperror"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenGenerator interface {
	Generate(userID string) (string, time.Time, error)
}

type TokenVerifier interface {
	Verify(tokenStr string) (jwt.MapClaims, error)
}

type TokenService struct {
	key []byte
	ttl time.Duration
}

func NewTokenService(key []byte, ttl time.Duration) *TokenService {
	return &TokenService{key: key, ttl: ttl}
}

func (s *TokenService) Generate(userID string) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.ttl)
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": expiresAt.Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(s.key)
	return signed, expiresAt, err
}

func (s *TokenService) Verify(tokenStr string) (jwt.MapClaims, error) {
	parsed, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.key, nil
	})
	if err != nil || !parsed.Valid {
		return nil, apperror.Unauthorized("invalid token: ", err)
	}
	return parsed.Claims.(jwt.MapClaims), nil
}
