package token

import (
	"time"
)

type RevokedToken struct {
	jti       string
	userId    string
	expiresAt time.Time
	revokedAt time.Time
}

func (t RevokedToken) Jti() string {
	return t.jti
}

func (t RevokedToken) UserId() string {
	return t.userId
}

func (t RevokedToken) ExpiresAt() time.Time {
	return t.expiresAt
}

func (t RevokedToken) RevokedAt() time.Time {
	return t.revokedAt
}
