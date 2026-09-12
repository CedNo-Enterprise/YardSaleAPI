package memory

import (
	"GarageSaleAPI/domain/token"
	"context"
	"time"
)

type InMemoryRevokedTokenRepository struct {
	revokedTokens map[string]token.RevokedToken
}

func (repo *InMemoryRevokedTokenRepository) Revoke(ctx context.Context, t *token.RevokedToken) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if repo.revokedTokens == nil {
		repo.revokedTokens = make(map[string]token.RevokedToken)
	}

	repo.revokedTokens[t.Jti()] = *t
	return nil
}

func (repo *InMemoryRevokedTokenRepository) IsRevoked(ctx context.Context, jti string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	_, found := repo.revokedTokens[jti]
	return found, nil
}

func (repo *InMemoryRevokedTokenRepository) DeleteExpired(ctx context.Context, cutoff time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	for jti, revoked := range repo.revokedTokens {
		if revoked.ExpiresAt().Before(cutoff) {
			delete(repo.revokedTokens, jti)
		}
	}
	return nil
}
