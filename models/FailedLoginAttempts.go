package models

import (
	"time"

	"github.com/gofrs/uuid"
)

// FailedLoginAttempt is used by pop to map your .model.Name.Proper.Pluralize.Underscore database table to your go code.
type FailedLoginAttempt struct {
	ID        uuid.UUID `json:"id" gorm:"column:id"`
	UserID    uuid.UUID `json:"user_id" gorm:"column:user_id"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
	User      User      `belongs_to:"user"`
}

func (FailedLoginAttempt) TableName() string {
	return "failed_login_attempts"
}

// FailedLoginAttempts is not required by pop and may be deleted
type FailedLoginAttempts []FailedLoginAttempt
