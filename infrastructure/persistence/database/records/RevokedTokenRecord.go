package records

import "time"

type RevokedTokenRecord struct {
	Jti       string    `gorm:"column:jti;type:uuid;primaryKey"`
	UserId    string    `gorm:"column:user_id;not null"`
	ExpiresAt time.Time `gorm:"column:expires_at;index:idx_revoked_tokens_expires_at;not null"`
	RevokedAt time.Time `gorm:"column:revoked_at;not null;default:now()"`
}

func (RevokedTokenRecord) TableName() string { return "revoked_tokens" }
