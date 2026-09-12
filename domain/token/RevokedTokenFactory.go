package token

import (
	"time"
)

func CreateRevokedToken(jti string, userId string, expiresAt time.Time, revokedAt time.Time) *RevokedToken {
	return &RevokedToken{
		jti:       jti,
		userId:    userId,
		expiresAt: expiresAt,
		revokedAt: revokedAt,
	}
}
