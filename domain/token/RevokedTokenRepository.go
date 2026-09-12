package token

import (
	"context"
	"time"
)

type RevokedTokenRepository interface {
	Revoke(context.Context, *RevokedToken) error
	IsRevoked(context.Context, string) (bool, error)
	DeleteExpired(context.Context, time.Time) error
}
