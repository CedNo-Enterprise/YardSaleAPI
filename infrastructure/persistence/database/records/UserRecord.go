package records

import "time"

type UserRecord struct {
	Id        string    `gorm:"column:id;type:uuid;primaryKey"`
	Username  string    `gorm:"column:username;not null"`
	Password  string    `gorm:"column:password;not null"`
	Email     string    `gorm:"column:email;uniqueIndex:idx_email;not null"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:now()"`
}

func (UserRecord) TableName() string { return "users" }
