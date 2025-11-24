package models

import (
	"time"

	"github.com/gofrs/uuid"
)

// UserPermission is used by pop to map your .model.Name.Proper.Pluralize.Underscore database table to your go code.
type UserPermission struct {
	ID           uuid.UUID `gorm:"primaryKey" json:"id"`
	UserID       uuid.UUID `json:"user_id" gorm:"column:user_id"`
	PermissionID uuid.UUID `json:"permission_id" gorm:"column:permission_id"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at"`
	User         User
	Permission   Permission
}

// UserPermissions is not required by pop and may be deleted
type UserPermissions []UserPermission
