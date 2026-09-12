package memory

import (
	"GarageSaleAPI/domain/token"
	"GarageSaleAPI/test"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func revoked(jti string, expiresAt time.Time) token.RevokedToken {
	return *token.CreateRevokedToken(jti, uuid.NewString(), expiresAt, time.Now())
}

func TestInMemoryRevokedTokenRepository_Revoke(t *testing.T) {
	type fields struct {
		revokedTokens map[string]token.RevokedToken
	}
	type args struct {
		token *token.RevokedToken
		ctx   context.Context
	}

	validToken := token.CreateRevokedToken(uuid.NewString(), uuid.NewString(), time.Now().Add(time.Hour), time.Now())

	tests := []struct {
		name       string
		fields     fields
		args       args
		wantLength int
		wantErr    bool
		textErr    string
	}{
		{
			name:       "revoke token",
			fields:     fields{revokedTokens: map[string]token.RevokedToken{}},
			args:       args{token: validToken, ctx: test.CreateTestContext(t)},
			wantLength: 1,
			wantErr:    false,
			textErr:    "",
		},
		{
			name:       "revoke already revoked token",
			fields:     fields{revokedTokens: map[string]token.RevokedToken{validToken.Jti(): *validToken}},
			args:       args{token: validToken, ctx: test.CreateTestContext(t)},
			wantLength: 1,
			wantErr:    false,
			textErr:    "",
		},
		{
			name:       "revoke token with timed out context",
			fields:     fields{revokedTokens: map[string]token.RevokedToken{}},
			args:       args{token: validToken, ctx: test.CreateTimedOutTestContext(t)},
			wantLength: 0,
			wantErr:    true,
			textErr:    context.DeadlineExceeded.Error(),
		},
		{
			name:       "revoke token with cancelled context",
			fields:     fields{revokedTokens: map[string]token.RevokedToken{}},
			args:       args{token: validToken, ctx: test.CreateCancelledTestContext()},
			wantLength: 0,
			wantErr:    true,
			textErr:    context.Canceled.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := InMemoryRevokedTokenRepository{
				revokedTokens: tt.fields.revokedTokens,
			}

			err := repo.Revoke(tt.args.ctx, tt.args.token)
			if err != nil && !tt.wantErr ||
				((err != nil) && err.Error() != tt.textErr) {
				t.Errorf("InMemoryRevokedTokenRepository.Revoke() error = %v, wantErr %v\ntext = %v, textErr = %v",
					err, tt.wantErr, err.Error(), tt.textErr)
			}

			if len(repo.revokedTokens) != tt.wantLength {
				t.Errorf("len(revokedTokens) = %d, want %d", len(repo.revokedTokens), tt.wantLength)
			}
		})
	}
}

func TestInMemoryRevokedTokenRepository_IsRevoked(t *testing.T) {
	type fields struct {
		revokedTokens map[string]token.RevokedToken
	}
	type args struct {
		jti string
		ctx context.Context
	}

	revokedToken := revoked(uuid.NewString(), time.Now().Add(time.Hour))

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
		textErr string
	}{
		{
			name:    "revoked token is reported revoked",
			fields:  fields{revokedTokens: map[string]token.RevokedToken{revokedToken.Jti(): revokedToken}},
			args:    args{jti: revokedToken.Jti(), ctx: test.CreateTestContext(t)},
			want:    true,
			wantErr: false,
			textErr: "",
		},
		{
			name:    "unknown token is not revoked",
			fields:  fields{revokedTokens: map[string]token.RevokedToken{revokedToken.Jti(): revokedToken}},
			args:    args{jti: uuid.NewString(), ctx: test.CreateTestContext(t)},
			want:    false,
			wantErr: false,
			textErr: "",
		},
		{
			name:    "empty denylist reports nothing revoked",
			fields:  fields{revokedTokens: map[string]token.RevokedToken{}},
			args:    args{jti: revokedToken.Jti(), ctx: test.CreateTestContext(t)},
			want:    false,
			wantErr: false,
			textErr: "",
		},
		{
			name:    "check with timed out context",
			fields:  fields{revokedTokens: map[string]token.RevokedToken{revokedToken.Jti(): revokedToken}},
			args:    args{jti: revokedToken.Jti(), ctx: test.CreateTimedOutTestContext(t)},
			want:    false,
			wantErr: true,
			textErr: context.DeadlineExceeded.Error(),
		},
		{
			name:    "check with cancelled context",
			fields:  fields{revokedTokens: map[string]token.RevokedToken{revokedToken.Jti(): revokedToken}},
			args:    args{jti: revokedToken.Jti(), ctx: test.CreateCancelledTestContext()},
			want:    false,
			wantErr: true,
			textErr: context.Canceled.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := InMemoryRevokedTokenRepository{
				revokedTokens: tt.fields.revokedTokens,
			}

			got, err := repo.IsRevoked(tt.args.ctx, tt.args.jti)
			if err != nil && !tt.wantErr ||
				((err != nil) && err.Error() != tt.textErr) {
				t.Errorf("InMemoryRevokedTokenRepository.IsRevoked() error = %v, wantErr %v\ntext = %v, textErr = %v",
					err, tt.wantErr, err.Error(), tt.textErr)
			}

			if got != tt.want {
				t.Errorf("IsRevoked() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInMemoryRevokedTokenRepository_DeleteExpired(t *testing.T) {
	expired := revoked(uuid.NewString(), time.Now().Add(-time.Hour))
	live := revoked(uuid.NewString(), time.Now().Add(time.Hour))

	repo := InMemoryRevokedTokenRepository{
		revokedTokens: map[string]token.RevokedToken{
			expired.Jti(): expired,
			live.Jti():    live,
		},
	}

	ctx := test.CreateTestContext(t)
	if err := repo.DeleteExpired(ctx, time.Now()); err != nil {
		t.Fatalf("DeleteExpired() error = %v", err)
	}

	if _, found := repo.revokedTokens[expired.Jti()]; found {
		t.Error("expected the expired token to be swept")
	}

	if _, found := repo.revokedTokens[live.Jti()]; !found {
		t.Error("expected the unexpired token to be kept")
	}
}

func TestInMemoryRevokedTokenRepository_DeleteExpired_CancelledContext(t *testing.T) {
	live := revoked(uuid.NewString(), time.Now().Add(time.Hour))

	repo := InMemoryRevokedTokenRepository{
		revokedTokens: map[string]token.RevokedToken{live.Jti(): live},
	}

	err := repo.DeleteExpired(test.CreateCancelledTestContext(), time.Now())
	if err == nil || err.Error() != context.Canceled.Error() {
		t.Errorf("DeleteExpired() error = %v, want %v", err, context.Canceled)
	}

	if len(repo.revokedTokens) != 1 {
		t.Errorf("len(revokedTokens) = %d, want 1", len(repo.revokedTokens))
	}
}

// The zero value has a nil map; Revoke has to cope, since that is how callers
// construct it.
func TestInMemoryRevokedTokenRepository_RevokeOnZeroValueRepository(t *testing.T) {
	repo := InMemoryRevokedTokenRepository{}
	ctx := test.CreateTestContext(t)

	revokedToken := token.CreateRevokedToken(uuid.NewString(), uuid.NewString(), time.Now().Add(time.Hour), time.Now())

	if err := repo.Revoke(ctx, revokedToken); err != nil {
		t.Fatalf("Revoke() error = %v", err)
	}

	got, err := repo.IsRevoked(ctx, revokedToken.Jti())
	if err != nil {
		t.Fatalf("IsRevoked() error = %v", err)
	}
	if !got {
		t.Error("IsRevoked() = false, want true")
	}
}
